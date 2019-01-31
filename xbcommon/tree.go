package xbcommon

import (
	"path"
)

type assetTree struct {
	Asset    Asset
	Path     string
	children map[string]*assetTree
	level    int
}

func newAssetTree() *assetTree {
	tree := &assetTree{Path: "."}
	tree.children = make(map[string]*assetTree)
	return tree
}

func (node *assetTree) child(name string) *assetTree {
	rv, ok := node.children[name]
	if !ok {
		rv = newAssetTree()
		rv.Path = path.Join(node.Path, name)
		node.children[name] = rv
	}
	return rv
}

func (root *assetTree) Add(route []string, asset Asset) {
	for _, name := range route {
		root = root.child(name)
	}
	root.level = len(route)
	root.Asset = asset
}

func (root *assetTree) Children() (children map[string]Node) {
	children = map[string]Node{}
	for name, child := range root.children {
		children[name] = child.Node()
	}
	return
}

func (root *assetTree) Node() Node {
	if root.Asset != nil {
		return root.Asset
	}
	return NewDir(root.level, root.Path, root.Children())
}
