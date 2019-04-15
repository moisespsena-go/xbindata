package xbcommon

import (
	"crypto/sha256"
	"os"

	"github.com/moisespsena-go/io-common"
)

type Node interface {
	os.FileInfo
	Path() string
	Save(dest string) (err error)
	Restore(baseDir string) (err error)
	Depth() int
	Index() int
	IsFirst() bool
	IsLast() bool
}

type NodeDir interface {
	Node
	List() []Node
	Each(cb func(name string, n Node) error) error
	Get(name string) (n Node, err error)
	GetChild(name string) (n Node, ok bool)
	Child(name string) (n Node)
	Dir(name string) (n NodeDir)
	Asset(name string) (n Asset)
	Walk(cb func(dir, name string, n Node, data interface{}) (ret interface{}, err error)) (err error)
	WalkPrefix(prefix string, cb func(dir, name string, n Node, data interface{}) (ret interface{}, err error), data interface{}) (err error)
	GetDir(pth string) (d NodeDir, err error)
}

type Asset interface {
	Node
	Reader() (iocommon.ReadSeekCloser, error)
	Digest() [sha256.Size]byte
	Data() ([]byte, error)
	DataS() (string, error)
	MustData() []byte
	MustDataS() string
}

type nodeCommonInterface interface {
	setIndex(int)
	setDepth(int)
	setFirst(bool)
	setLast(bool)
}

type nodeCommon struct {
	Dpth  int
	Idx   int
	First bool
	Last  bool
}

func (ic *nodeCommon) setIndex(v int) {
	ic.Idx = v
}

func (ic *nodeCommon) setDepth(v int) {
	ic.Dpth = v
}

func (ic *nodeCommon) setFirst(v bool) {
	ic.First = v
}

func (ic *nodeCommon) setLast(v bool) {
	ic.Last = v
}

func (ic *nodeCommon) Depth() int {
	return ic.Dpth
}
func (ic *nodeCommon) Index() int {
	return ic.Idx
}
func (ic *nodeCommon) IsFirst() bool {
	return ic.First
}
func (ic *nodeCommon) IsLast() bool {
	return ic.Last
}
