package xbindata

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/gobwas/glob"
)

// Finder recursively finds all the file paths in the given directory tree.
// They are added to the given map as keys. Values will be safe function names
// for each file, which will be used when generating the output code.
type Finder struct {
	recursive    bool
	toc          *[]Asset
	ignore       []*regexp.Regexp
	ignoreGlob   []glob.Glob
	knownFuncs   map[string]int
	visitedPaths map[string]bool
}

// find now
func (f Finder) find(dir, prefix string) (err error) {
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
		f.visitedPaths[dirpath] = true
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
		for _, re := range f.ignore {
			if re.MatchString(asset.Path) {
				ignoring = true
				break
			}
		}
		for _, ignore := range f.ignoreGlob {
			if ignore.Match(asset.Path) {
				ignoring = true
				break
			}
		}
		if ignoring {
			continue
		}

		if file.IsDir() {
			if f.recursive {
				recursivePath := filepath.Join(dir, file.Name())
				f.visitedPaths[asset.Path] = true
				if err = f.find(recursivePath, prefix); err != nil {
					return
				}
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
			if _, ok := f.visitedPaths[linkPath]; !ok {
				f.visitedPaths[linkPath] = true
				if err = f.find(asset.Path, prefix); err != nil {
					return
				}
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

		asset.Func = safeFunctionName(asset.Name, f.knownFuncs)
		asset.Path, err = filepath.Abs(asset.Path)
		if err != nil {
			return err
		}
		asset.Size = file.Size()
		*f.toc = append(*f.toc, asset)
	}

	return nil
}
