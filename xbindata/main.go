// This work is subject to the CC0 1.0 Universal (CC0 1.0) Public Domain Dedication
// license. Its contents can be found at:
// http://creativecommons.org/publicdomain/zero/1.0/

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gobwas/glob"

	"github.com/moisespsena-go/xbindata"
)

func main() {
	cfg := parseArgs()
	err := xbindata.Translate(cfg)

	if err != nil {
		fmt.Fprintf(os.Stderr, "xbindata: %v\n", err)
		os.Exit(1)
	}
}

// parseArgs create s a new, filled configuration instance
// by reading and parsing command line options.
//
// This function exits the program with an error, if
// any of the command line options are incorrect.
func parseArgs() *xbindata.Config {
	var longVersion bool
	var version bool

	c := xbindata.NewConfig()

	flag.Usage = func() {
		fmt.Printf("Usage: %s [options] <input directories>\n\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.BoolVar(&c.Debug, "debug", c.Debug, "Do not embed the assets, but provide the embedding API. Contents will still be loaded from disk.")
	flag.BoolVar(&c.Dev, "dev", c.Dev, "Similar to debug, but does not emit absolute paths. Expects a rootDir variable to already exist in the generated code's package.")
	flag.StringVar(&c.Tags, "tags", c.Tags, "Optional set of build tags to include.")
	flag.StringVar(&c.Prefix, "prefix", c.Prefix, "Optional path prefix to strip off asset names.")
	flag.BoolVar(&c.FileSystem, "fs", c.FileSystem, "Include asset file system.")
	flag.StringVar(&c.Package, "pkg", c.Package, "Package name to use in the generated code.")
	flag.BoolVar(&c.NoMemCopy, "nomemcopy", c.NoMemCopy, "Use a .rodata hack to get rid of unnecessary memcopies. Refer to the documentation to see what implications this carries.")
	flag.BoolVar(&c.NoCompress, "nocompress", c.NoCompress, "Assets will *not* be GZIP compressed when this flag is specified.")
	flag.BoolVar(&c.NoMetadata, "nometadata", c.NoMetadata, "Assets will not preserve size, mode, and modtime info.")
	flag.BoolVar(&c.Embed, "embed", c.NoCompress, "Assets contents will *not* compiled, but generate OUTPUT_DIR/data_store/main.go "+
		"program for store content data. Skips nomemcopy and nocompress flags. After compile your program, executes OUTPUT_DIR/data_store/main.go for "+
		"append asset contents into your executable. Example: `cd my_project; xbindata -embed -pkg assets -o assets/data.go assets/data; go build -o bin/my_project; go run assets/data_store/main.go bin/my_project`")
	flag.StringVar(&c.EmbedArchive, "embed-archive", c.EmbedArchive, "Optional. Append embed contents into this file.")
	flag.BoolVar(&c.EmbedArchiveTruncate, "embed-archive-truncate", c.EmbedArchiveTruncate, "Optional. Truncate embed asset archive if exists.")
	flag.UintVar(&c.Mode, "mode", c.Mode, "Optional file mode override for all files.")
	flag.Int64Var(&c.ModTime, "modtime", c.ModTime, "Optional modification unix timestamp override for all files.")
	flag.StringVar(&c.Output, "o", c.Output, "Optional name of the output file to be generated.")
	flag.BoolVar(&longVersion, "version", false, "Displays verbose version information.")
	flag.BoolVar(&version, "v", false, "Displays version information.")

	ignore := make([]string, 0)
	flag.Var((*AppendSliceValue)(&ignore), "ignore", "Regex pattern to ignore")

	ignoreGlob := make([]string, 0)
	flag.Var((*AppendSliceValue)(&ignoreGlob), "ignore", "Glob pattern to ignore. See https://github.com/gobwas/glob for more details")

	flag.Parse()
	var (
		err          error
		regexPattern *regexp.Regexp
		globPattern  glob.Glob
	)
	regexPatterns := make([]*regexp.Regexp, 0)
	for i, pattern := range ignore {
		if regexPattern, err = regexp.Compile(pattern); err != nil {
			fmt.Println(fmt.Errorf("Invalid Regex Pattern %d %q: %v", i, pattern, err))
			os.Exit(2)
		}
		regexPatterns = append(regexPatterns, regexPattern)
	}
	c.Ignore = regexPatterns

	globPatterns := make([]glob.Glob, 0)
	for i, pattern := range ignoreGlob {
		if globPattern, err = glob.Compile(pattern); err != nil {
			fmt.Println(fmt.Errorf("Invalid Glob Pattern %d %q: %v", i, pattern, err))
			os.Exit(2)
		}
		globPatterns = append(globPatterns, globPattern)
	}
	c.IgnoreGlob = globPatterns

	if version {
		fmt.Printf("%s\n", Version())
		os.Exit(0)
	}
	if longVersion {
		fmt.Printf("%s\n", LongVersion())
		os.Exit(0)
	}

	// Make sure we have input paths.
	if flag.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "Missing <input dir>\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// Create input configurations.
	c.Input = make([]xbindata.InputConfig, flag.NArg())
	for i := range c.Input {
		c.Input[i] = parseInput(flag.Arg(i))
	}

	return c
}

// parseRecursive determines whether the given path has a recursive indicator and
// returns a new path with the recursive indicator chopped off if it does.
//
//  ex:
//      /path/to/foo/...    -> (/path/to/foo, true)
//      /path/to/bar        -> (/path/to/bar, false)
func parseInput(path string) xbindata.InputConfig {
	if strings.HasSuffix(path, "/...") {
		return xbindata.InputConfig{
			Path:      filepath.Clean(path[:len(path)-4]),
			Recursive: true,
		}
	} else {
		return xbindata.InputConfig{
			Path:      filepath.Clean(path),
			Recursive: false,
		}
	}

}
