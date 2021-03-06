package outlined

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/moisespsena-go/xbindata/xbcommon"

	path_helpers "github.com/moisespsena-go/path-helpers"
	"github.com/pkg/errors"
)

type Headers []*Header

var binaryDir = xbcommon.BinaryDir

func (headers Headers) Store(w io.Writer) (err error) {
	cHash := sha256.New()

	for i, asset := range headers {
		if err = asset.LoadDigest(cHash); err != nil {
			return errors.Wrapf(err, "Header[%d] Load Digest", i)
		}
	}

	if _, err = w.Write(cHash.Sum(nil)); err != nil {
		return fmt.Errorf("Write content hash failed: %v", err)
	}

	if _, err = w.Write([]byte("\n")); err != nil {
		err = fmt.Errorf("write NL failed: %v", err)
		return
	}

	if err = binary.Write(w, binaryDir, uint64(time.Now().UTC().Unix())); err != nil {
		return fmt.Errorf("Write build time failed: %v", err)
	}

	if _, err = w.Write([]byte("\n")); err != nil {
		err = fmt.Errorf("write NL failed: %v", err)
		return
	}

	if err = headers.write(w); err != nil {
		return
	}

	for i := range headers {
		if err = headers.do(i, w); err != nil {
			return errors.Wrapf(err, "Header[%d]", i)
		}
	}

	return
}

func (headers Headers) StoreFile(pth string, wrap ...func(w io.WriteCloser) io.WriteCloser) (err error) {
	log.Infof("Writes to %q\n", pth)
	mode, err := path_helpers.ResolveFileMode(pth)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(pth, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	var w io.WriteCloser = f
	for _, wrap := range wrap {
		w = wrap(w)
	}
	defer w.Close()
	return headers.Store(w)
}

func (headers Headers) StoreFileGz(pth string) (err error) {
	return headers.StoreFile(pth+".gz", func(w io.WriteCloser) io.WriteCloser {
		return gzip.NewWriter(w)
	})
}

func (headers Headers) Append(pth string, wrap ...func(w io.WriteCloser) io.WriteCloser) (err error) {
	log.Infof("Appends to %q\n", pth)
	mode, err := path_helpers.ResolveFileMode(pth)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(pth, os.O_APPEND|os.O_CREATE|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	defer f.Close()

	s, _ := f.Stat()
	log.Infof("Old size: %d", s.Size())

	if err = headers.AppendW(f); err == nil {
		s, _ = f.Stat()
		newSize := s.Size()
		log.Infof("New size: %d", newSize)
	}
	return
}

func (headers Headers) AppendW(w io.Writer) (err error) {
	wc := &writeCounter{Writer: w}
	w = wc
	if err = headers.Store(w); err != nil {
		return
	}
	size := wc.count
	log.Infof("Outlined size: %d", size)
	err = errors.Wrapf(binary.Write(w, binaryDir, uint32(size)), "write end size")
	return
}

func (headers Headers) AppendGz(pth string) (err error) {
	return headers.Append(pth, func(w io.WriteCloser) io.WriteCloser {
		return gzip.NewWriter(w)
	})
}

func (headers Headers) do(i int, w io.Writer) (err error) {
	a := headers[i]
	defer func() {
		if err != nil {
			err = errors.Wrapf(err, "%q", a.Path())
		}
	}()
	s, err := os.Stat(a.SysPath)
	if err != nil {
		return errors.New("os.Stat")
	}

	if s.Size() != a.Size() {
		return errors.New("File size changed.")
	}

	r, err := os.Open(a.SysPath)
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
	if err = binary.Write(w, binaryDir, count); err != nil {
		err = fmt.Errorf("write headers count failed: %v", err)
		return
	}
	if _, err = w.Write([]byte("\n")); err != nil {
		err = fmt.Errorf("write NL failed: %v", err)
		return
	}
	for i, asset := range headers {
		if err = asset.Marshal(w); err != nil {
			return fmt.Errorf("write header of asset %d failed: %v", i, asset.Path())
		}
		if _, err = w.Write([]byte("\n")); err != nil {
			err = fmt.Errorf("write NL failed: %v", err)
			return
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
