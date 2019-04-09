// This work is subject to the CC0 1.0 Universal (CC0 1.0) Public Domain Dedication
// license. Its contents can be found at:
// http://creativecommons.org/publicdomain/zero/1.0/

package xbindata

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"path"
	"sort"
	"strings"
)

type assetTree struct {
	Asset    Asset
	Path     string
	Children map[string]*assetTree
}

func newAssetTree() *assetTree {
	tree := &assetTree{Path: "."}
	tree.Children = make(map[string]*assetTree)
	return tree
}

func (node *assetTree) child(name string) *assetTree {
	rv, ok := node.Children[name]
	if !ok {
		rv = newAssetTree()
		rv.Path = path.Join(node.Path, name)
		// TODO: maintain this in sorted order
		node.Children[name] = rv
	}
	return rv
}

func (root *assetTree) Add(route []string, asset Asset) {
	for _, name := range route {
		root = root.child(name)
	}
	root.Asset = asset
}

func ident(buf *bytes.Buffer, n int) {
	for i := 0; i < n; i++ {
		buf.WriteByte('\t')
	}
}

func (root *assetTree) funcOrNil() string {
	if root.Asset.Func == "" {
		return "nil"
	} else {
		return root.Asset.Func
	}
}

func (root *assetTree) writeGoMap(buf *bytes.Buffer, nident int) {
	buf.Grow(35) // at least this size
	if root.Asset.Func != "" {
		_, _ = fmt.Fprintf(buf, "%s", root.Asset.Func)
	} else {
		_, _ = fmt.Fprintf(buf, "bc.NewDir(%d, %q, map[string]bc.Node{", nident, root.Path)

		if len(root.Children) > 0 {
			buf.WriteByte('\n')

			// Sort to make output stable between invocations
			filenames := make([]string, len(root.Children))
			i := 0
			for filename := range root.Children {
				filenames[i] = filename
				i++
			}
			sort.Strings(filenames)

			for _, p := range filenames {
				ident(buf, nident+1)
				buf.WriteByte('"')
				buf.WriteString(p)
				buf.WriteString(`": `)
				root.Children[p].writeGoMap(buf, nident+1)
			}
			ident(buf, nident)
		}

		buf.WriteString("})")
	}
	if nident > 0 {
		buf.WriteByte(',')
	}
	buf.WriteByte('\n')
}

func (root *assetTree) WriteAsGoMap(w io.Writer) error {
	_, err := w.Write([]byte(`
var _root bc.NodeDir = `))
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	root.writeGoMap(buf, 0)
	fmted, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}
	_, writeErr := w.Write(fmted)
	return writeErr
}

func writeTOCTree(w io.Writer, toc []Asset) error {
	tree := newAssetTree()
	for i := range toc {
		pathList := strings.Split(toc[i].Name, "/")
		tree.Add(pathList, toc[i])
	}
	return tree.WriteAsGoMap(w)
}

// writeTOC writes the table of contents file.
func writeTOC(c *Config, buf *bytes.Buffer, toc []Asset) error {
	writeTOCHeader(buf)

	var start int64
	for i := range toc {
		if i != 0 {
			// Newlines between elements make gofmt happy.
			buf.WriteByte('\n')
		}
		if err := writeTOCAsset(c, start, buf, &toc[i]); err != nil {
			return err
		}
		start += toc[i].Size
	}

	writeTOCFooter(buf)
	return nil
}

// writeTOCHeader writes the table of contents file header.
func writeTOCHeader(buf *bytes.Buffer) {
	buf.WriteString(`
// Embed is a table, holding each asset generator, mapped to its name.
var Embed *bc.Assets

func LoadDefault() {
	Embed = bc.NewAssets(
`)
}

// writeTOCAsset writes a TOC entry for the given asset.
func writeTOCAsset(c *Config, start int64, w io.Writer, asset *Asset) (err error) {
	_, err = w.Write([]byte("\t\t" + asset.Func + ","))
	return err
}

// writeTOCFooter writes the table of contents file footer.
func writeTOCFooter(buf *bytes.Buffer) {
	buf.WriteString(`
	)

	DefaultFS = fs.NewFileSystem(Embed)
}
`)
}
