package xbindata

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func outlinedHeadersWrite(w io.Writer, toc []Asset, c *Config) (err error) {
	fn := "StoreFile"
	if c.OutlinedNoTruncate {
		fn = "Append"
	}
	if !c.NoCompress {
		fn += "Gz"
	}

	data := `package main

import (
	"os"
	"time"

	bc "github.com/moisespsena-go/xbindata/xbcommon"
	"github.com/moisespsena-go/xbindata/outlined"
)

func main() {
	for _, dest := range os.Args[1:] { 
		if err := headers.` + fn + `(dest); err != nil {
			panic(err)
		}
	}
}

var headers = outlined.Headers{
`
	cwd, _ := os.Getwd()
	for _, asset := range toc {
		info, _ := asset.InfoExport(c)
		rpth, err := filepath.Rel(cwd, asset.Path)
		if err != nil {
			rpth = asset.Path
		}
		data += fmt.Sprintf("\toutlined.NewHeader(bc.NewFileInfo(%q, %s), %q),\n", asset.Name, info, rpth)
	}

	data += "}\n"

	_, err = w.Write([]byte(data))
	return
}
