package archive

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/moisespsena-go/xbindata/xbcommon"

	"github.com/moisespsena/go-error-wrap"

	"github.com/moisespsena/go-path-helpers"
	"github.com/pkg/errors"
)

type Headers []*Header

func (headers Headers) Store(w io.Writer, base string) (err error) {
	for i, asset := range headers {
		if err = asset.LoadDigest(base); err != nil {
			return errwrap.Wrap(err, "Header[%d] Load Digest", i)
		}
	}

	if err = headers.write(w); err != nil {
		return
	}

	for i := range headers {
		if err = headers.do(i, base, w); err != nil {
			return errwrap.Wrap(err, "Header[%d]", i)
		}
	}
	return
}

func (headers Headers) StoreFile(pth string, baseDir string) (err error) {
	fmt.Printf("Writes to %q\n", pth)
	mode, err := path_helpers.ResolveFileMode(pth)
	if err != nil {
		return err
	}

	f, err := os.Create(pth)
	if err != nil {
		return err
	}
	if err = os.Chmod(pth, mode); err != nil {
		return
	}
	defer f.Close()
	return headers.Store(f, baseDir)
}

func (headers Headers) do(i int, base string, w io.Writer) (err error) {
	a := headers[i]
	defer func() {
		if err != nil {
			err = errwrap.Wrap(err, "%q", a.Path())
		}
	}()
	pth := a.Path()
	if base != "" {
		pth = filepath.Join(base, pth)
	}
	s, err := os.Stat(pth)
	if err != nil {
		return errors.New("os.Stat")
	}

	if s.Size() != a.Size() {
		return errors.New("File size changed.")
	}

	r, err := os.Open(pth)
	if err != nil {
		return
	}
	defer r.Close()
	_, err = io.Copy(w, r)
	return
}

// Tar takes a source and variable writers and walks 'source' writing each file
// found to the tar writer; the purpose for accepting multiple writers is to allow
// for multiple outputs (for example a file, or md5 hash)
func (headers Headers) write(w io.Writer) (err error) {
	count := uint32(len(headers))
	if err = binary.Write(w, binary.BigEndian, count); err != nil {
		err = fmt.Errorf("Write headers count failed: %v", err)
		return
	}

	for i, asset := range headers {
		if err = asset.Marshal(w); err != nil {
			return fmt.Errorf("write header of asset %d failed: %v", i, asset.Path())
		}
	}
	return
}

func (headers Headers) Assets(readerFactory AssetReaderFactory) (assets []xbcommon.Asset) {
	assets = make([]xbcommon.Asset, len(headers), len(headers))
	headers.EachAssets(readerFactory, func(i int, asset xbcommon.Asset) {
		assets[i] = asset
	})
	return
}

func (headers Headers) EachAssets(readerFactory AssetReaderFactory, cb func(i int, asset xbcommon.Asset)) {
	var start int64
	for i, h := range headers {
		cb(i, xbcommon.NewFile(h.FileInfo, readerFactory(start, h.Size()), h.digest))
		start += h.Size()
	}
	return
}
