package walker

import (
	"github.com/moisespsena-go/error-wrap"
	"github.com/moisespsena-go/xbindata/ignore"
	"os"
	"path/filepath"
	"regexp"

	"github.com/gobwas/glob"
)

const XbWalkName = ".xbwalk"

type (
	WalkCallback = func(info FileInfo) error
	WalkFunc     = func(visited *map[string]bool, recursive bool, cb WalkCallback) (err error)
)

type FileInfo struct {
	os.FileInfo
	Path       string
	NamePrefix []string
}

func (info FileInfo) SetNamePrefix(prefix ...string) FileInfo {
	info.NamePrefix = append(prefix, info.NamePrefix...)
	return info
}

type Walker struct {
	Recursive    bool
	VisitedPaths *map[string]bool
	IgnoreNames  map[string]bool
	IgnorePaths  map[string]bool
	IgnoreRes    []*regexp.Regexp
	IgnoreGlobs  []glob.Glob
	IgnoreFuncs  []func(pth string) bool
	depth        int
}

func New() *Walker {
	w := &Walker{}
	w.DefaultIgnores()
	return w
}

func (w *Walker) Recursives() *Walker {
	w.Recursive = true
	return w
}

func (w *Walker) DefaultIgnores() *Walker {
	w.IgnoreGlobs, _ = (ignore.IgnoreGlobSlice{".*"}).Items()
	return w
}

func (w *Walker) IgnoreName(names ...string) *Walker {
	if w.IgnoreNames == nil {
		w.IgnoreNames = map[string]bool{}
	}
	for _, name := range names {
		w.IgnoreNames[name] = true
	}
	return w
}

func (w *Walker) IgnorePath(pth ...string) *Walker {
	if w.IgnorePaths == nil {
		w.IgnorePaths = map[string]bool{}
	}
	for _, pth := range pth {
		w.IgnorePaths[pth] = true
	}
	return w
}

func (w *Walker) IgnoreRe(re ...*regexp.Regexp) *Walker {
	w.IgnoreRes = append(w.IgnoreRes, re...)
	return w
}

func (w *Walker) IgnoreReS(pattern ...string) *Walker {
	var gs = make(ignore.IgnoreSlice, len(pattern))
	for i, pattern := range pattern {
		gs[i] = pattern
	}
	var items, err = gs.Items()
	if err != nil {
		panic(err)
	}
	w.IgnoreRes = append(w.IgnoreRes, items...)
	return w
}

func (w *Walker) IgnoreGlob(g ...glob.Glob) *Walker {
	w.IgnoreGlobs = append(w.IgnoreGlobs, g...)
	return w
}

func (w *Walker) IgnoreGlobS(pattern ...string) *Walker {
	var gs = make(ignore.IgnoreGlobSlice, len(pattern))
	for i, pattern := range pattern {
		gs[i] = pattern
	}
	var items, err = gs.Items()
	if err != nil {
		panic(err)
	}
	w.IgnoreGlobs = append(w.IgnoreGlobs, items...)
	return w
}

func (w *Walker) IgnoreFunc(f ...func(pth string) bool) *Walker {
	w.IgnoreFuncs = append(w.IgnoreFuncs, f...)
	return w
}

// find now
func (w Walker) Walk(dirpath string, cb WalkCallback) (err error) {
	fi, err := os.Stat(dirpath)
	if err != nil {
		return err
	}

	if w.VisitedPaths == nil {
		vp := map[string]bool{}
		w.VisitedPaths = &vp
	}

	return w.walk(fi, dirpath, cb)
}

func (w Walker) Accepts(pth string) bool {
	if w.IgnorePaths != nil {
		if _, ok := w.IgnorePaths[pth]; ok {
			return false
		}
	}
	if w.IgnoreNames != nil {
		if _, ok := w.IgnoreNames[filepath.Base(pth)]; ok {
			return false
		}
	}
	for _, f := range w.IgnoreFuncs {
		if f(pth) {
			return false
		}
	}
	for _, re := range w.IgnoreRes {
		if re.MatchString(pth) {
			return false
		}
	}
	for _, ignore := range w.IgnoreGlobs {
		if ignore.Match(pth) {
			return false
		}
	}
	return true
}

// find now
func (w Walker) walk(fi os.FileInfo, pth string, cb WalkCallback) (err error) {
	if fi.Name() == XbWalkName {
		return nil
	}

	if pth != "." && !w.Accepts(pth) {
		if fi.IsDir() {
			return filepath.SkipDir
		}
		return nil
	}

	if !fi.IsDir() {
		return cb(FileInfo{FileInfo: fi, Path: pth})
	}

	var list []os.FileInfo

	(*w.VisitedPaths)[pth] = true
	fd, err := os.Open(pth)
	if err != nil {
		return err
	}

	func() {
		defer fd.Close()
		list, err = fd.Readdir(0)
	}()

	if err != nil {
		return err
	}

	for _, fi := range list {
		pth := filepath.Join(pth, fi.Name())
		if fi.IsDir() {
			if w.Recursive {
				(*w.VisitedPaths)[pth] = true
				if err = w.walk(fi, pth, cb); err != nil {
					if err == filepath.SkipDir {
						err = nil
						continue
					}
					return
				}
			}
			continue
		} else if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
			var linkPath string
			if linkPath, err = os.Readlink(pth); err != nil {
				return err
			}
			if !filepath.IsAbs(linkPath) {
				if linkPath, err = filepath.Abs(filepath.Join(pth, linkPath)); err != nil {
					return err
				}
			}
			if _, ok := (*w.VisitedPaths)[linkPath]; !ok {
				(*w.VisitedPaths)[linkPath] = true
				if err = w.Walk(pth, cb); err != nil {
					if err == filepath.SkipDir {
						err = nil
						continue
					}
					return
				}
			}
			continue
		}
		if err = errwrap.Wrap(w.walk(fi, pth, cb), "%q Walk failed", pth); err != nil {
			return
		}
	}

	return nil
}
