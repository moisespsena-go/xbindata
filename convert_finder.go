package xbindata

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/moisespsena-go/xbindata/walker"

	"github.com/gobwas/glob"
)

// Finder recursively finds all the file paths in the given directory tree.
// They are added to the given map as keys. Values will be safe function names
// for each file, which will be used when generating the output code.
type Finder struct {
	toc          *tocRegister
	ignore       []*regexp.Regexp
	ignoreGlob   []glob.Glob
	knownFuncs   map[string]int
	visitedPaths map[string]bool
}

// find now
func (f Finder) find(input *InputConfig, prefix string) (err error) {
	var dirpath string
	if len(prefix) > 0 {
		dirpath, _ = filepath.Abs(input.Path)
		prefix = filepath.ToSlash(prefix)
	}

	_, err = os.Stat(dirpath)
	if err != nil {
		return err
	}

	return input.Walk(&f.visitedPaths, func(info walker.FileInfo) (err error) {
		if info.IsDir() {
			return nil
		}
		for _, re := range f.ignore {
			if re.MatchString(info.Path) {
				return nil
			}
		}
		for _, ignore := range f.ignoreGlob {
			if ignore.Match(info.Path) {
				return nil
			}
		}
		var asset Asset
		asset.Name = filepath.ToSlash(info.Path)

		if strings.HasPrefix(asset.Name, prefix) {
			asset.Name = asset.Name[len(prefix):]
		}

		// If we have a leading slash, get rid of it.
		if len(asset.Name) > 0 && asset.Name[0] == '/' {
			asset.Name = asset.Name[1:]
		}

		if len(info.NamePrefix) > 0 {
			asset.Name = path.Join(append(info.NamePrefix, asset.Name)...)
		}

		// This shouldn't happen.
		if len(asset.Name) == 0 {
			return fmt.Errorf("Invalid file: %v", asset.Path)
		}

		asset.Func = safeFunctionName(asset.Name, f.knownFuncs)
		asset.Path, err = filepath.Abs(info.Path)
		if err != nil {
			return err
		}
		asset.info = info
		asset.Size = info.Size()
		f.toc.Append(asset)
		return nil
	})
}
