package archive

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/moisespsena-go/xbindata/xbreader"

	"github.com/moisespsena-go/io-common"

	"github.com/moisespsena-go/xbindata/xbcommon"
)

type Archive struct {
	Headers     Headers
	HeadersSize int64
	Path        string
	Len         int
}

func New() *Archive {
	return &Archive{}
}

func OpenFile(pth string) (archive *Archive, err error) {
	archive = &Archive{Path: pth}
	if err = archive.ReadFile(pth); err != nil {
		return nil, err
	}
	return
}

func (archive *Archive) Read(r io.Reader) (err error) {
	rc := &readCounter{Reader: r}
	var i uint32

	if err = binary.Read(rc, binary.BigEndian, &i); err != nil {
		err = fmt.Errorf("Read headers count failed: %v", err)
		return
	}

	archive.Len = int(i)

	if archive.Len == 0 {
		return
	}

	headers := make(Headers, int(i), int(i))

	for i := 0; i < archive.Len; i++ {
		headers[i] = &Header{}
		if err = headers[i].Unmarshal(rc); err != nil {
			err = fmt.Errorf("Read headers %d failed: %v", i, err)
			return
		}
	}

	archive.HeadersSize = rc.count
	archive.Headers = headers
	return
}

func (archive *Archive) ReadFile(pth string) (err error) {
	var f *os.File
	if f, err = os.Open(pth); err != nil {
		return
	}
	archive.Path = pth
	defer f.Close()
	return archive.Read(f)
}

func (archive *Archive) ReaderFactory(start, size int64) func() (reader iocommon.ReadSeekCloser, err error) {
	return func() (reader iocommon.ReadSeekCloser, err error) {
		return xbreader.Open(archive.Path, archive.HeadersSize+start, size)
	}
}

func (archive *Archive) EachAsset(cb func(i int, asset xbcommon.Asset), readerFactory ...AssetReaderFactory) {
	var rf AssetReaderFactory = archive.ReaderFactory
	if len(readerFactory) > 0 && readerFactory[0] != nil {
		var headersSize = archive.HeadersSize
		rf = func(start, size int64) func() (reader iocommon.ReadSeekCloser, err error) {
			return readerFactory[0](headersSize+start, size)
		}
	}
	archive.Headers.EachAssets(rf, cb)
}

func (archive *Archive) Assets(readerFactory ...AssetReaderFactory) (assets []xbcommon.Asset) {
	assets = make([]xbcommon.Asset, archive.Len, archive.Len)
	archive.EachAsset(func(i int, asset xbcommon.Asset) {
		assets[i] = asset
	}, readerFactory...)
	return
}

func (archive *Archive) AssetsMap(readerFactory ...AssetReaderFactory) (assets map[string]xbcommon.Asset) {
	assets = map[string]xbcommon.Asset{}
	archive.EachAsset(func(i int, asset xbcommon.Asset) {
		assets[asset.Path()] = asset
	}, readerFactory...)
	return
}
