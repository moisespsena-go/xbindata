package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/moisespsena-go/xbindata/sample/assets"
)

func main() {
	// Below is an example of using our PrintMemUsage() function
	// Print our starting memory usage (should be around 0mb)
	PrintMemUsage()

	for i := 0; i < 4; i++ {
		_, _ = assets.Assets.Get("sample/data/a.txt")

		// Print our memory usage at each interval
		PrintMemUsage()
		time.Sleep(time.Second)
	}

	// Clear our memory and print usage, unless the GC has run 'Alloc' will remain the same
	PrintMemUsage()

	// Force GC to clear up, should see a memory drop
	runtime.GC()
	PrintMemUsage()
}

// PrintMemUsage outputs the current, total and OS memory being used. As well as the number
// of garage collection cycles completed.
func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("%v\t", bToMb(m.Alloc))
	fmt.Printf("%v\t", bToMb(m.TotalAlloc))
	fmt.Printf("%v\t", bToMb(m.Sys))
	fmt.Printf("%v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024
}
