// Code generated by xbindata. DO NOT EDIT.
// sources:
package normal

import (
	"errors"
	"github.com/moisespsena-go/assetfs"
	fsapi "github.com/moisespsena-go/assetfs/assetfsapi"
	"github.com/moisespsena-go/io-common"
	"github.com/moisespsena-go/path-helpers"
	"github.com/moisespsena-go/xbindata/outlined"
	bc "github.com/moisespsena-go/xbindata/xbcommon"
	"github.com/moisespsena-go/xbindata/xbfs"
	br "github.com/moisespsena-go/xbindata/xbreader"
	"os"
	"path/filepath"
	"sync"
)

var (
	_outlined     *outlined.Outlined
	outlinedPath  string
	outlinedPaths []string
	ended         bool

	StartPos int64
	Assets   bc.Assets

	pkg                   = path_helpers.GetCalledDir()
	OpenOutlined          = br.Open
	OutlinedReaderFactory = func(start, size int64) func() (reader iocommon.ReadSeekCloser, err error) {
		return func() (reader iocommon.ReadSeekCloser, err error) {
			return OpenOutlined(outlinedPath, _outlined.StartPos+start, size)
		}
	}
)

func OutlinedPath() string {
	return outlinedPath
}

func Outlined() (archiv *outlined.Outlined, err error) {
	if _outlined == nil {
		mu.Lock()
		defer mu.Unlock()

		if _outlined, err = outlined.OpenFile(outlinedPath, ended); err != nil {
			return
		}
	}
	return _outlined, nil
}

func LoadDefault() {
	if outlinedPath == "" {
		pths := []string{
			path.Join("_assets", pkg+".xb.gz"),
			path.Join("_assets", pkg+".xb"),
		}

		for _, pth := range pths {
			if pth == "" {
				continue
			}
			pth = bc.FilePath(pth)

			if _, err := os.Stat(pth); err == nil {

				outlinedPath = pth
				break
			} else if !os.IsNotExist(err) {
				panic(err)
			}
		}

		if outlinedPath == "" {
			panic(errors.New("outlined path not defined"))
		}
	}

	Assets.Factory = func() (assets map[string]bc.Asset, err error) {
		archiv, err := Outlined()
		if err != nil {
			return nil, err
		}

		return archiv.AssetsMap(OutlinedReaderFactory), nil
	}

	loaded = true
}

var (
	loaded bool
	mu     sync.Mutex
	fs     fsapi.Interface
)
var fsLoadCallbacks []func(fs fsapi.Interface)

func OnFsLoad(cb ...func(fs fsapi.Interface)) {
	fsLoadCallbacks = append(fsLoadCallbacks, cb...)
}

func callFsLoadCallbacks() {
	for _, f := range fsLoadCallbacks {
		f(fs)
	}
}

func IsLocal() bool {
	if _, err := os.Stat("assets/inputs/a"); err == nil {
		return true
	}
	return false
}

func FS() fsapi.Interface {
	Load()
	return fs
}

var DefaultFS fsapi.Interface = xbfs.NewFileSystem(&Assets)
var LocalFS = assetfs.NewAssetFileSystem()

func LoadLocal() {
	var inputs = []string{
		"assets/inputs/a",
		"assets/inputs/b",
		"assets/inputs/c",
	}
	localDir := filepath.Join("_assets", filepath.FromSlash(pkg))
	if _, err := os.Stat(localDir); err == nil {
		for i, pth := range inputs {
			inputs[i] = filepath.Join(localDir, pth)
		}
	} else if !os.IsNotExist(err) {
		panic(err)
	}
	for _, pth := range inputs {
		if err := LocalFS.RegisterPath(pth); err != nil {
			panic(err)
		}
	}
}

func Load() {
	if loaded {
		return
	}
	mu.Lock()
	if loaded {
		mu.Unlock()
		return
	}
	defer callFsLoadCallbacks()
	defer mu.Unlock()
	defer func() { loaded = true }()
	if IsLocal() {
		LoadLocal()
		fs = LocalFS
		return
	}
	LoadDefault()
	fs = DefaultFS
}

func init() { Load() }
