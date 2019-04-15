package main

import (
	"fmt"
	"github.com/moisespsena-go/xbindata/samples/mix/assets/outlineds/embed_executable"
)

func main() {
	fmt.Println(embed_executable.FS().MustAsset("a.txt").MustDataS())
}
