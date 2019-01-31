package archive

import (
	"crypto/sha256"
	"errors"
	"io"
	"os"

	"path/filepath"

	"github.com/moisespsena-go/xbindata/xbcommon"
)

type Header struct {
	*xbcommon.FileInfo
	digest *[sha256.Size]byte
}

func NewHeader(fileInfo *xbcommon.FileInfo) *Header {
	return &Header{FileInfo: fileInfo}
}

func (a *Header) Digest() *[sha256.Size]byte {
	return a.digest
}

func (a *Header) DigestReader() func() [sha256.Size]byte {
	d := *a.digest
	return func() [32]byte {
		return d
	}
}

func (a *Header) LoadDigest(baseDir string) error {
	if a.digest != nil {
		return nil
	}
	f, err := os.Open(filepath.Join(baseDir, a.Path()))
	if err != nil {
		return err
	}

	defer f.Close()

	h := sha256.New()
	var n int64
	if n, err = io.Copy(h, f); err != nil {
		return err
	}
	if n != a.Size() {
		return errors.New("Writed size is not eq to file size.")
	}
	var d [sha256.Size]byte
	copy(d[:], h.Sum(nil))
	a.digest = &d
	return nil
}

func (a *Header) Marshal(w io.Writer) (err error) {
	if err = a.FileInfo.Marshal(w); err == nil {
		_, err = w.Write(a.digest[:])
	}
	return
}

func (a *Header) Unmarshal(r io.Reader) (err error) {
	a.FileInfo = &xbcommon.FileInfo{}
	if err = a.FileInfo.Unmarshal(r); err == nil {
		var b = make([]byte, sha256.Size)
		if _, err = r.Read(b); err == nil {
			var d [sha256.Size]byte
			copy(d[:], b)
			a.digest = &d
		}
	}
	return
}
