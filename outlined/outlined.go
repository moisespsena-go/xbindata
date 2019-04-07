package outlined

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/moisespsena-go/path-helpers"
	"github.com/op/go-logging"

	"github.com/moisespsena-go/xbindata/xbreader"

	"github.com/moisespsena-go/io-common"

	"errors"

	"github.com/moisespsena-go/xbindata/xbcommon"
)

var log = logging.MustGetLogger(path_helpers.GetCalledDir())

type Outlined struct {
	Headers     Headers
	HeadersSize int64
	Path        string
	Len         int
	Hash        [sha256.Size]byte
	BuildDate   time.Time
}

func New() *Outlined {
	return &Outlined{}
}

func OpenFile(pth string) (outlined *Outlined, err error) {
	outlined = &Outlined{Path: pth}
	if err = outlined.ReadFile(pth); err != nil {
		return nil, err
	}
	return
}

func (outlined *Outlined) readHeaders(r io.Reader) (err error) {
	hash := make([]byte, sha256.Size)
	if _, err = r.Read(hash); err != nil {
		err = fmt.Errorf("read content hash failed: %v", err)
	}

	if err = readNL(r); err != nil {
		return
	}

	copy(outlined.Hash[:], hash)

	var (
		i64 uint64
		i   uint32
	)
	if err = binary.Read(r, binaryDir, &i64); err != nil {
		err = fmt.Errorf("read build time failed: %v", err)
		return
	}
	if err = readNL(r); err != nil {
		return
	}

	t := time.Unix(int64(i64), 0)
	outlined.BuildDate = t

	if err = binary.Read(r, binaryDir, &i); err != nil {
		err = fmt.Errorf("Read headers count failed: %v", err)
		return
	}

	if err = readNL(r); err != nil {
		return
	}

	outlined.Len = int(i)

	if outlined.Len == 0 {
		return
	}

	var headers = make(Headers, outlined.Len, outlined.Len)

	for i := 0; i < outlined.Len; i++ {
		h := &Header{}
		if err = h.Unmarshal(r); err != nil {
			err = fmt.Errorf("Read headers %d failed: %v", i, err)
			return
		}

		if err = readNL(r); err != nil {
			return
		}

		headers[i] = h
	}
	outlined.Headers = headers
	return nil
}

func (outlined *Outlined) Read(r io.Reader) (err error) {
	rc := &readCounter{Reader: r}
	if err = outlined.readHeaders(rc); err != nil {
		return
	}
	outlined.HeadersSize = rc.count
	return
}

func (outlined *Outlined) Uncompress(pth string) (n int64, err error) {
	s, err := os.Open(pth)
	if err != nil {
		return 0, err
	}
	defer s.Close()
	dst := strings.TrimSuffix(pth, ".gz")
	d, err := os.Create(dst)
	if err != nil {
		return 0, fmt.Errorf("Create %q failed: %v", pth, err)
	}
	defer d.Close()
	gr, err := gzip.NewReader(s)
	defer gr.Close()
	if err != nil {
		return 0, fmt.Errorf("Create Gzip Reader failed: %v", err)
	}
	n, err = io.Copy(d, gr)
	if err != nil {
		return 0, fmt.Errorf("Copy failed: %v", err)
	}
	return n, nil
}

func (outlined *Outlined) ReadFile(pth string) (err error) {
	var f *os.File
	if !strings.HasSuffix(pth, ".gz") {
		pth += ".gz"
	}

	if _, err := os.Stat(pth); err == nil {
		dpth := strings.TrimSuffix(pth, ".gz")
		log.Infof("Uncompressing %q to %q", pth, dpth)
		n, err := outlined.Uncompress(pth)
		if err != nil {
			return fmt.Errorf("Uncompress %q to %q failed: %v", pth, dpth, err)
		}
		log.Infof("Uncompressing done. Destination size is %s", humanize.Bytes(uint64(n)))
		if os.Getenv("XBINDATA_ARCHIVE_NOT_REMOVE_GZ") == "" {
			os.Remove(pth)
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	pth = strings.TrimSuffix(pth, ".gz")

	if f, err = os.Open(pth); err != nil {
		return
	}
	outlined.Path = pth
	defer f.Close()
	return outlined.Read(f)
}

func (outlined *Outlined) ReaderFactory(start, size int64) func() (reader iocommon.ReadSeekCloser, err error) {
	return func() (reader iocommon.ReadSeekCloser, err error) {
		return xbreader.Open(outlined.Path, outlined.HeadersSize+start, size)
	}
}

func (outlined *Outlined) EachAsset(cb func(i int, asset xbcommon.Asset), readerFactory ...AssetReaderFactory) {
	var rf AssetReaderFactory = outlined.ReaderFactory
	if len(readerFactory) > 0 && readerFactory[0] != nil {
		var headersSize = outlined.HeadersSize
		rf = func(start, size int64) func() (reader iocommon.ReadSeekCloser, err error) {
			return readerFactory[0](headersSize+start, size)
		}
	}
	outlined.Headers.EachAssets(rf, cb)
}

func (outlined *Outlined) Assets(readerFactory ...AssetReaderFactory) (assets []xbcommon.Asset) {
	assets = make([]xbcommon.Asset, outlined.Len, outlined.Len)
	outlined.EachAsset(func(i int, asset xbcommon.Asset) {
		assets[i] = asset
	}, readerFactory...)
	return
}

func (outlined *Outlined) AssetsMap(readerFactory ...AssetReaderFactory) (assets map[string]xbcommon.Asset) {
	assets = map[string]xbcommon.Asset{}
	outlined.EachAsset(func(i int, asset xbcommon.Asset) {
		assets[asset.Path()] = asset
	}, readerFactory...)
	return
}

func readNL(r io.Reader) (err error) {
	var b = make([]byte, 1)
	if _, err = r.Read(b); err != nil {
		err = fmt.Errorf("read NL failed: %v", err)
		return
	} else if string(b) != "\n" {
		return errors.New("NL expected")
	}
	return
}
