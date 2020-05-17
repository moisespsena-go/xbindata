package xbcommon

import (
	"context"
	"strings"
)

type Tree struct {
	root *assetTree
}

func NewTree(assets ...Asset) *Tree {
	tree := newAssetTree()
	for _, asset := range assets {
		pathList := strings.Split(asset.Path(), "/")
		tree.Add(pathList, asset)
	}
	return &Tree{tree}
}

func (assets *Tree) Get(name string) (asset Asset, ok bool) {
	asset = assets.MustGet(name)
	ok = asset != nil
	return
}

func (assets *Tree) MustGet(name string) (asset Asset) {
	node := assets.root.Node().(NodeDir).Find(name)
	if node != nil {
		asset, _ = node.(Asset)
	}
	return
}

func (assets *Tree) GetC(_ context.Context, name string) (Asset, bool) {
	return assets.Get(name)
}

func (assets *Tree) MustGetC(_ context.Context, name string) (asset Asset) {
	return assets.MustGet(name)
}

func (assets *Tree) Root() NodeDir {
	return assets.root.Node().(NodeDir)
}
