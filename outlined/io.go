package outlined

import (
	"io"

	"github.com/moisespsena-go/io-common"
)

type writeCounter struct {
	io.Writer
	count int64
}

func (w *writeCounter) Write(p []byte) (n int, err error) {
	n, err = w.Writer.Write(p)
	w.count += int64(n)
	return
}

type readCounter struct {
	io.Reader
	count int64
}

func (r *readCounter) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	r.count += int64(n)
	return
}

type AssetReaderFactory func(start, size int64) func() (reader iocommon.ReadSeekCloser, err error)
