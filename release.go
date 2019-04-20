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
	"sort"
	"strings"
	"unicode/utf8"
)

type fsLoadCallbacksSlice []struct {
	pkg, pkgName string
	fun          string
}

// writeRelease writes the release code file.
func writeRelease(w io.Writer, c *Config, toc []Asset) error {
	err := writeReleaseHeader(w, c, toc)
	if err != nil {
		return err
	}

	if !c.Outlined && !c.NoStore {
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
	var (
		err             error
		imports         []string
		fsLoadCallbacks fsLoadCallbacksSlice
	)

	if c.FileSystem && len(c.FileSystemLoadCallbacks) > 0 {
		sort.Strings(c.FileSystemLoadCallbacks)

		for i, cb := range c.FileSystemLoadCallbacks {
			if pos := strings.LastIndexByte(cb, '.'); pos < 1 {
				return fmt.Errorf("invalid filesystem load callback %q", cb)
			} else {
				var (
					name = fmt.Sprintf("cb%02d", i)
					pkg  = cb[0:pos]
				)
				fsLoadCallbacks = append(fsLoadCallbacks, struct {
					pkg, pkgName string
					fun          string
				}{pkg, name, cb[pos+1:]})

				imports = append(imports, name+` "`+pkg+`"`)
			}
		}
	}

	if c.Outlined {
		err = header_outlined(w, c, toc, imports...)
	} else {
		if c.NoCompress {
			if c.NoMemCopy {
				err = header_uncompressed_nomemcopy(w, c, imports...)
			} else {
				err = header_uncompressed_memcopy(w, c, imports...)
			}
		} else {
			if c.NoMemCopy {
				err = header_compressed_nomemcopy(w, c, imports...)
			} else {
				err = header_compressed_memcopy(w, c, imports...)
			}
		}
	}

	if err == nil {
		data := `var (
    loaded        bool
    mu            sync.Mutex
`
		if c.FileSystem {
			data += `	fs fsapi.Interface
`
		}
		data += `
)
`
		_, err = w.Write([]byte(data))
	}

	if err != nil {
		return err
	}
	return header_release_common(w, c, fsLoadCallbacks)
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

func header_compressed_nomemcopy(w io.Writer, c *Config, imports ...string) (err error) {
	imports = append(imports,
		"compress/gzip",
		"fmt",
		"os",
		"strings",
		"time",
		"sync",
	)

	if err = write_imports(w, c, imports...); err != nil {
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

func header_compressed_memcopy(w io.Writer, c *Config, imports ...string) (err error) {
	imports = append(imports,
		"bytes",
		"compress/gzip",
		"fmt",
		"os",
		"time",
		"sync",
	)

	if err = write_imports(w, c, imports...); err != nil {
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

func header_uncompressed_nomemcopy(w io.Writer, c *Config, imports ...string) (err error) {
	imports = append(imports,
		"os",
		"reflect",
		"time",
		"unsafe",
		"sync",
	)

	if err = write_imports(w, c, imports...); err != nil {
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

func header_uncompressed_memcopy(w io.Writer, c *Config, imports ...string) (err error) {
	imports = append(imports,
		"os",
		"time",
		"sync",
	)

	return write_imports(w, c, imports...)
}

func header_outlined(w io.Writer, c *Config, toc []Asset, imports ...string) (err error) {
	imports = append(imports,
		"os",
		`br "github.com/moisespsena-go/xbindata/xbreader"`,
		`"github.com/moisespsena-go/xbindata/outlined"`,
		"github.com/moisespsena-go/path-helpers",
		`fsapi "github.com/moisespsena-go/assetfs/assetfsapi"`,
		"sync",
		"strings",
	)

	if err = write_imports(w, c, imports...); err != nil {
		return
	}
	var size int64
	for i := range toc {
		size += toc[i].Size
	}
	var (
		outlineds []string
	)

	if c.OutlinedProgram {
		outlineds = append(outlineds, "os.Args[0]")
	} else if c.Output != "" {
		if !c.NoCompress {
			outlineds = append(outlineds, `path.Join("_assets", pkg + ".xb.gz")`)
		}
		outlineds = append(outlineds, `path.Join("_assets", pkg + ".xb")`)
	}

	outlined := "\n\t\t\t" + strings.Join(outlineds, ",\n\t\t\t") + ",\n\t\t"

	preInit := c.EmbedPreInitSource
	if preInit != "" {
		preInit = "\n" + preInit
	}

	data := `
var (
	pkg          = path_helpers.GetCalledDir()
    envName      = "XB_"+strings.NewReplacer("/", "_", ".", "", "-", "").Replace(strings.ToUpper(pkg))

	_outlined     *outlined.Outlined
	outlinedPaths []string
	outlinedPath  = os.Getenv(envName)
    ended         = os.Getenv(envName+"_ENDED") == "true"

	StartPos int64
	Assets   bc.Assets

	OpenOutlined = br.Open
	OutlinedReaderFactory = func(start, size int64) func() (reader iocommon.ReadSeekCloser, err error) {
		return func() (reader iocommon.ReadSeekCloser, err error) {
			return OpenOutlined(outlinedPath, _outlined.StartPos + start, size)
		}
	}
)

func OutlinedPath() string {
	return outlinedPath
}

func Outlined() (archiv *outlined.Outlined, err error) {
	if _outlined == nil {
        mu.Lock()
		defer mu.Unlock() 

		if _outlined, err = outlined.OpenFile(outlinedPath, ended); err != nil {
			return
		}
	}
	return _outlined, nil
}

func load() {` + preInit + `
	if outlinedPath == "" {
`
	if c.OutlinedProgram {
		data += `       outlinedPath, ended = os.Args[0], true
`
	} else {
		data += `		pths := []string{` + outlined + `}

		for _, pth := range pths {
    	    if pth == "" { continue }
			if _, err := os.Stat(pth); err == nil {
				outlinedPath = pth
				break;
			} else if !os.IsNotExist(err) {
				panic(err)
			}
		}

		if outlinedPath == "" {
			panic(errors.New("outlined path not defined"))
		}
`
	}
	data += `	}

    Assets.Factory = func() (assets map[string]bc.Asset, err error) {`
	data += `		archiv, err := Outlined()
		if err != nil {
			return nil, err
		}
    
		return archiv.AssetsMap(OutlinedReaderFactory), nil`
	data += `
	}

    fs = xbfs.NewFileSystem(&Assets)
	println("xbindata outlined: PATH='"+outlinedPath+"' ENDED=", ended)
}

`

	_, err = fmt.Fprintf(w, data)
	return
}

func header_release_common(w io.Writer, c *Config, fsLoadCallbacks fsLoadCallbacksSlice) (err error) {
	var data string
	if c.FileSystem {
		data += `var fsLoadCallbacks []func(fs fsapi.Interface)

func OnFsLoad(cb ...func(fs fsapi.Interface)) {
	fsLoadCallbacks = append(fsLoadCallbacks, cb...)
}

func callFsLoadCallbacks() {
`
		if len(fsLoadCallbacks) > 0 {
			for _, cb := range fsLoadCallbacks {
				data += "\t" + cb.pkgName + "." + cb.fun + "(fs)\n"
			}
			data += ""
		}
		data += `	for _, f := range fsLoadCallbacks {
		f(fs)
	}
}
`
		data += `
func FS() fsapi.Interface {
	Load()
	return fs
}
`
	}

	data += `func Load() {
    if loaded { return }
	mu.Lock()
	if loaded { mu.Unlock(); return }
`
	if c.FileSystem {
		data += `	defer callFsLoadCallbacks()
`
	}

	data += `	defer mu.Unlock()
	defer func() { loaded = true }()
`
	data += `	load()
`
	data += `}
`

	if !c.NoAutoLoad {
		data += "\nfunc init() { Load() }\n"
	}

	if data != "" {
		_, err = fmt.Fprintf(w, data)
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
	_, err = fmt.Fprintf(w, `var %s = bc.NewFile(bc.NewFileInfo(%q, %s), %s,  &%#v)
`, asset.Func, asset.Name, info, readerFunc, digest)
	return err
}
