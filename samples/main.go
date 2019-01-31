package main

import (
	"github.com/moisespsena-go/xbindata"
)

func main() {
	cfg := xbindata.NewConfig()
	cfg.Output = "samples/embed/assets/headers.go"
	cfg.Input = []xbindata.InputConfig{{"samples/embed/assets_root", true}}
	cfg.Package = "assets"
	cfg.FileSystem = true
	cfg.NoCompress = true
	cfg.NoMemCopy = true
	cfg.Embed = true
	if err := xbindata.Translate(cfg); err != nil {
		panic(err)
	}

}
