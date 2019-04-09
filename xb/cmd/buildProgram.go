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

	"github.com/moisespsena-go/xbindata"

	"github.com/spf13/cobra"
)

// buildProgramCmd represents the buildProgram command
var buildProgramCmd = &cobra.Command{
	Use:   "program PROGRAM_PATH...",
	Args:  cobra.MinimumNArgs(1),
	Short: "Build outlined pkg program into one or more program executables",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		var cfg xbindata.ManyConfig
		if err = unmarshalConfig(&cfg); err != nil {
			return
		}

		if err = cfg.Validate(); err != nil {
			return
		}

		for _, cfg := range cfg.Outlined {
			if cfg.Program {
				c, err := cfg.Config()
				if err != nil {
					return err
				}
				c.OutlinedNoTruncate = true
				c.NoCompress = true
				c.OulinedSkipApi = true
				for _, pth := range args {
					log.Println("translating outlined " + cfg.Pkg + " into `" + pth + "`")
					c.Output = pth
					if err = xbindata.Translate(c); err != nil {
						return err
					}
				}
				return nil
			}
		}
		log.Println("No outlined found")
		return
	},
}

func init() {
	buildCmd.AddCommand(buildProgramCmd)
}
