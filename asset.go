// This work is subject to the CC0 1.0 Universal (CC0 1.0) Public Domain Dedication
// license. Its contents can be found at:
// http://creativecommons.org/publicdomain/zero/1.0/

package xbindata

import (
	"crypto/sha256"
	"fmt"
	"os"
	"time"

	"github.com/moisespsena-go/xbindata/digest"

	"github.com/djherbis/times"
)

// Asset holds information about a single asset to be processed.
type Asset struct {
	Path string // Full file path.
	Name string // Key used in TOC -- name by which asset is referenced.
	Func string // Function name for the procedure returning the asset contents.
	Size int64

	Prefix string

	info  os.FileInfo
	ctime time.Time

	digest *[sha256.Size]byte
}

func (a *Asset) Info() (info os.FileInfo, err error) {
	if a.info != nil {
		return a.info, nil
	}

	if a.info, err = os.Stat(a.Path); err != nil {
		return
	}
	t := times.Get(a.info)
	a.ctime = t.ChangeTime()
	return a.info, nil
}

func (a *Asset) InfoExport(c *Config) (string, error) {
	fi, err := a.Info()
	if err != nil {
		return "", err
	}
	mode := uint(fi.Mode())
	modTime := fi.ModTime().Unix()
	changeTime := a.ctime.Unix()
	size := fi.Size()
	if c.NoMetadata {
		mode = 0
		modTime = 0
		size = 0
		changeTime = 0
	}
	if c.Mode > 0 {
		mode = uint(os.ModePerm) & c.Mode
	}
	if c.ChangeTime > 0 {
		changeTime = c.ChangeTime
	}
	return fmt.Sprintf("%d, os.FileMode(%d), time.Unix(%d, 0), time.Unix(%d, 0)", size, mode, modTime, changeTime), nil
}

func (a *Asset) SourceCode(c *Config, start int64) (code string, err error) {
	digest, err := a.Digest()
	if err != nil {
		return "", err
	}
	var readerFunc string
	if c.Outlined {
		readerFunc = fmt.Sprintf("newOpener(%d, %d)", start, a.Size)
	} else {
		readerFunc = fmt.Sprintf("%sReader", a.Func)
	}
	info, err := a.InfoExport(c)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("bc.NewFile(bc.NewFileInfo(%q, %s), %s,  func()[sha256.Size]byte{return %#v})",
		a.Name, info, readerFunc, *digest), nil
}

func (a *Asset) Digest() (dig *[sha256.Size]byte, err error) {
	if a.digest != nil {
		return a.digest, nil
	}

	if dig, err = digest.Digest(a.Path); err != nil {
		return
	}
	a.digest = dig
	return
}
