// Code generated by xbindata. DO NOT EDIT.
// sources:
// assets/inputs/a/a.txt (1B)
// assets/inputs/b/b.txt (1B)
// assets/inputs/b/sub/d.txt (1B)

package e_multiple_inputs_mix_prefix

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/moisespsena-go/assetfs"
	fsapi "github.com/moisespsena-go/assetfs/assetfsapi"
	"github.com/moisespsena-go/io-common"
	"github.com/moisespsena-go/path-helpers"
	bc "github.com/moisespsena-go/xbindata/xbcommon"
	"github.com/moisespsena-go/xbindata/xbfs"
	"os"
	"sync"
	"time"
)

func bindataReader(data []byte, name string) (iocommon.ReadSeekCloser, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("read %q: %v", name, err)
	}
	return iocommon.NoSeeker(gz), nil
}

var pkg = path_helpers.GetCalledDir()

var (
	loaded bool
	mu     sync.Mutex
	fs     fsapi.Interface
)
var fsLoadCallbacks []func(fs fsapi.Interface)

func OnFsLoad(cb ...func(fs fsapi.Interface)) {
	fsLoadCallbacks = append(fsLoadCallbacks, cb...)
}

func callFsLoadCallbacks() {
	for _, f := range fsLoadCallbacks {
		f(fs)
	}
}

func IsLocal() bool {
	if _, err := os.Stat("assets/inputs/a"); err == nil {
		return true
	}
	return false
}

func FS() fsapi.Interface {
	Load()
	return fs
}

var DefaultFS fsapi.Interface
var LocalFS = assetfs.NewAssetFileSystem()

func LoadLocal() {
	var inputs = []string{
		"assets/inputs/a",
		"assets/inputs/b",
	}
	for _, pth := range inputs {
		if err := LocalFS.RegisterPath(pth); err != nil {
			panic(err)
		}
	}
}

func Load() {
	if loaded {
		return
	}
	mu.Lock()
	if loaded {
		mu.Unlock()
		return
	}
	defer callFsLoadCallbacks()
	defer mu.Unlock()
	defer func() { loaded = true }()
	if IsLocal() {
		LoadLocal()
		fs = LocalFS
		return
	}
	LoadDefault()
	fs = DefaultFS
}

func init() { Load() }

var _aTxt = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x72\x04\x04\x00\x00\xff\xff\x8b\x9e\xd9\xd3\x01\x00\x00\x00")

func aTxtReader() (iocommon.ReadSeekCloser, error) {
	return bindataReader(_aTxt, "a.txt")
}

var aTxt = bc.NewFile(bc.NewFileInfo("a.txt", 1, os.FileMode(436), time.Unix(1554918023, 0), time.Unix(1554918023, 0)), aTxtReader, &[32]uint8{0x55, 0x9a, 0xea, 0xd0, 0x82, 0x64, 0xd5, 0x79, 0x5d, 0x39, 0x9, 0x71, 0x8c, 0xdd, 0x5, 0xab, 0xd4, 0x95, 0x72, 0xe8, 0x4f, 0xe5, 0x55, 0x90, 0xee, 0xf3, 0x1a, 0x88, 0xa0, 0x8f, 0xdf, 0xfd})
var _bTxt = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x72\x02\x04\x00\x00\xff\xff\x31\xcf\xd0\x4a\x01\x00\x00\x00")

func bTxtReader() (iocommon.ReadSeekCloser, error) {
	return bindataReader(_bTxt, "b.txt")
}

var bTxt = bc.NewFile(bc.NewFileInfo("b.txt", 1, os.FileMode(436), time.Unix(1554918023, 0), time.Unix(1554918023, 0)), bTxtReader, &[32]uint8{0xdf, 0x7e, 0x70, 0xe5, 0x2, 0x15, 0x44, 0xf4, 0x83, 0x4b, 0xbe, 0xe6, 0x4a, 0x9e, 0x37, 0x89, 0xfe, 0xbc, 0x4b, 0xe8, 0x14, 0x70, 0xdf, 0x62, 0x9c, 0xad, 0x6d, 0xdb, 0x3, 0x32, 0xa, 0x5c})
var _subDTxt = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x4a\x01\x04\x00\x00\xff\xff\xcc\x4a\xdd\x98\x01\x00\x00\x00")

func subDTxtReader() (iocommon.ReadSeekCloser, error) {
	return bindataReader(_subDTxt, "sub/d.txt")
}

var subDTxt = bc.NewFile(bc.NewFileInfo("sub/d.txt", 1, os.FileMode(436), time.Unix(1554918964, 0), time.Unix(1554918964, 0)), subDTxtReader, &[32]uint8{0x18, 0xac, 0x3e, 0x73, 0x43, 0xf0, 0x16, 0x89, 0xc, 0x51, 0xe, 0x93, 0xf9, 0x35, 0x26, 0x11, 0x69, 0xd9, 0xe3, 0xf5, 0x65, 0x43, 0x64, 0x29, 0x83, 0xf, 0xaf, 0x9, 0x34, 0xf4, 0xf8, 0xe4})

// Assets is a table, holding each asset generator, mapped to its name.
var Assets *bc.Assets

func LoadDefault() {
	Assets = bc.NewAssets(
		aTxt,
		bTxt,
		subDTxt,
	)

	DefaultFS = xbfs.NewFileSystem(Assets)
}
