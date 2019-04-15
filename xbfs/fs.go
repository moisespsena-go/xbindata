package xbfs

import (
	"context"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/moisespsena-go/assetfs/local"

	"github.com/moisespsena-go/assetfs/assetfsapi"

	"github.com/moisespsena-go/assetfs"
	"github.com/moisespsena-go/xbindata/xbcommon"
)

var (
	ERR_BINDATA_FILE = errors.New("Bindata file.")
	ERR_BINDATA_DIR  = errors.New("Bindata dir.")
)

type FileSystem struct {
	assetfsapi.AssetGetterInterface
	assetfsapi.TraversableInterface
	assets      *xbcommon.Assets
	root        *FileSystem
	path        string
	parent      *FileSystem
	nameSpaces  map[string]*FileSystem
	nameSpace   string
	callbacks   []assetfsapi.PathRegisterCallback
	HttpHandler http.Handler
	notExists   bool

	local.LocalSourcesAttribute
}

func NewFileSystem(assets *xbcommon.Assets) (fs *FileSystem) {
	fs = &FileSystem{assets: assets}
	fs.init()
	return
}

func (fs *FileSystem) init() {
	fs.AssetGetterInterface = &assetfs.AssetGetter{
		AssetFunc: func(ctx context.Context, path string) ([]byte, error) {
			return asset(fs, ctx, path)
		},
		AssetInfoFunc: func(ctx context.Context, path string) (assetfsapi.FileInfo, error) {
			return assetInfo(fs, ctx, path)
		},
	}
	fs.TraversableInterface = &assetfs.Traversable{
		fs,
		func(pth string, cb assetfsapi.CbWalkFunc, mode assetfsapi.WalkMode) error {
			return walk(fs, pth, cb, mode)
		},
		func(pth string, cb assetfsapi.CbWalkInfoFunc, mode assetfsapi.WalkMode) error {
			return walkInfo(fs, pth, cb, mode)
		},
		func(dir string, cb assetfsapi.CbWalkInfoFunc, skipDir bool) error {
			return readDir(fs, dir, cb, skipDir)
		},
		func(pattern assetfsapi.GlobPattern, cb func(pth string, isDir bool) error) (err error) {
			return glob(fs, pattern, cb)
		},
		func(pattern assetfsapi.GlobPattern, cb func(info assetfsapi.FileInfo) error) error {
			return globInfo(fs, pattern, cb)
		},
	}
	if fs.parent != nil {
		asset, err := fs.parent.AssetInfo(fs.nameSpace)
		if err != nil || !asset.IsDir() {
			fs.notExists = true
		}
	}
}

func (fs *FileSystem) GetPath() string {
	return fs.path
}

func (fs *FileSystem) Root() assetfsapi.Interface {
	return fs.root
}

// Compile compile assetfs
func (fs *FileSystem) Compile() error {
	return nil
}

func (fs *FileSystem) GetNameSpace(nameSpace string) (assetfsapi.NameSpacedInterface, error) {
	var (
		ns *FileSystem
		ok bool
	)
	for _, name := range strings.Split(strings.Trim(nameSpace, "/"), "/") {
		if fs.nameSpaces == nil {
			return nil, os.ErrNotExist
		} else if ns, ok = fs.nameSpaces[name]; !ok {
			return nil, os.ErrNotExist
		}
		fs = ns
	}
	return ns, nil
}

func (fs *FileSystem) NameSpaces() (items []assetfsapi.NameSpacedInterface) {
	for _, v := range fs.nameSpaces {
		items = append(items, v)
	}
	return
}

// NameSpace return namespaced filesystem
func (fs *FileSystem) NameSpace(nameSpace string) assetfsapi.NameSpacedInterface {
	if nameSpace == "" || nameSpace == "." {
		return nil
	}

	var (
		ns *FileSystem
		ok bool
	)
	for _, name := range strings.Split(strings.Trim(nameSpace, "/"), "/") {
		if fs.nameSpaces == nil {
			fs.nameSpaces = make(map[string]*FileSystem)
		}
		if ns, ok = fs.nameSpaces[name]; !ok {
			path := name
			root := fs.root
			if root == nil {
				root = fs
			} else {
				path = filepath.Join(fs.path, path)
			}
			ns = &FileSystem{assets: fs.assets, path: path, root: root, parent: fs, nameSpace: name}
			ns.init()
			fs.nameSpaces[nameSpace] = ns
		}
		fs = ns
	}

	return fs
}

func (fs *FileSystem) GetName() string {
	return fs.nameSpace
}

func (fs *FileSystem) GetParent() assetfsapi.Interface {
	return fs.parent
}

func (fs *FileSystem) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if fs.HttpHandler == nil {
		fs.HttpHandler = assetfs.HTTPStaticHandler(fs)
	}
	fs.HttpHandler.ServeHTTP(w, r)
}

func (fs *FileSystem) RegisterPlugin(plugins ...assetfsapi.Plugin) {
	for _, p := range plugins {
		p.Init(fs)
	}
	for _, p := range plugins {
		p.PathRegisterCallback(fs)
	}
}

func (fs *FileSystem) DumpFiles(cb func(info assetfsapi.FileInfo) error) error {
	return fs.dump(true, cb)
}

func (fs *FileSystem) Dump(cb func(info assetfsapi.FileInfo) error, ignore ...func(pth string) bool) error {
	return fs.dump(false, cb)
}

func (fs *FileSystem) dump(onlyFiles bool, cb func(info assetfsapi.FileInfo) error) error {
	mode := assetfsapi.WalkAll
	if onlyFiles {
		mode ^= assetfsapi.WalkDirs
	}
	return fs.WalkInfo(".", cb, mode)
}
