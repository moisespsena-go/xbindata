package main

import (
	"fmt"

	"github.com/moisespsena-go/xbindata/samples/mix/assets/embed_go_source/e_multiple_inputs_mix_prefix"
)

func main() {
	fmt.Println(e_multiple_inputs_mix_prefix.FS().MustAsset("a.txt").MustDataS())
}
