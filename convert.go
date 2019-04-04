// This work is subject to the CC0 1.0 Universal (CC0 1.0) Public Domain Dedication
// license. Its contents can be found at:
// http://creativecommons.org/publicdomain/zero/1.0/

package xbindata

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"syscall"
	"unicode"

	"github.com/gobwas/glob"

	"github.com/moisespsena-go/bits2str"

	"github.com/moisespsena-go/path-helpers"
)

var (
	pkg  = path_helpers.GetCalledDir()
	pkgu = strings.NewReplacer("/", "", "\\", "").Replace(pkg)
)

// Translate reads assets from an input directory, converts them
// to Go code and writes new files to the output specified
// in the given configuration.
func Translate(c *Config) error {
	var (
		toc          []Asset
		hibridOutput string
	)

	if c.Hibrid {
		c.FileSystem = true
		c.Tags = append(c.Tags, "!"+c.HibridTagName)
		hibridOutput = strings.TrimSuffix(c.Output, ".go") + "_default.go"
		c.Output = strings.TrimSuffix(c.Output, ".go") + "_compiled.go"
	}

	// Ensure our configuration has sane values.
	if err := c.validate(); err != nil {
		return err
	}

	var (
		knownFuncs   = make(map[string]int)
		visitedPaths = make(map[string]bool)
		finder       = &Finder{false, &toc, c.Ignore, c.IgnoreGlob, knownFuncs, visitedPaths}
	)

	// Locate all the assets.
	for _, input := range c.Input {
		finder.recursive = input.Recursive
		if err := finder.find(input.Path, c.Prefix); err != nil {
			return err
		}
	}

	// Create output file.
	buf := new(bytes.Buffer)
	// Write the header. This makes e.g. Github ignore diffs in generated files.
	if _, err := fmt.Fprint(buf, "// Code generated by xbindata. DO NOT EDIT.\n"); err != nil {
		return err
	}
	if _, err := fmt.Fprint(buf, "// sources:\n"); err != nil {
		return err
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	if !c.Embed {
		for _, asset := range toc {
			relative, err := filepath.Rel(wd, asset.Path)
			if err != nil {
				return err
			}
			if _, err = fmt.Fprintf(buf, "// %s (%s)\n", filepath.ToSlash(relative), bits2str.Bits(asset.Size)*bits2str.Byte); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprint(buf, "\n"); err != nil {
			return err
		}
	}

	// Write build tags, if applicable.
	if len(c.Tags) > 0 {
		if _, err := fmt.Fprintf(buf, "// +build %s\n\n", strings.Join(c.Tags, ",")); err != nil {
			return err
		}
	}

	// Write package declaration.
	_, err = fmt.Fprintf(buf, "package %s\n\n", c.Package)
	if err != nil {
		return err
	}

	// Write assets.
	if c.Debug || c.Dev {
		if os.Getenv("GO_BINDATA_TEST") == "true" {
			// If we don't do this, people running the tests on different
			// machines get different git diffs.
			for i := range toc {
				toc[i].Path = strings.Replace(toc[i].Path, wd, "/test", 1)
			}
		}
		err = writeDebug(buf, c, toc)
	} else {
		err = writeRelease(buf, c, toc)
	}
	if err != nil {
		return err
	}

	if !c.Embed {
		// Write table of contents
		if err := writeTOC(c, buf, toc); err != nil {
			return err
		}
	}

	if err = safefileWriteFile(c.Output, buf.Bytes(), 0666); err != nil {
		return err
	}

	if hibridOutput != "" {
		buf.Reset()
		buf.WriteString("// +build " + c.HibridTagName + "\n\n")
		buf.WriteString("package " + c.Package + "\n")
		buf.WriteString(`
import (
	"github.com/moisespsena-go/assetfs"
)

var (
	AssetFS    = assetfs.NewAssetFileSystem()
)

func init() {
`)
		// Locate all the assets.
		for _, input := range c.Input {
			buf.WriteString(fmt.Sprintf("    if err := AssetFS.RegisterPath(%q); err != nil {panic(err)}\n", input.Path))
		}

		buf.WriteString("}\n")

		if err = safefileWriteFile(hibridOutput, buf.Bytes(), 0666); err != nil {
			return err
		}
	}

	if c.Embed {
		buf.Reset()

		if err = archiveHeadersWrite(buf, toc, c); err == nil && c.ArchiveHeadersOutput != "" {
			fmt.Printf("user headers file: `%v`\n", c.ArchiveHeadersOutput)

			if err = path_helpers.MkdirAll(filepath.Dir(c.ArchiveHeadersOutput)); err != nil {
				return err
			}

			if mode, err := path_helpers.ResolveMode(c.ArchiveHeadersOutput); err != nil {
				return err
			} else if err = safefileWriteFile(c.ArchiveHeadersOutput, buf.Bytes(), mode); err != nil {
				return err
			}
		} else if err != nil {
			return fmt.Errorf("write headers to buffer failed: %v", err.Error())
		}

		headerPath := filepath.Join(filepath.Dir(c.Output), ".archive_compile", "main.go")

		fmt.Printf("main headers file: `%v`\n", headerPath)

		if err = path_helpers.MkdirAll(filepath.Dir(headerPath)); err != nil {
			return err
		} else if mode, err := path_helpers.ResolveMode(headerPath); err != nil {
			return err
		} else if err = safefileWriteFile(headerPath, buf.Bytes(), mode); err != nil {
			return err
		}

		if err != nil {
			return err
		}

		var dest = c.EmbedArchive

		if dest == "" {
			dest = "assets.bin"
		}

		println("Destination file:", dest)

		if dest, err = filepath.Abs(dest); err != nil {
			return err
		}

		cmd := exec.Command("go", "run", "main.go", dest)
		cmd.Dir = filepath.Dir(headerPath)
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err = cmd.Run(); err == nil {
			fmt.Printf("remove main headers file `%v`\n", headerPath)
			if err2 := os.RemoveAll(filepath.Dir(headerPath)); err2 != nil {
				fmt.Printf("remove main headers file failed: %v\n", err2)
			}
		} else {
			err = fmt.Errorf("compile archive failed: %v", err.Error())
		}
	}

	return err
}

// Implement sort.Interface for []os.FileInfo based on Name()
type byName []os.FileInfo

func (v byName) Len() int           { return len(v) }
func (v byName) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }
func (v byName) Less(i, j int) bool { return v[i].Name() < v[j].Name() }

// findFiles recursively finds all the file paths in the given directory tree.
// They are added to the given map as keys. Values will be safe function names
// for each file, which will be used when generating the output code.
func findFiles(dir, prefix string, recursive bool, toc *[]Asset, ignore []*regexp.Regexp, ignoreGlob []glob.Glob, knownFuncs map[string]int, visitedPaths map[string]bool) error {
	dirpath := dir
	if len(prefix) > 0 {
		dirpath, _ = filepath.Abs(dirpath)
		prefix, _ = filepath.Abs(prefix)
		prefix = filepath.ToSlash(prefix)
	}

	fi, err := os.Stat(dirpath)
	if err != nil {
		return err
	}

	var list []os.FileInfo

	if !fi.IsDir() {
		dirpath = filepath.Dir(dirpath)
		list = []os.FileInfo{fi}
	} else {
		visitedPaths[dirpath] = true
		fd, err := os.Open(dirpath)
		if err != nil {
			return err
		}

		defer fd.Close()

		list, err = fd.Readdir(0)
		if err != nil {
			return err
		}

		// Sort to make output stable between invocations
		sort.Sort(byName(list))
	}

	for _, file := range list {
		var asset Asset
		asset.Path = filepath.Join(dirpath, file.Name())
		asset.Name = filepath.ToSlash(asset.Path)

		ignoring := false
		for _, re := range ignore {
			if re.MatchString(asset.Path) {
				ignoring = true
				break
			}
		}
		if ignoring {
			continue
		}

		if file.IsDir() {
			if recursive {
				recursivePath := filepath.Join(dir, file.Name())
				visitedPaths[asset.Path] = true
				findFiles(recursivePath, prefix, recursive, toc, ignore, ignoreGlob, knownFuncs, visitedPaths)
			}
			continue
		} else if file.Mode()&os.ModeSymlink == os.ModeSymlink {
			var linkPath string
			if linkPath, err = os.Readlink(asset.Path); err != nil {
				return err
			}
			if !filepath.IsAbs(linkPath) {
				if linkPath, err = filepath.Abs(dirpath + "/" + linkPath); err != nil {
					return err
				}
			}
			if _, ok := visitedPaths[linkPath]; !ok {
				visitedPaths[linkPath] = true
				findFiles(asset.Path, prefix, recursive, toc, ignore, ignoreGlob, knownFuncs, visitedPaths)
			}
			continue
		}

		if strings.HasPrefix(asset.Name, prefix) {
			asset.Name = asset.Name[len(prefix):]
		} else {
			asset.Name = filepath.Join(dir, file.Name())
		}

		// If we have a leading slash, get rid of it.
		if len(asset.Name) > 0 && asset.Name[0] == '/' {
			asset.Name = asset.Name[1:]
		}

		// This shouldn't happen.
		if len(asset.Name) == 0 {
			return fmt.Errorf("Invalid file: %v", asset.Path)
		}

		asset.Func = safeFunctionName(asset.Name, knownFuncs)
		asset.Path, err = filepath.Abs(asset.Path)
		if err != nil {
			return err
		}
		asset.Size = file.Size()
		*toc = append(*toc, asset)
	}

	return nil
}

var regFuncName = regexp.MustCompile(`[^a-zA-Z0-9_]`)

// safeFunctionName converts the given name into a name
// which qualifies as a valid function identifier. It
// also compares against a known list of functions to
// prevent conflict based on name translation.
func safeFunctionName(name string, knownFuncs map[string]int) string {
	var inBytes, outBytes []byte
	var toUpper bool

	name = strings.ToLower(name)
	inBytes = []byte(name)

	for i := 0; i < len(inBytes); i++ {
		if regFuncName.Match([]byte{inBytes[i]}) {
			toUpper = true
		} else if toUpper {
			outBytes = append(outBytes, []byte(strings.ToUpper(string(inBytes[i])))...)
			toUpper = false
		} else {
			outBytes = append(outBytes, inBytes[i])
		}
	}

	name = string(outBytes)

	// Identifier can't start with a digit.
	if unicode.IsDigit(rune(name[0])) {
		name = "_" + name
	}

	if num, ok := knownFuncs[name]; ok {
		knownFuncs[name] = num + 1
		name = fmt.Sprintf("%s%d", name, num)
	} else {
		knownFuncs[name] = 2
	}

	return name
}
