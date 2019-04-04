// Copyright © 2019 Moises P. Sena <moisespsena@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"
	"regexp"

	"github.com/gobwas/glob"
	"github.com/moisespsena-go/xbindata"

	"github.com/spf13/cobra"
)

// archiveCreateCmd represents the archiveCreate command
var archiveCreateCmd = &cobra.Command{
	Use:   "create INPUT...",
	Short: "Create binary archive with contents",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := archiveCreateConfig

		// Create input configurations.
		c.Input = make([]xbindata.InputConfig, len(args))
		for i := range c.Input {
			c.Input[i] = parseInput(args[i])
		}

		return xbindata.Translate(c)
	},
}

var archiveCreateConfig = xbindata.NewConfig()

func init() {
	archiveCmd.AddCommand(archiveCreateCmd)
	flag := archiveCreateCmd.Flags()

	c := archiveCreateConfig

	c.Embed = true
	c.HibridTagName = "xblocalfs"

	flag.BoolVar(&c.Debug, "debug", c.Debug, "Do not embed the assets, but provide the embedding API. Contents will still be loaded from disk.")
	flag.BoolVar(&c.Dev, "dev", c.Dev, "Similar to debug, but does not emit absolute paths. Expects a rootDir variable to already exist in the generated code's package.")
	flag.StringVarP(&c.Prefix, "prefix", "P", c.Prefix, "Optional path prefix to strip off asset names.")
	flag.BoolVar(&c.FileSystem, "fs", c.FileSystem, "Include asset file system.")
	flag.StringVar(&c.Package, "pkg", c.Package, "Package name to use in the generated code.")
	flag.BoolVar(&c.NoMetadata, "nometadata", c.NoMetadata, "Assets will not preserve size, mode, and modtime info.")
	flag.StringVarP(&c.Output, "api", "A", c.Output, "Name of assets api file.")
	flag.StringVarP(&c.EmbedArchive, "output", "o", c.EmbedArchive, "Optional. Append embed contents into this file.")
	flag.StringVarP(&c.ArchiveHeadersOutput, "headers-output", "O", c.ArchiveHeadersOutput, "Optional OUTPUT. Append embed contents into this file.")
	flag.BoolVarP(&c.EmbedArchiveTruncate, "truncate", "T", c.EmbedArchiveTruncate, "Optional. Truncate archive if exists.")
	flag.UintVarP(&c.Mode, "mode", "m", c.Mode, "Optional file mode override for all files.")
	flag.Int64VarP(&c.ModTime, "modtime", "t", c.ModTime, "Optional modification unix timestamp override for all files.")
	flag.BoolVarP(&c.ArchiveGziped, "compress", "C", c.ArchiveGziped, "Compress output with gzip.")
	flag.BoolVarP(&c.Hibrid, "hibrid", "H", c.Hibrid, "Create assets with in two files: `_default` and `_compiled`. "+
		"- The `*_default` was compiled when have `xblocalfs` build tag and load assets from local file system.\n"+
		"- The `*_compiled` was compiled when does not have `xblocalfs` build tag and load assets from OUTPUT.")
	flag.BoolVarP(&c.ArchiveAutoloadDisabled, "autoload-disabled", "L", c.ArchiveAutoloadDisabled, "Disable auto load assets.")
	flag.StringVar(&c.HibridTagName, "hibrid-tag", c.HibridTagName, "HTAG. Set name of hibrid tag.\n"+
		"Runs with local assets (dir of OUTPUT), not compiled: `go run -tag HTAG main.go`.\n"+
		"Runs compiled: `go run main.go`.")

	flag.Var(&StringsValue{setter: func(i int, tag string) {
		c.Tags = append(c.Tags, tag)
	}}, "tags", "Optional set of build tags to include.")

	flag.Var(&StringsValue{setter: func(i int, pattern string) {
		if regexPattern, err := regexp.Compile(pattern); err != nil {
			fmt.Println(fmt.Errorf("Invalid Regex Pattern %d %q: %v", i, pattern, err))
			os.Exit(2)
		} else {
			c.Ignore = append(c.Ignore, regexPattern)
		}
	}}, "ignore", "Regex pattern to ignore")

	flag.Var(&StringsValue{setter: func(i int, pattern string) {
		if globPattern, err := glob.Compile(pattern); err != nil {
			fmt.Println(fmt.Errorf("Invalid Glob Pattern %d %q: %v", i, pattern, err))
			os.Exit(2)
		} else {
			c.IgnoreGlob = append(c.IgnoreGlob, globPattern)
		}
	}}, "ignore-glob", "Glob pattern to ignore. See https://github.com/gobwas/glob for more details")
}
