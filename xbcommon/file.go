package xbcommon

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/moisespsena-go/xbindata/digest"

	"github.com/moisespsena-go/file-utils"
	"github.com/moisespsena-go/path-helpers"

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

func (f *File) ImportLocal(localPath, name string, info os.FileInfo) (err error) {
	if f.digest, err = digest.Digest(localPath); err != nil {
		return
	}
	if info == nil {
		if info, err = os.Stat(localPath); err != nil {
			return
		}
	}
	f.FileInfo = NewFileInfo(name, info.Size(), info.Mode(), info.ModTime(), time.Time{})
	f.reader = func() (r iocommon.ReadSeekCloser, e error) {
		r, e = os.Open(localPath)
		return
	}
	return
}

func (f *File) Reader() (iocommon.ReadSeekCloser, error) {
	return f.reader()
}

func (f *File) Digest() (d [sha256.Size]byte) {
	if f.digest == nil {
		return
	}
	return *f.digest
}

func (f File) Data() ([]byte, error) {
	if r, err := f.Reader(); err != nil {
		return nil, err
	} else {
		defer func() {
			_ = r.Close()
		}()
		return ioutil.ReadAll(r)
	}
}

func (f *File) DataS() (string, error) {
	if b, err := f.Data(); err != nil {
		return "", err
	} else {
		return string(b), nil
	}
}

func (f *File) MustData() []byte {
	if b, err := f.Data(); err != nil {
		panic(fmt.Errorf("[file %q] MustaData: %v", f.path, err))
	} else {
		return b
	}
}

func (f *File) MustDataS() string {
	return string(f.MustData())
}

func (f *File) Save(dest string) (err error) {
	var r io.ReadCloser
	if r, err = f.Reader(); err != nil {
		return
	}
	return fileutils.CreateFileSync(dest, r, f)
}

// Restore restores an asset under the given directory.
func (f *File) Restore(baseDir string) (err error) {
	fdir := FilePath(baseDir, filepath.Dir(f.path))
	if err = path_helpers.MkdirAllIfNotExists(fdir); err != nil {
		return
	}
	return f.Save(filepath.Join(fdir, f.Name()))
}
