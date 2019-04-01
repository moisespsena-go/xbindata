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
	"os"

	"github.com/moisespsena-go/xbindata/samples/embed/assets"
	"github.com/moisespsena-go/path-helpers"
	"github.com/spf13/cobra"
)

// fsRestoreCmd represents the fsRestore command
var fsRestoreCmd = &cobra.Command{
	Use:   "restore DEST_DIR [ASSET_DIR]...",
	Short: "Restore asset directories into DEST_DIR. if ASSET_DIR is empty, restore asset three.",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var dest = args[0]

		if err := path_helpers.MkdirAllIfNotExists(dest); err != nil {
			return err
		}

		args = args[1:]
		if len(args) > 0 {
			for _, pth := range args {
				dir, err := assets.Assets.Root.GetDir(pth)
				if err != nil {
					return err
				}
				if err = dir.Restore(dest); err != nil {
					return err
				}
			}
			return nil
		} else {
			return assets.Assets.Root.Restore(os.Args[2])
		}
	},
}

func init() {
	fsCmd.AddCommand(fsRestoreCmd)
}
