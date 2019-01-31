package xbcommon

import (
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/moisespsena-go/os-common"
)

type Dir struct {
	*DirInfo
	nodeCommon
	children map[string]Node
	sorted   []string
}

func NewDir(depth int, pth string, children map[string]Node) *Dir {
	var (
		names []string
		l     = len(children) - 1
	)
	for name := range children {
		names = append(names, name)
	}

	sort.Strings(names)

	for i, name := range names {
		c := children[name].(nodeCommonInterface)
		c.setDepth(depth + 1)
		c.setIndex(i)
		c.setFirst(i == 0)
		c.setLast(i == l)
	}

	d := &Dir{DirInfo: &DirInfo{pth, path.Base(pth)}, children: children, sorted: names}
	d.setDepth(depth)
	return d
}

func (t *Dir) Info() os.FileInfo {
	return t
}

func (t *Dir) GetChild(name string) (n Node, ok bool) {
	n, ok = t.children[name]
	return
}

func (t *Dir) Child(name string) (n Node) {
	return t.children[name]
}

func (t *Dir) Dir(name string) (d NodeDir) {
	d, _ = t.GetDir(name)
	return
}

func (t *Dir) Asset(name string) (a Asset) {
	n, _ := t.Get(name)
	return n.(Asset)
}

func (t *Dir) Get(name string) (n Node, err error) {
	var (
		dir         = path.Dir(name)
		d   NodeDir = t
	)
	if dir != "." {
		if d, err = d.GetDir(dir); err != nil {
			return nil, err
		}
	}
	var ok bool
	if n, ok = d.GetChild(path.Base(name)); !ok {
		err = oscommon.ErrNotFound(name)
	}
	return
}

func (t *Dir) List() []Node {
	var (
		items = make([]Node, len(t.children), len(t.children))
	)
	for i, name := range t.sorted {
		items[i] = t.children[name]
	}
	return items
}

func (t *Dir) Each(cb func(name string, n Node) error) (err error) {
	for _, name := range t.sorted {
		item := t.children[name]
		if item.IsDir() {
			if err = cb(name, item); err != nil {
				return
			}
		} else {
			if err = cb(name, item); err != nil {
				return
			}
		}
	}
	return nil
}

func (t *Dir) WalkPrefix(prefix string, cb func(dir, name string, n Node, data interface{}) (interface{}, error), data interface{}) (err error) {
	return t.Each(func(name string, n Node) (err error) {
		if n.IsDir() {
			var data2 interface{}
			if data2, err = cb(prefix, name, n, data); err != nil {
				if err == filepath.SkipDir {
					return nil
				}
				return
			}
			if err = n.(NodeDir).WalkPrefix(path.Join(prefix, name), cb, data2); err != nil {
				return
			}
		} else {
			_, err = cb(prefix, name, n, data)
		}
		return
	})
}

func (t *Dir) Walk(cb func(dir, name string, n Node, data interface{}) (interface{}, error)) (err error) {
	return t.WalkPrefix(".", cb, nil)
}

func (t *Dir) GetDir(pth string) (d NodeDir, err error) {
	d = t
	var (
		parts = strings.Split(pth, "/")
		n     Node
		ok    bool
	)

	for i, name := range parts {
		if n, ok = d.GetChild(name); !ok {
			return nil, oscommon.ErrNotFound(strings.Join(parts[0:i+1], "/"))
		} else if !n.IsDir() {
			return nil, oscommon.ErrNotDir(pth)
		}
		d = n.(NodeDir)
	}

	return
}

func (t *Dir) Save(dest string) (err error) {
	return t.Walk(func(dir, name string, n Node, _ interface{}) (interface{}, error) {
		if !n.IsDir() {
			return nil, n.(Asset).Save(FilePath(dest, path.Join(dir, name)))
		}
		return nil, nil
	})
}

// Restore restores an asset under the given directory.
func (t *Dir) Restore(baseDir string) (err error) {
	return t.Walk(func(dir, name string, n Node, _ interface{}) (interface{}, error) {
		if !n.IsDir() {
			return nil, n.(Asset).Save(FilePath(baseDir, dir, n.Name()))
		}
		return nil, nil
	})
}
