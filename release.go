// This work is subject to the CC0 1.0 Universal (CC0 1.0) Public Domain Dedication
// license. Its contents can be found at:
// http://creativecommons.org/publicdomain/zero/1.0/

package xbindata

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

// writeRelease writes the release code file.
func writeRelease(w io.Writer, c *Config, toc []Asset) error {
	err := writeReleaseHeader(w, c, toc)
	if err != nil {
		return err
	}

	if !c.Outlined {
		var start int64
		for i := range toc {
			err = writeReleaseAsset(start, w, c, &toc[i])
			if err != nil {
				return err
			}
			start += toc[i].Size
		}
	}

	return nil
}

// writeReleaseHeader writes output file headers.
// This targets release builds.
func writeReleaseHeader(w io.Writer, c *Config, toc []Asset) error {
	var err error
	if c.Outlined {
		err = header_outlined(w, c, toc)
	} else if c.NoCompress {
		if c.NoMemCopy {
			err = header_uncompressed_nomemcopy(w, c)
		} else {
			err = header_uncompressed_memcopy(w, c)
		}
	} else {
		if c.NoMemCopy {
			err = header_compressed_nomemcopy(w, c)
		} else {
			err = header_compressed_memcopy(w, c)
		}
	}
	if err != nil {
		return err
	}
	return header_release_common(w, c)
}

// writeReleaseAsset write a release entry for the given asset.
// A release entry is a function which embeds and returns
// the file's byte content.
func writeReleaseAsset(start int64, w io.Writer, c *Config, asset *Asset) error {
	fd, err := os.Open(asset.Path)
	if err != nil {
		return err
	}

	defer fd.Close()

	h := sha256.New()
	tr := io.TeeReader(fd, h)
	if !c.Outlined {
		if c.NoCompress {
			if c.NoMemCopy {
				err = uncompressed_nomemcopy(w, asset, tr)
			} else {
				err = uncompressed_memcopy(w, asset, tr)
			}
		} else {
			if c.NoMemCopy {
				err = compressed_nomemcopy(w, asset, tr)
			} else {
				err = compressed_memcopy(w, asset, tr)
			}
		}
		if err != nil {
			return err
		}
	}
	var digest [sha256.Size]byte
	copy(digest[:], h.Sum(nil))
	return asset_release_common(start, w, c, asset, digest)
}

var (
	backquote = []byte("`")
	bom       = []byte("\xEF\xBB\xBF")
)

// sanitize prepares a valid UTF-8 string as a raw string constant.
// Based on https://code.google.com/p/go/source/browse/godoc/static/makestatic.go?repo=tools
func sanitize(b []byte) []byte {
	var chunks [][]byte
	for i, b := range bytes.Split(b, backquote) {
		if i > 0 {
			chunks = append(chunks, backquote)
		}
		for j, c := range bytes.Split(b, bom) {
			if j > 0 {
				chunks = append(chunks, bom)
			}
			if len(c) > 0 {
				chunks = append(chunks, c)
			}
		}
	}

	var buf bytes.Buffer
	sanitizeChunks(&buf, chunks)
	return buf.Bytes()
}

func sanitizeChunks(buf *bytes.Buffer, chunks [][]byte) {
	n := len(chunks)
	if n >= 2 {
		buf.WriteString("(")
		sanitizeChunks(buf, chunks[:n/2])
		buf.WriteString(" + ")
		sanitizeChunks(buf, chunks[n/2:])
		buf.WriteString(")")
		return
	}
	b := chunks[0]
	if bytes.Equal(b, backquote) {
		buf.WriteString("\"`\"")
		return
	}
	if bytes.Equal(b, bom) {
		buf.WriteString(`"\xEF\xBB\xBF"`)
		return
	}
	buf.WriteString("`")
	buf.Write(b)
	buf.WriteString("`")
}

func write_imports(w io.Writer, c *Config, imports ...string) (err error) {
	imports = append(
		imports,
		`bc "github.com/moisespsena-go/xbindata/xbcommon"`,
		`github.com/moisespsena-go/io-common`,
	)

	if c.FileSystem {
		imports = append(
			imports,
			`fs "github.com/moisespsena-go/xbindata/xbfs"`,
		)
	}

	sort.Slice(imports, func(i, j int) bool {
		a, b := imports[i], imports[j]
		ia, ib := strings.IndexByte(a, ' '), strings.IndexByte(b, ' ')
		if ia == ib {
			if len(a) < len(b) {
				return true
			}
			if len(a) == len(b) {
				return a < b
			}
			return false
		} else if ia > 0 && ib < 0 {
			return false
		} else if ib > 0 && ia < 0 {
			return true
		}
		return a[0:ia] < b[0:ib]
	})

	for i, imp := range imports {
		if imp[len(imp)-1] != '"' {
			imp = `"` + imp + `"`
		}
		imports[i] = "\t" + imp
	}
	_, err = fmt.Fprintf(w, "import (\n"+strings.Join(imports, "\n")+"\n)\n")
	return
}

func header_compressed_nomemcopy(w io.Writer, c *Config) (err error) {
	if err = write_imports(w, c,
		"compress/gzip",
		"crypto/sha256",
		"fmt",
		"os",
		"strings",
		"time",
	); err != nil {
		return
	}
	_, err = fmt.Fprintf(w, `
func bindataReader(data, name string) (iocommon.ReadSeekCloser, error) {
	gz, err := gzip.NewReader(strings.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("read %%q: %%v", name, err)
	}
	return iocommon.NoSeeker(gz), nil
}

`)
	return
}

func header_compressed_memcopy(w io.Writer, c *Config) (err error) {
	if err = write_imports(w, c,
		"bytes",
		"compress/gzip",
		"crypto/sha256",
		"fmt",
		"os",
		"time",
	); err != nil {
		return
	}
	_, err = fmt.Fprintf(w, `
func bindataReader(data []byte, name string) (iocommon.ReadSeekCloser, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("read %%q: %%v", name, err)
	}
	return iocommon.NoSeeker(gz), nil
}

`)
	return
}

func header_uncompressed_nomemcopy(w io.Writer, c *Config) (err error) {
	if err = write_imports(w, c,
		"crypto/sha256",
		"os",
		"reflect",
		"time",
		"unsafe",
	); err != nil {
		return
	}
	_, err = fmt.Fprintf(w, `
func bindataReader(data *string, name string) (iocommon.ReadSeekCloser, error) {
	var empty [0]byte
	sx := (*reflect.StringHeader)(unsafe.Pointer(data))
	b := empty[:]
	bx := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	bx.Data = sx.Data
	bx.Len = len(*data)
	bx.Cap = bx.Len

	return iocommon.NewBytesReadCloser(b), nil
}

`)
	return err
}

func header_uncompressed_memcopy(w io.Writer, c *Config) (err error) {
	if err = write_imports(w, c,
		"crypto/sha256",
		"fmt",
		"io/ioutil",
		"os",
		"path/filepath",
		"strings",
		"time",
	); err != nil {
		return
	}
	return
}

func header_outlined(w io.Writer, c *Config, toc []Asset) (err error) {
	var imports = []string{
		"os",
		`br "github.com/moisespsena-go/xbindata/xbreader"`,
		`"github.com/moisespsena-go/xbindata/outlined"`,
		"github.com/moisespsena-go/path-helpers",
		`fsapi "github.com/moisespsena-go/assetfs/assetfsapi"`,
		"sync",
		"regexp",
		"strings",
		"path/filepath",
		"errors",
	}

	if c.Hybrid {
		imports = append(imports, "github.com/moisespsena-go/assetfs")
	}

	if err = write_imports(w, c, imports...); err != nil {
		return
	}
	var size int64
	for i := range toc {
		size += toc[i].Size
	}
	var outlineds []string

	if c.Output != "" {
		if !c.NoCompress {
			outlineds = append(outlineds, strconv.Quote(c.Output+".gz"))
		}
		outlineds = append(outlineds, strconv.Quote(c.Output))
	}
	if c.OutlineEmbeded {
		outlineds = append(outlineds, "os.Args[0]")
	}
	outlined := strings.Join(outlineds, ", ")

	preInit := c.EmbedPreInitSource
	if preInit != "" {
		preInit = "\n" + preInit
	}

	data := `
var (
	_outlined     *outlined.Outlined
	outlinedPath  string
	outlinedPaths []string
    mu           sync.Mutex

	StartPos int64
	Assets   bc.Assets

	pkg           = path_helpers.GetCalledDir()
	envName       = "XBINDATA_ARCHIVE__" + strings.ToUpper(regexp.MustCompile("[\\W]+").ReplaceAllString(pkg, "_"))
	OpenOutlined   = br.Open
	OutlinedReaderFactory = func(start, size int64) func() (reader iocommon.ReadSeekCloser, err error) {
		return func() (reader iocommon.ReadSeekCloser, err error) {
			return OpenOutlined(outlinedPath, start, size)
		}
	}
)

func EnvName() string {
	return envName
}

func OutlinedPath() string {
	return outlinedPath
}

func Outlined() (archiv *outlined.Outlined, err error) {
	if _outlined == nil {
        mu.Lock()
		defer mu.Unlock() 

		if _outlined, err = outlined.OpenFile(outlinedPath); err != nil {
			return
		}
	}
	return _outlined, nil
}

func LoadOutlined() {` + preInit + `
	if outlinedPath == "" {
		for _, pth := range append(strings.Split(os.Getenv(envName), string(filepath.ListSeparator)), ` + outlined + `) {
    	    if pth == "" { continue }

			if _, err := os.Stat(pth); err == nil {
				outlinedPath = pth
				break;
			} else if !os.IsNotExist(err) {
				panic(err)
			}
		}
	}

	if outlinedPath == "" {
		panic(errors.New("outlined path not defined"))
	}

    Assets.Factory = func() (assets map[string]bc.Asset, err error) {`
	data += `
		archiv, err := Outlined()
		if err != nil {
			return nil, err
		}
    
		return archiv.AssetsMap(OutlinedReaderFactory), nil`
	data += `
	}
}

`

	if c.Hybrid {
		data += "func IsLocal() bool {\n"
		data += fmt.Sprintf("\tif _, err := os.Stat(%q); err == nil {\n\t\t", filepath.Join(c.Package, "data"))
		data += "return true\n\t\t"
		data += "}\n\t"
		data += "return false\n}\n\n"
	}

	data += "func Load() {\n"
	if c.Hybrid {
		data += `	if IsLocal() {
		LoadLocal()
	} else {
		LoadOutlined()
	}
`
	} else {
		data += "LoadOutlined()\n"
	}

	data += "}\n"

	if !c.NoAutoLoad {
		data += "\nfunc init() { Load() }\n"
	}

	data += "\nfunc FS() fsapi.Interface {\n"
	if c.Hybrid {
		data += `	if IsLocal() {
		return LocalFS
	}
`
		data += "\treturn OutlinedFS"
	}
	data += "\n}\n"

	_, err = fmt.Fprintf(w, data)
	return
}

func header_release_common(w io.Writer, c *Config) (err error) {
	if c.FileSystem {
		if c.Hybrid {
			_, err = fmt.Fprintf(w, `
var OutlinedFS = fs.NewFileSystem(&Assets)
`)
		} else {
			_, err = fmt.Fprintf(w, `
var AssetFS fsapi.Interface = fs.NewFileSystem(&Assets)
`)
		}
	}
	return
}

func compressed_nomemcopy(w io.Writer, asset *Asset, r io.Reader) error {
	_, err := fmt.Fprintf(w, `var _%s = "`, asset.Func)
	if err != nil {
		return err
	}

	gz := gzip.NewWriter(&StringWriter{Writer: w})
	_, err = io.Copy(gz, r)
	gz.Close()

	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(w, `"

func %sReader()(iocommon.ReadSeekCloser, error) {
	return bindataReader(_%s, %q)
}

`, asset.Func, asset.Func, asset.Name)
	return err
}

func compressed_memcopy(w io.Writer, asset *Asset, r io.Reader) error {
	_, err := fmt.Fprintf(w, `var _%s = []byte("`, asset.Func)
	if err != nil {
		return err
	}

	gz := gzip.NewWriter(&StringWriter{Writer: w})
	_, err = io.Copy(gz, r)
	gz.Close()

	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(w, `")

func %sReader()(iocommon.ReadSeekCloser, error) {
	return bindataReader(_%s, %q)
}

`, asset.Func, asset.Func, asset.Name)
	return err
}

func uncompressed_nomemcopy(w io.Writer, asset *Asset, r io.Reader) error {
	_, err := fmt.Fprintf(w, `var _%s = "`, asset.Func)
	if err != nil {
		return err
	}

	_, err = io.Copy(&StringWriter{Writer: w}, r)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(w, `"

func %sReader()(iocommon.ReadSeekCloser, error) {
	return bindataReader(&_%s, %q)
}

`, asset.Func, asset.Func, asset.Name)
	return err
}

func uncompressed_memcopy(w io.Writer, asset *Asset, r io.Reader) error {
	_, err := fmt.Fprintf(w, `var _%s = []byte(`, asset.Func)
	if err != nil {
		return err
	}

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	if len(b) > 0 && utf8.Valid(b) && !bytes.Contains(b, []byte{0}) {
		w.Write(sanitize(b))
	} else {
		fmt.Fprintf(w, "%+q", b)
	}

	_, err = fmt.Fprintf(w, `)
func %sReader()(iocommon.ReadSeekCloser, error) {
	return iocommon.NewBytesReadCloser(_%s), nil
}

`, asset.Func, asset.Func)
	return err
}

func asset_release_common(start int64, w io.Writer, c *Config, asset *Asset, digest [sha256.Size]byte) error {
	var readerFunc string
	if c.Outlined {
		readerFunc = fmt.Sprintf("newOpener(%d, %d)", start, asset.Size)
	} else {
		readerFunc = fmt.Sprintf("%sReader", asset.Func)
	}
	info, err := asset.InfoExport(c)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, `var %s = bc.NewFile(bc.NewFileInfo(%q, %s), %s,  func()[sha256.Size]byte{return %#v})
`, asset.Func, asset.Name, info, readerFunc, digest)
	return err
}
