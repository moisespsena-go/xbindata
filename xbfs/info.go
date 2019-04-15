package xbfs

import (
	"io"

	"github.com/moisespsena-go/assetfs/assetfsapi"

	"github.com/moisespsena-go/assetfs"
	"github.com/moisespsena-go/io-common"
	"github.com/moisespsena-go/xbindata/xbcommon"
)

type FileInfo struct {
	xbcommon.Asset
	path string
}

func NewFileInfo(a xbcommon.Asset, path string) *FileInfo { return &FileInfo{a, path} }

func (b *FileInfo) Path() string     { return b.path }
func (b *FileInfo) RealPath() string { return b.Asset.Path() }
func (FileInfo) Type() assetfsapi.FileType {
	return assetfsapi.FileTypeBindata | assetfsapi.FileTypeNormal
}
func (FileInfo) Writer() (io.WriteCloser, error)   { return nil, ERR_BINDATA_FILE }
func (FileInfo) Appender() (io.WriteCloser, error) { return nil, ERR_BINDATA_FILE }
func (b *FileInfo) String() string                 { return assetfs.StringifyFileInfo(b) }

type DirInfo struct {
	xbcommon.NodeDir
	path string
}

func NewDirInfo(node xbcommon.NodeDir, pth string) *DirInfo { return &DirInfo{node, pth} }

func (d *DirInfo) Path() string                              { return d.path }
func (d *DirInfo) RealPath() string                          { return d.NodeDir.Path() }
func (DirInfo) Type() assetfsapi.FileType                    { return assetfsapi.FileTypeBindata | assetfsapi.FileTypeDir }
func (d *DirInfo) String() string                            { return assetfs.StringifyFileInfo(d) }
func (rf *DirInfo) Reader() (iocommon.ReadSeekCloser, error) { return nil, ERR_BINDATA_DIR }
func (DirInfo) Writer() (io.WriteCloser, error)              { return nil, ERR_BINDATA_DIR }
func (DirInfo) Appender() (io.WriteCloser, error)            { return nil, ERR_BINDATA_DIR }
func (b *DirInfo) Data() ([]byte, error)                     { return nil, ERR_BINDATA_DIR }
