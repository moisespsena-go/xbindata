package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/moisespsena-go/xbindata/sample/assets"
)

func main() {
	fs := assetfsapi.NewHttpFileSystem(assets.AssetFS)
	t := &http.Transport{}
	t.RegisterProtocol("file", http.NewFileTransport(fs))
	c := &http.Client{Transport: t}
	res, err := c.Get("file:///sample/data")
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))
}
