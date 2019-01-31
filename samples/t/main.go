package main

import (
	"fmt"
	"strconv"
)

var i int
var keys []*string

func k() *string {
	k := strconv.Itoa(i)
	keys = append(keys, &k)
	i++
	return &k
}

func main() {
	m := map[*string]int{
		k(): 1,
		k(): 2,
		k(): 3,
	}

	fmt.Println(keys)
	fmt.Println(m)
}
