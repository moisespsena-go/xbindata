package xbcommon

import (
	"crypto/sha256"
	"fmt"
	"runtime"
	"strings"
	"sync"

	"github.com/moisespsena-go/os-common"
)

type Assets struct {
	root    NodeDir
	Assets  *map[string]Asset
	mu      sync.Mutex
	Factory func() (assets map[string]Asset, err error)
}

func NewAssets(assets ...Asset) *Assets {
	var assetsMap = make(map[string]Asset)
	for _, asset := range assets {
		assetsMap[asset.Path()] = asset
	}
	return &Assets{Assets: &assetsMap}
}

func (assets *Assets) check() {
	if assets.Assets == nil && assets.Factory != nil {
		assets.mu.Lock()
		defer assets.mu.Unlock()
		data, err := assets.Factory()
		if err != nil {
			panic(fmt.Errorf("Assets.check.Factory failed: %v", err))
		}
		if data == nil {
			data = map[string]Asset{}
		}
		assets.Assets = &data
		assets.Factory = nil
	}
}

func (assets *Assets) Root() NodeDir {
	assets.check()
	if assets.root == nil {
		assets.mu.Lock()
		defer assets.mu.Unlock()
		if assets.root == nil {
			defer runtime.GC()
			tree := newAssetTree()
			for name, asset := range *assets.Assets {
				pathList := strings.Split(name, "/")
				tree.Add(pathList, asset)
			}
			assets.root = tree.Node().(NodeDir)
			tree = nil
		}
	}
	return assets.root
}

func (assets *Assets) Get(name string) (asset Asset, ok bool) {
	assets.check()
	asset, ok = (*assets.Assets)[name]
	return
}

// Digests returns a map of all known files and their checksums.
func (assets *Assets) Digests() (map[string][sha256.Size]byte, error) {
	assets.check()
	mp := make(map[string][sha256.Size]byte, len(*assets.Assets))
	for name, asset := range *assets.Assets {
		mp[name] = asset.Digest()
	}
	return mp, nil
}

// Names returns the names of the assets.
func (assets *Assets) Names() []string {
	assets.check()
	names := make([]string, 0, len(*assets.Assets))
	for name := range *assets.Assets {
		names = append(names, name)
	}
	return names
}

// RestoreAsset restores an asset under the given directory.
func (assets *Assets) RestoreAsset(dir, name string) (err error) {
	assets.check()
	var n Node
	if n, err = assets.Root().Get(name); err != nil {
		return
	}
	if n.IsDir() {
		return oscommon.ErrNotFile(name)
	}
	return n.(Asset).Restore(dir)
}

// RestoreAssets restores an asset under the given directory recursively.
func (assets *Assets) RestoreDir(dir, name string) (err error) {
	assets.check()
	var n Node
	if n, err = assets.Root().GetDir(name); err != nil {
		return
	}
	if n.IsDir() {
		return oscommon.ErrNotFile(name)
	}
	return n.(Asset).Restore(dir)
}
