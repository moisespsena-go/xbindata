package xbindata

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/moisespsena-go/bits2str"
)

func logTocAndInputs(dir, base string, c *Config, toc []Asset) (err error) {
	var wd string
	if wd, err = os.Getwd(); err != nil {
		return
	}
	if !filepath.IsAbs(dir) {
		if dir, err = filepath.Abs(filepath.Join(wd, dir)); err != nil {
			return
		}
	}

	var ignores = []string{
		base + "toc_names.yml",
		base + "toc_paths.yml",
		base + "inputs.yml",
	}

	if c.Outlined {
		ignores = append(ignores, path.Join(filepath.ToSlash(dir), base)+"dev.go")
	}

	if err = gitIgnore(dir, ignores...); err != nil {
		return err
	}

	func() {
		var (
			inputsPth = filepath.Join(dir, base+"inputs.yml")
			f         *os.File
		)

		if f, err = os.Create(inputsPth); err != nil {
			return
		}
		defer f.Close()

		if _, err = f.WriteString("# " + generatedHeader); err != nil {
			return
		}

		tmpl := "- { path: %q, recursive: %v, ns: %q, prefix: %q }\n"

		for _, input := range c.Input {
			relative := input.Path
			if !filepath.IsAbs(relative) {
				relative = filepath.Join(wd, relative)
			}
			relative, err = filepath.Rel(dir, relative)
			if err != nil {
				return
			}

			prefix := input.Prefix
			if !filepath.IsAbs(prefix) {
				prefix = filepath.Join(wd, prefix)
			}
			prefix, err = filepath.Rel(dir, prefix)
			if err != nil {
				return
			}

			if _, err = fmt.Fprintf(f, tmpl, relative, input.Recursive, input.NameSpace, prefix); err != nil {
				return
			}
		}
	}()

	if err != nil {
		return err
	}

	func() {
		var (
			tocPth               = filepath.Join(dir, base+"toc_")
			ptocNames            = tocPth + "names.yml"
			ptocPaths            = tocPth + "paths.yml"
			ftocNames, ftocPaths *os.File
		)

		if ftocNames, err = os.Create(ptocNames); err != nil {
			return
		}
		defer ftocNames.Close()

		if ftocPaths, err = os.Create(ptocPaths); err != nil {
			return
		}
		defer ftocPaths.Close()

		if _, err = ftocNames.WriteString("# " + generatedHeader); err != nil {
			return
		}
		if _, err = ftocPaths.WriteString("# " + generatedHeader); err != nil {
			return
		}

		for _, asset := range toc {
			relative, err := filepath.Rel(dir, asset.Path)
			if err != nil {
				return
			}
			if _, err = fmt.Fprintf(ftocNames, "- { name: %q, size: %s }\n", asset.Name, bits2str.Bits(asset.Size)*bits2str.Byte); err != nil {
				return
			}
			if _, err = ftocPaths.WriteString("- " + strconv.Quote(relative) + "\n"); err != nil {
				return
			}
		}
	}()

	return
}
