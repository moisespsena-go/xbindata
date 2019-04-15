package xbcommon

import (
	"crypto/sha256"

	"github.com/moisespsena-go/assetfs"
	"github.com/moisespsena-go/assetfs/local"
)

type LocalFile struct {
	*assetfs.RealFileInfo
	asset        Asset
	digest       *[sha256.Size]byte
	digestLoaded bool
	nodeCommon
}

func (f LocalFile) Save(dest string) (err error) {
	panic("implement me")
}

func (f LocalFile) Restore(baseDir string) (err error) {
	panic("implement me")
}

func (f LocalFile) Digest() [sha256.Size]byte {
	if !f.digestLoaded && f.digest == nil {
		f.digestLoaded = true
		f.digest, _ = local.Digest(f.RealPath())
	}
	return *f.digest
}
