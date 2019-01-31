package xbindata_test

import (
	"go/format"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	bindata "github.com/moisespsena-go/xbindata"
)

func dirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}

func BenchmarkBindata(b *testing.B) {
	size, err := dirSize("testdata/benchmark")
	if err != nil {
		b.Fatal(err)
	}
	b.SetBytes(size)
	cfg := &bindata.Config{
		Package: "assets",
		Input: []bindata.InputConfig{
			{Path: "testdata/benchmark", Recursive: true},
		},
		Output:     "testdata/assets/xbindata.go",
		NoMemCopy:  false,
		NoCompress: true,
		Debug:      false,
		NoMetadata: true,
		Mode:       0x0, ModTime: 0,
		Ignore: nil,
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := bindata.Translate(cfg); err != nil {
			b.Fatal(err)
		}
	}
}

var formatSink []byte

func BenchmarkFormatSource(b *testing.B) {
	// https://github.com/golang/go/issues/26528
	// unformatted.out is any large xbindata source file.
	data, err := ioutil.ReadFile("testdata/unformatted.out")
	if os.IsNotExist(err) {
		b.Skip("source file does not exist")
		return
	}
	if err != nil {
		b.Fatal(err)
	}
	b.SetBytes(int64(len(data)))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		formatSink, err = format.Source(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}
