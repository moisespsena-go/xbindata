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
	"github.com/spf13/cobra"
)

var archivePath string

// archiveLsCmd represents the archiveList command
var archiveLsCmd = &cobra.Command{
	Use:   "ls ASSET_NAME...",
	Short: "List asset or assets into dir",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		return
	},
}

func init() {
	archiveCmd.AddCommand(archiveLsCmd)

	pflag := archiveLsCmd.PersistentFlags()
	pflag.StringVarP(&archivePath, "archive", "A", "assets.bin", "The archive path.")

	flag := archiveLsCmd.Flags()
	flag.BoolP("long", "l", false, "Print as long list.")
}
