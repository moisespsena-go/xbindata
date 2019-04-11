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
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/moisespsena-go/error-wrap"

	"github.com/moisespsena-go/xbindata"

	"github.com/spf13/cobra"
)

// buildProgramCmd represents the buildProgram command
var buildProgramCmd = &cobra.Command{
	Use: "program PROGRAM_PATH...",
	Args: func(cmd *cobra.Command, args []string) error {
		fi, err := os.Stdin.Stat()
		if err != nil {
			return errwrap.Wrap(err, "stat of Stdin")
		}

		if fi.Mode()&os.ModeNamedPipe != 0 {
			return nil
		}
		return cobra.MinimumNArgs(1)(cmd, args)
	},
	Short: "Build outlined pkg program into one or more program executables",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		var cfg xbindata.ManyConfig
		if err = unmarshalConfig(&cfg); err != nil {
			return
		}

		if cfgFile, err = filepath.Abs(cfgFile); err != nil {
			return
		}

		var cwd string

		if cwd, err = os.Getwd(); err != nil {
			return
		} else if cwd != filepath.Dir(cfgFile) {
			cwd = filepath.Dir(cfgFile)
			for i, arg := range args {
				if arg == xbindata.OutputToStdout {
					continue
				}
				if !filepath.IsAbs(arg) {
					if arg, err = filepath.Abs(arg); err != nil {
						return errwrap.Wrap(err, "filepath.Abs to %q", arg)
					}
					if args[i], err = filepath.Rel(cwd, arg); err != nil {
						return errwrap.Wrap(err, "filepath.Rel to %q", arg)
					}
				}
			}
			if err = os.Chdir(cwd); err != nil {
				return
			}
		}

		if err = cfg.Validate(); err != nil {
			return
		}

		for _, cfg := range cfg.Outlined {
			if cfg.Program {
				var c *xbindata.Config
				c, err = cfg.Config()
				if err != nil {
					return err
				}
				c.OutlinedNoTruncate = true
				c.NoCompress = true
				c.OulinedSkipApi = true

				var fi os.FileInfo

				fi, err = os.Stdin.Stat()
				if err != nil {
					return errwrap.Wrap(err, "stat of Stdin")
				}

				if fi.Mode()&os.ModeNamedPipe != 0 {
					if _, err = io.Copy(os.Stdout, os.Stdin); err != nil {
						return
					}
					if len(args) == 0 {
						args = append(args, xbindata.OutputToStdout)
					}
				}

				for _, pth := range args {
					log.Println("translating outlined " + cfg.Pkg + " into `" + pth + "`")
					c.OutputWriter = nil
					if pth == xbindata.OutputToStdout {
						c.OutputWriter = os.Stdout
					}
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
