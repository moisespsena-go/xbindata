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
	"log"
	"os"
	"path/filepath"

	"github.com/moisespsena-go/path-helpers"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init [DIR...]",
	Short: "Create config template into DIR",

	RunE: func(cmd *cobra.Command, args []string) (err error) {
		if len(args) == 0 {
			args = append(args, ".")
		}

		for _, dir := range args {
			pth := filepath.Join(dir, ".xb.yaml")
			if path_helpers.IsExistingRegularFile(pth) {
				log.Println("`" + pth + "` ignore")
				continue
			}

			log.Println("`" + pth + "` initializing...")
			if err = path_helpers.MkdirAllIfNotExists(dir); err != nil {
				return
			}
			var mode os.FileMode
			if mode, err = path_helpers.ResolveFileMode(pth); err != nil {
				return
			}
			var f *os.File
			if f, err = os.OpenFile(pth, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode); err != nil {
				return
			}
			func() {
				defer f.Close()
				_, err = f.WriteString(configTemplate)
			}()
			if err != nil {
				return
			}
		}
		return
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

const configTemplate = `# #### EXAMPLE ####
# embedded:
#   - pkg: assets/embeded
#     prefix: assets/program/assets
#     fs: true
#     hybrid: true
#     inputs:
#       - path: assets/program/assets
#         recursive: true
# 
# outlined:
#   - pkg: assets/program
#     program: true
#     prefix: _
#     fs: true
#     hybrid: false
#     compressed: true
# 
#   - pkg: assets/outlined
#     prefix: assets/program/assets
#     fs: true
#     hybrid: true
#     compressed: true
#     inputs:
#       - path: assets/program/assets
#         recursive: true
# 
# ## with many inputs ##
#   - pkg: assets
#     prefix: _
#     fs: true
#     hybrid: true
#     compressed: false
#     inputs:
#       - path: inputs/a
#         prefix: _
#         recursive: true
# 
#       - path: inputs/b
#         prefix: _
#         recursive: true
# 
#       - path: another_input
#         prefix: _
#         recursive: true
`
