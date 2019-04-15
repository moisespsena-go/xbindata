// Code generated by xbindata. DO NOT EDIT.
// sources:
// assets/inputs/a/a.txt (1B)

package e_simple_input

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
	"path/filepath"
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
	}
	localDir := filepath.Join("", filepath.FromSlash(pkg))
	if _, err := os.Stat(localDir); err == nil {
		for i, pth := range inputs {
			inputs[i] = filepath.Join(localDir, pth)
		}
	} else if !os.IsNotExist(err) {
		panic(err)
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

// Embed is a table, holding each asset generator, mapped to its name.
var Embed *bc.Assets

func LoadDefault() {
	Embed = bc.NewAssets(
		aTxt,
	)

	DefaultFS = xbfs.NewFileSystem(Embed)
}
