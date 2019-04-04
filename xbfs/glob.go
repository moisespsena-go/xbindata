package xbfs

import (
	"path"

	"github.com/moisespsena-go/xbindata/xbcommon"
	"github.com/moisespsena-go/assetfs/assetfsapi"
)

func _glob(root xbcommon.NodeDir, pattern assetfsapi.GlobPattern, cb func(pth string, isDir bool) error) (err error) {
	if pattern.IsRecursive() {
		return root.Walk(func(dir, name string, n xbcommon.Node, _ interface{}) (r interface{}, err error) {
			isDir := n.IsDir()
			if isDir {
				if !pattern.AllowDirs() {
					return
				}
			} else if !pattern.AllowFiles() {
				return
			}
			if pattern.Match(name) {
				return nil, cb(path.Join(dir, name), isDir)
			}
			return
		})
	}

	return root.Each(func(name string, n xbcommon.Node) (err error) {
		isDir := n.IsDir()
		if isDir {
			if !pattern.AllowDirs() {
				return
			}
		} else if !pattern.AllowFiles() {
			return
		}
		if pattern.Match(name) {
			return cb(name, isDir)
		}
		return nil
	})
}

func _globInfo(root xbcommon.NodeDir, pattern assetfsapi.GlobPattern, cb func(info assetfsapi.FileInfo) error) (err error) {
	if root, err = root.GetDir(pattern.Dir()); err != nil {
		return
	}

	if pattern.IsRecursive() {
		return root.Walk(func(dir, name string, n xbcommon.Node, _ interface{}) (r interface{}, err error) {
			isDir := n.IsDir()
			if isDir {
				if !pattern.AllowDirs() {
					return
				}
			} else if !pattern.AllowFiles() {
				return
			}
			if pattern.Match(name) {
				if isDir {
					return nil, cb(&DirInfo{n.(xbcommon.NodeDir), path.Join(dir, name)})
				}
				return nil, cb(&FileInfo{n.(xbcommon.Asset), path.Join(dir, name)})
			}
			return
		})
	}

	return root.Each(func(name string, n xbcommon.Node) (err error) {
		isDir := n.IsDir()
		if isDir {
			if !pattern.AllowDirs() {
				return
			}
		} else if !pattern.AllowFiles() {
			return
		}
		if pattern.Match(name) {
			if isDir {
				return cb(NewDirInfo(n.(xbcommon.NodeDir), name))
			}
			return cb(NewFileInfo(n.(xbcommon.Asset), name))
		}
		return nil
	})
}
