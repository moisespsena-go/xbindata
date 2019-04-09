package xbindata

import (
	"fmt"
	"io"
	"path/filepath"
	"strconv"
)

func outlinedHeadersWrite(w io.Writer, toc []Asset, c *Config) (err error) {
	var prefix string
	if c.Prefix != "" {
		prefix, _ = filepath.Abs(c.Prefix)
	}

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
		if err := headers.` + fn + `(dest, baseDir); err != nil {
			panic(err)
		}
	}
}

const baseDir = ` + strconv.Quote(prefix) + `
var headers = outlined.Headers{
`
	for _, asset := range toc {
		info, _ := asset.InfoExport(c)
		data += fmt.Sprintf("\toutlined.NewHeader(bc.NewFileInfo(%q, %s)),\n", asset.Name, info)
	}

	data += "}\n"

	_, err = w.Write([]byte(data))
	return
}
