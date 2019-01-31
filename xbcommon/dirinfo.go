package xbcommon

import (
	"os"
	"time"

	"github.com/moisespsena/go-assetfs/assetfsapi"
)

type DirInfo struct {
	path string
	name string
}

func (d DirInfo) Path() string {
	return d.path
}

func (b DirInfo) Name() string {
	return b.name
}

func (DirInfo) Type() assetfsapi.FileType {
	return assetfsapi.FileTypeBindata | assetfsapi.FileTypeDir
}

func (DirInfo) Size() int64 {
	return -1
}

func (DirInfo) Mode() os.FileMode {
	return os.ModeDir | 0755
}

func (DirInfo) ModTime() (t time.Time) {
	return
}

func (DirInfo) IsDir() bool {
	return true
}

func (DirInfo) Sys() interface{} {
	return nil
}
