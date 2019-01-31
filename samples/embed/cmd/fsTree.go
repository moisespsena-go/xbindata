// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
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
	"strings"

	"github.com/moisespsena-go/bits2str"
	"github.com/moisespsena-go/xbindata/samples/embed/assets"
	"github.com/moisespsena-go/xbindata/xbcommon"
	"github.com/phayes/permbits"
	"github.com/spf13/cobra"
)

// fsTreeCmd represents the fsTree command
var fsTreeCmd = &cobra.Command{
	Use:   "tree",
	Short: "Print assets tree",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(".")
		_ = assets.Assets.Root.Walk(func(dir, name string, n xbcommon.Node, d interface{}) (r interface{}, err error) {
			var line = name
			if n.IsDir() {
				line += "/"
			} else {
				line += fmt.Sprint("\t", bits2str.Bits(n.Size())*bits2str.Byte, " ", permbits.FileMode(n.Mode()).String(), " ", n.ModTime())
			}
			var pad = ""
			if dir != "." {
				dir += "/"
			}
			if name != "." {
				pad = strings.Repeat("│  ", n.Depth()-1)
				if n.IsLast() {
					pad += "└──"
				} else {
					pad += "├──"
				}
			}
			fmt.Println(pad + " " + line)
			return
		})
	},
}

func init() {
	fsCmd.AddCommand(fsTreeCmd)
}
