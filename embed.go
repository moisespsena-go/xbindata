package xbindata

import (
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
)

func archiveHeadersWrite(w io.Writer, toc []Asset, c *Config) (err error) {
	var imports string
	var archive = c.EmbedArchive
	if len(toc) > 0 {
		imports = "\"os\""
	}
	if archive == "" {
		archive = "os.Args[1]"
	} else {
		archive = fmt.Sprintf("%q", archive)
	}
	var prefix string
	if c.Prefix != "" {
		prefix, _ = filepath.Abs(c.Prefix)
	}

	var gz string
	if c.ArchiveGziped {
		gz = "Gz"
	}

	_, err = w.Write([]byte(`package main

import (
	` + imports + `

    "errors"
	"time"

	bc "github.com/moisespsena-go/xbindata/xbcommon"
	"github.com/moisespsena-go/xbindata/archive"
)

func main() {
	dest := "assets.xb"
    if len(os.Args) > 1 {
		dest = os.Args[1]
	}
	if dest == "" {
		panic(errors.New("destination file is empty"))
	}

	if err := headers.StoreFile` + gz + `(dest, baseDir); err != nil {
		panic(err)
	}
}

const baseDir = ` + strconv.Quote(prefix) + `
var headers = archive.Headers{
`))
	for i := range toc {
		info, _ := toc[i].InfoExport(c)
		pth, _ := filepath.Abs(toc[i].Path)
		if prefix != "" {
			pth = strings.TrimPrefix(strings.Replace(pth, prefix, "", -1), string(filepath.Separator))
		}
		if _, err = fmt.Fprintf(w, "\tarchive.NewHeader(bc.NewFileInfo(%q, %s)),\n", pth, info); err != nil {
			return err
		}
	}
	_, err = w.Write([]byte("}\n"))
	return
}
