package xbreader

import (
	"os"

	"github.com/moisespsena-go/rpool"

	"github.com/moisespsena-go/io-common"
)

var (
	provider = &Provider{}
)

type Provider struct {
	Pool rpool.Pool
}

func (p *Provider) Open(archive string, start, size int64) (reader iocommon.ReadSeekCloser, err error) {
	var f *os.File
	if f, err = os.Open(archive); err != nil {
		return
	}

	if size == -1 {
		var info os.FileInfo
		if info, err = f.Stat(); err != nil {
			return
		}
		size = info.Size()
	}

	if reader, err = iocommon.NewLimitedReader(f, start, size); err != nil {
		return nil, err
	}
	return
}

type PulledProvider struct {
	pool rpool.Pool
}

func NewPulledProvider(initialCap, maxCap int) (provider *PulledProvider, err error) {
	var pool rpool.Pool
	if pool, err = rpool.NewChannelPool(initialCap, maxCap, func() (closer iocommon.ReadSeekCloser, e error) {
		if executable, err := os.Executable(); err != nil {
			return nil, err
		} else {
			return os.Open(executable)
		}
	}); err != nil {
		return
	}

	return &PulledProvider{pool: pool}, nil
}

func (p *Provider) open(start, size int64) (reader iocommon.ReadSeekCloser, err error) {
	if executable, err := os.Executable(); err != nil {
		return nil, err
	} else if f, err := os.Open(executable); err != nil {
		return nil, err
	} else if reader, err = iocommon.NewLimitedReader(f, start, size); err != nil {
		return nil, err
	}
	return
}

func (p *PulledProvider) Open(start, size int64) (reader iocommon.ReadSeekCloser, err error) {
	if reader, err = p.pool.Get(); err == nil {
		if reader, err = iocommon.NewLimitedReader(reader, start, size); err != nil {
			reader.Close()
		}
	}
	return
}

func Open(archive string, start, size int64) (reader iocommon.ReadSeekCloser, err error) {
	return provider.Open(archive, start, size)
}

func NewOpener(archive string, start, size int64) func() (reader iocommon.ReadSeekCloser, err error) {
	return func() (reader iocommon.ReadSeekCloser, err error) {
		return Open(archive, start, size)
	}
}
