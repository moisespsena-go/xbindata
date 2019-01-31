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
	"io"
	"os"

	"github.com/go-errors/errors"
	"github.com/moisespsena-go/xbindata/samples/embed/assets"
	"github.com/moisespsena-go/xbindata/xbcommon"

	"github.com/spf13/cobra"
)

// fsCatCmd represents the fsCat command
var fsCatCmd = &cobra.Command{
	Use:   "cat [options] ASSET...",
	Short: "Print asset content into OUTPUT.",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		items := make([]xbcommon.Asset, len(args))
		var notFound, ok bool
		for i, pth := range args {
			if items[i], ok = assets.Assets.Get(pth); !ok {
				notFound = true
				fmt.Printf("Not found: %q\n", pth)
			}
		}

		if notFound {
			return errors.New("Assets not found.")
		}

		var out io.Writer = os.Stdout
		outFlag, _ := cmd.Flags().GetString("output")

		if outFlag != "" && outFlag != "-" {
			if f, err := os.Create(outFlag); err != nil {
				return err
			} else {
				out = f
				defer f.Close()
			}
		}

		for i, asset := range items {
			r, err := asset.Reader()
			if err != nil {
				return errors.New(fmt.Sprintf("Get reader of %q failed: %v", args[i], err))
			}
			if _, err = io.Copy(out, r); err != nil {
				return err
			}
		}
		return nil
	},
}

func init() {
	fsCmd.AddCommand(fsCatCmd)
	fsCatCmd.Flags().StringP("output", "o", "-", "Output file. Use HIFEN for use STDOUT. Default is STDOUT.")
}
