package xbcommon

import (
	"crypto/sha256"
	"io"
	"io/ioutil"
	"path/filepath"

	"github.com/moisespsena-go/file-utils"
	"github.com/moisespsena/go-path-helpers"

	"github.com/moisespsena-go/io-common"
)

type File struct {
	*FileInfo
	nodeCommon
	reader func() (iocommon.ReadSeekCloser, error)
	digest *[sha256.Size]byte
}

func NewFile(fileInfo *FileInfo, reader func() (iocommon.ReadSeekCloser, error), digest *[sha256.Size]byte) *File {
	return &File{FileInfo: fileInfo, reader: reader, digest: digest}
}

func (a *File) Reader() (iocommon.ReadSeekCloser, error) {
	return a.reader()
}

func (a *File) Digest() (d [sha256.Size]byte) {
	if a.digest == nil {
		return
	}
	return *a.digest
}

func (a *File) Data() ([]byte, error) {
	if r, err := a.reader(); err != nil {
		return nil, err
	} else {
		defer func() {
			_ = r.Close()
		}()
		return ioutil.ReadAll(r)
	}
}

func (a *File) String() (string, error) {
	if b, err := a.Data(); err != nil {
		return "", err
	} else {
		return string(b), nil
	}
}

func (a *File) Save(dest string) (err error) {
	var r io.ReadCloser
	if r, err = a.Reader(); err != nil {
		return
	}
	return fileutils.CreateFileSync(dest, r, a)
}

// Restore restores an asset under the given directory.
func (a *File) Restore(baseDir string) (err error) {
	fdir := FilePath(baseDir, filepath.Dir(a.path))
	if err = path_helpers.MkdirAllIfNotExists(fdir); err != nil {
		return
	}
	return a.Save(filepath.Join(fdir, a.Name()))
}
