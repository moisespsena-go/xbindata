// This work is subject to the CC0 1.0 Universal (CC0 1.0) Public Domain Dedication
// license. Its contents can be found at:
// http://creativecommons.org/publicdomain/zero/1.0/

package xbindata

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"unicode"

	"github.com/moisespsena-go/xbindata/outlined"
	"github.com/moisespsena-go/xbindata/xbcommon"

	"github.com/moisespsena-go/bits2str"

	"github.com/moisespsena-go/path-helpers"
)

var (
	pkg  = path_helpers.GetCalledDir()
	pkgu = strings.NewReplacer("/", "", "\\", "").Replace(pkg)
)

type tocRegister struct {
	toc    []Asset
	byName map[string]int
}

func (t *tocRegister) Append(asset ...Asset) {
	for _, asset := range asset {
		if i, ok := t.byName[asset.Name]; ok {
			t.toc[i] = asset
		} else {
			t.byName[asset.Name] = len(t.toc)
			t.toc = append(t.toc, asset)
		}
	}
}

// Translate reads assets from an input directory, converts them
// to Go code and writes new files to the output specified
// in the given configuration.
func Translate(c *Config) error {

	// Ensure our configuration has sane values.
	if err := c.validate(); err != nil {
		return err
	}

	var (
		knownFuncs   = make(map[string]int)
		visitedPaths = make(map[string]bool)
		toc          []Asset
	)

	{
		tocr := &tocRegister{byName: map[string]int{}}

		// Locate all the assets.
		for _, input := range c.Input {
			finder := Finder{
				toc:          tocr,
				ignore:       append(c.Ignore, input.Ignore...),
				ignoreGlob:   append(c.IgnoreGlob, input.IgnoreGlob...),
				knownFuncs:   knownFuncs,
				visitedPaths: visitedPaths,
			}

			prefix := c.Prefix
			if input.Prefix != "" {
				prefix = input.Prefix
			}

			if err := finder.find(&input, prefix); err != nil {
				return err
			}
		}

		toc = tocr.toc

		sort.Slice(toc, func(i, j int) bool {
			return toc[i].Name < toc[j].Name
		})
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

	if c.Outlined {
		if c.Output == "" {
			c.Output = filepath.Join("xb", filepath.FromSlash(c.Package)+".xb")
		}
		if c.OutlinedApi == "" {
			c.OutlinedApi = "assets.go"
		}
	} else {
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
	_, err = fmt.Fprintf(buf, "package %s\n\n", path.Base(c.Package))
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

	if !c.Outlined {
		// Write table of contents
		if err := writeTOC(c, buf, toc); err != nil {
			return err
		}
	}

	if !c.Outlined || !c.OulinedSkipApi {
		dest := c.Output
		if c.Outlined {
			dest = c.OutlinedApi
		}

		if err = safefileWriteFile(dest, buf.Bytes(), 0); err != nil {
			return err
		}
	}

	if c.Outlined {
		buf.Reset()

		if !c.OulinedSkipApi {
			if err = outlinedHeadersWrite(buf, toc, c); err == nil && c.OutlinedHeadersOutput != "" {
				log.Printf("user headers file: `%v`\n", c.OutlinedHeadersOutput)

				if err = safefileWriteFile(c.OutlinedHeadersOutput, buf.Bytes(), 0); err != nil {
					return err
				}
			} else if err != nil {
				return fmt.Errorf("write headers to buffer failed: %v", err.Error())
			}

			outlinedDir := filepath.Dir(c.OutlinedApi)
			outlinedDirGitIgnore := filepath.Join(outlinedDir, ".gitignore")
			if f, err := os.Open(outlinedDirGitIgnore); err == nil {
				s, err := f.Stat()
				if err != nil {
					f.Close()
					return err
				}
				var has bool

				err = func() error {
					defer f.Close()
					var r = bufio.NewReader(f)
					var line []byte
					for {
						if line, _, err = r.ReadLine(); err == nil {
							if string(line) == "xb_compiler/" {
								has = true
								break
							}
						} else if err == io.EOF {
							break
						} else {
							return fmt.Errorf("read %q gitignore failed: `%v`\n", outlinedDirGitIgnore, err)
						}
					}
					return nil
				}()
				if err != nil {
					return err
				}
				if !has {
					if f, err = os.OpenFile(outlinedDirGitIgnore, os.O_APPEND|os.O_CREATE|os.O_WRONLY, s.Mode()); err != nil {
						return err
					}
					err = func() (err error) {
						defer f.Close()
						if _, err = f.WriteString("xb_compiler/\n"); err != nil {
							return fmt.Errorf("add `xb_compiler/` to gitignore failed: `%v`\n", err)
						}
						return
					}()
					if err != nil {
						return err
					}
				}
			} else if os.IsNotExist(err) {
				if err = safefileWriteFile(outlinedDirGitIgnore, []byte("xb_compiler/\n"), 0); err != nil {
					return err
				}
			} else {
				return err
			}

			headerPath := filepath.Join(outlinedDir, "xb_compiler", "main.go")

			log.Printf("creating main headers file: `%v`\n", headerPath)

			if err = safefileWriteFile(headerPath, buf.Bytes(), 0); err != nil {
				return err
			}

			if err != nil {
				return err
			}
		}

		if !c.OutlinedProgram || (c.OutputWriter != nil || c.Output != OutputToProgram) {
			headers := make(outlined.Headers, len(toc))

			cwd, _ := os.Getwd()

			for i, asset := range toc {
				info, _ := asset.Info()
				rpth, err := filepath.Rel(cwd, asset.Path)
				if err != nil {
					rpth = asset.Path
				}
				headers[i] = outlined.NewHeader(xbcommon.NewFileInfo(asset.Name, info.Size(), info.Mode(), info.ModTime(), asset.ctime), rpth)
			}

			if c.OutlinedProgram && c.OutputWriter != nil {
				err = headers.AppendW(c.OutputWriter)
			} else {
				outputFile := c.Output
				if outputFile, err = filepath.Abs(outputFile); err != nil {
					return fmt.Errorf("abs path failed: %v", err.Error())
				}
				log.Println("destination file: `" + outputFile + "`")
				if c.OutlinedProgram {
					err = headers.Append(outputFile)
				} else if c.NoCompress {
					err = headers.StoreFile(outputFile)
				} else {
					err = headers.StoreFileGz(outputFile)
				}
			}
		}
	}

	return err
}

var regFuncName = regexp.MustCompile(`[^a-zA-Z0-9_]`)

// safeFunctionName converts the given name into a name
// which qualifies as a valid function identifier. It
// also compares against a known list of functions to
// prevent conflict based on name translation.
func safeFunctionName(name string, knownFuncs map[string]int, mux *sync.RWMutex) string {
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

	mux.RLock()
	num, ok := knownFuncs[name]
	mux.RUnlock()

	mux.Lock()
	defer mux.Unlock()
	if ok {
		knownFuncs[name] = num + 1
		name = fmt.Sprintf("%s%d", name, num)
	} else {
		knownFuncs[name] = 2
	}

	return name
}
