package xbcommon

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"time"
)

type FileInfo struct {
	path       string
	name       string
	size       int64
	mode       os.FileMode
	modTime    time.Time
	changeTime time.Time
}

func NewFileInfo(pth string, size int64, mode os.FileMode, modTime, changeTime time.Time) *FileInfo {
	return &FileInfo{path: pth, name: path.Base(pth), size: size, mode: mode, modTime: modTime, changeTime: changeTime}
}

func (fi *FileInfo) Path() string {
	return fi.path
}

func (fi *FileInfo) Name() string {
	return fi.name
}
func (fi *FileInfo) Size() int64 {
	return fi.size
}
func (fi *FileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi *FileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi *FileInfo) ChangeTime() time.Time {
	return fi.changeTime
}
func (fi FileInfo) IsDir() bool {
	return false
}
func (fi FileInfo) Sys() interface{} {
	return nil
}

func (fi *FileInfo) Marshal(w io.Writer) (err error) {
	if err = binary.Write(w, binary.BigEndian, fi.size); err != nil {
		return fmt.Errorf("Write Size: %v", err)
	}
	if err = binary.Write(w, binary.BigEndian, uint32(fi.mode)); err != nil {
		return fmt.Errorf("Write Mode: %v", err)
	}
	if err = binary.Write(w, binary.BigEndian, uint64(fi.modTime.UnixNano())); err != nil {
		return fmt.Errorf("Write ModTime: %v", err)
	}
	if err = binary.Write(w, binary.BigEndian, uint64(fi.changeTime.UnixNano())); err != nil {
		return fmt.Errorf("Write ChangeTime: %v", err)
	}
	pth := filepath.ToSlash(fi.path)
	if err = binary.Write(w, binary.BigEndian, uint32(len(pth))); err != nil {
		return fmt.Errorf("Write Path Size: %v", err)
	}
	if _, err = w.Write([]byte(pth)); err != nil {
		return fmt.Errorf("Write Path: %v", err)
	}
	return
}

func (fi *FileInfo) Unmarshal(r io.Reader) (err error) {
	var (
		i   uint32
		i64 uint64
	)
	if err = binary.Read(r, binary.BigEndian, &fi.size); err != nil {
		return fmt.Errorf("Read Size: %v", err)
	}
	if err = binary.Read(r, binary.BigEndian, &i); err != nil {
		return fmt.Errorf("Read Mode: %v", err)
	}
	fi.mode = os.FileMode(i)

	if err = binary.Read(r, binary.BigEndian, &i64); err != nil {
		return fmt.Errorf("Read ModTime: %v", err)
	}
	fi.modTime = time.Unix(0, int64(i64))
	if err = binary.Read(r, binary.BigEndian, &i64); err != nil {
		return fmt.Errorf("Read ChangeTime: %v", err)
	}
	fi.changeTime = time.Unix(0, int64(i64))
	if err = binary.Read(r, binary.BigEndian, &i); err != nil {
		return fmt.Errorf("Read Path Size: %v", err)
	}
	var (
		b = make([]byte, i)
		n int
	)
	if n, err = r.Read(b); err != nil {
		return fmt.Errorf("Read Path: %v", err)
	} else if n != int(i) {
		return fmt.Errorf("Read Path Count %d != %d ", n, int(i))
	}
	fi.path = string(b)
	fi.name = path.Base(fi.path)
	return
}
