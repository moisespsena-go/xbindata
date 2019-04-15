package xbfs

import (
	"context"
	"github.com/moisespsena-go/assetfs/assetfsapi"
	"path"

	"github.com/moisespsena-go/os-common"

	"github.com/moisespsena-go/xbindata/xbcommon"
)

// Names list matched files from assetfs
func glob(fs *FileSystem, pattern assetfsapi.GlobPattern, cb func(pth string, isDir bool) error) (err error) {
	var root = fs.assets.Root()
	if fs.root != nil {
		if root, err = root.GetDir(pattern.Dir()); err != nil {
			return
		}

		return _glob(root, pattern, cb)
	}
	return _glob(root, pattern, cb)
}

// Names list matched files from assetfs
func globInfo(fs *FileSystem, pattern assetfsapi.GlobPattern, cb func(info assetfsapi.FileInfo) error) (err error) {
	var root = fs.assets.Root()
	if fs.path != "" {
		pattern = pattern.Wrap(fs.path)
	}
	return _globInfo(root, pattern, cb)
}

func asset(fs *FileSystem, ctx context.Context, name string) ([]byte, error) {
	if fs.root != nil {
		name = path.Join(fs.path, name)
	}
	if asset, ok := fs.assets.GetC(ctx, name); !ok {
		return nil, oscommon.ErrNotFound(name)
	} else {
		return asset.Data()
	}
}

func assetInfo(fs *FileSystem, ctx context.Context, name string) (assetfsapi.FileInfo, error) {
	if fs.root != nil {
		name = path.Join(fs.path, name)
	}
	if asset, ok := fs.assets.GetC(ctx, name); ok {
		return NewFileInfo(asset, name), nil
	}
	return nil, oscommon.ErrNotFound(name)
}

func readDir(fs *FileSystem, dir string, cb assetfsapi.CbWalkInfoFunc, skipDir bool) (err error) {
	var n = fs.assets.Root()

	if n, err = n.GetDir(dir); err != nil {
		return
	}

	return n.Each(func(name string, n xbcommon.Node) error {
		var (
			f assetfsapi.FileInfo
		)
		if n.IsDir() {
			if skipDir {
				return nil
			}
			f = NewDirInfo(n.(xbcommon.NodeDir), path.Join(dir, name))
		} else {
			f = NewFileInfo(n.(xbcommon.Asset), path.Join(dir, name))
		}
		return cb(f)
	})
}

func walk(fs *FileSystem, dir string, cb assetfsapi.CbWalkFunc, mode assetfsapi.WalkMode) (err error) {
	var n = fs.assets.Root()

	if dir == "" {
		dir = "."
	}
	if dir != "." {
		if n, err = n.GetDir(dir); err != nil {
			return
		}
	}
	return n.Walk(func(dir, name string, n xbcommon.Node, _ interface{}) (d interface{}, err error) {
		isDir := n.IsDir()
		if isDir {
			if !mode.IsDirs() {
				return
			}
		} else {
			if !mode.IsFiles() {
				return
			}
		}
		return nil, cb(path.Join(dir, name), isDir)
	})
}

func walkInfo(fs *FileSystem, dir string, cb assetfsapi.CbWalkInfoFunc, mode assetfsapi.WalkMode) (err error) {
	var n = fs.assets.Root()

	if n, err = n.GetDir(dir); err != nil {
		return
	}
	return n.Walk(func(dir, name string, n xbcommon.Node, _ interface{}) (r interface{}, err error) {
		var info assetfsapi.FileInfo
		if n.IsDir() {
			if !mode.IsDirs() {
				return
			}
			info = NewDirInfo(n.(xbcommon.NodeDir), path.Join(dir, name))
		} else {
			if !mode.IsFiles() {
				return
			}
			info = NewFileInfo(n.(xbcommon.Asset), path.Join(dir, name))
		}
		return nil, cb(info)
	})
}
