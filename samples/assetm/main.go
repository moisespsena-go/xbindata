package main

import (
	"fmt"
	"io"
	"os"
	"path"

	"github.com/moisespsena-go/xbindata/sample/assets"
	"github.com/moisespsena-go/xbindata/xbcommon"
)

func main() {
	switch os.Args[1] {
	case "tree":
		assets.Assets.Root.Walk(func(dir, name string, n xbcommon.Node) error {
			fmt.Println(path.Join(dir, name))
			return nil
		})
	case "cat":
		a, ok := assets.Assets.Get(os.Args[2])
		if !ok {
			println("not found")
		} else {
			r, err := a.Reader()
			if err != nil {
				panic(err)
			}
			n, err := io.Copy(os.Stdout, r)
			if err != nil {
				panic(err)
			}
			fmt.Println("Size:", n)
		}
	case "restore":
		err := assets.Assets.Root.Restore(os.Args[2])
		if err != nil {
			panic(err)
		}
	case "env":
		println(assets.EnvName())
	}
}
