package outlined

import (
	"crypto/sha256"
	"errors"
	"io"
	"os"

	"github.com/moisespsena-go/xbindata/xbcommon"
)

type Header struct {
	*xbcommon.FileInfo
	digest     *[sha256.Size]byte
	storeSize  int64
	compressed bool
	SysPath    string
}

func (a *Header) Compressed(storeSize int64) *Header {
	a.compressed = true
	a.storeSize = storeSize
	return a
}

func (a *Header) IsCompressed() bool {
	return a.compressed
}

func NewHeader(fileInfo *xbcommon.FileInfo, sysPath string) *Header {
	return &Header{FileInfo: fileInfo, SysPath: sysPath}
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

func (a *Header) LoadDigest(allW ...io.Writer) error {
	if a.digest != nil {
		return nil
	}
	f, err := os.Open(a.SysPath)
	if err != nil {
		return err
	}

	defer f.Close()

	h := sha256.New()
	var n int64
	r := teereader{f, allW}
	if n, err = io.Copy(h, r); err != nil {
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

type teereader struct {
	r   io.Reader
	dst []io.Writer
}

func (t teereader) Read(p []byte) (n int, err error) {
	n, err = t.r.Read(p)
	if n > 0 {
		for _, w := range t.dst {
			if n, err := w.Write(p[:n]); err != nil {
				return n, err
			}
		}
	}
	return
}
