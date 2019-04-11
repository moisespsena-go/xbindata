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

	"github.com/moisespsena-go/assetfs"

	"github.com/gobwas/glob/syntax"

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
	Long: `Build outlined pkg program into one or more program executables.

PROGRAM_PATH is path of program. Accepts Glob pattern (https://github.com/gobwas/glob).

Examples:
	$ ` + prog + ` build program ./dist/linux_amd64/program
	$ ` + prog + ` build program ./dist/linux_*/program
	$ cat ./dist/linux_amd64/program | ` + prog + ` build program > ./dist/linux_amd64/program_with_assets

	To stdout:
	$ ` + prog + ` build program -
`,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		var (
			cfg     xbindata.ManyConfig
			newArgs []string
		)

	largs:
		for _, pth := range args {
			for i := 0; i < len(pth); i++ {
				if syntax.Special(pth[i]) {
					g := assetfs.NewSimpleGlobPattern(pth)
					err = filepath.Walk(filepath.FromSlash(g.Dir()), func(path string, info os.FileInfo, err error) error {
						if err != nil {
							return err
						}
						if g.Match(filepath.ToSlash(path)) {
							newArgs = append(newArgs, path)
						}
						return nil
					})
					if err != nil {
						return errwrap.Wrap(err, "Walk %q", g.Dir())
					}
					continue largs
				}
			}
			newArgs = append(newArgs, pth)
		}

		args = newArgs

		if cfgFile, err = filepath.Abs(cfgFile); err != nil {
			return
		}

		var cwd string

		var toStdout bool

		if cwd, err = os.Getwd(); err != nil {
			return
		} else if cwd != filepath.Dir(cfgFile) {
			cwd = filepath.Dir(cfgFile)
			for i, arg := range args {
				if arg == xbindata.OutputToStdout {
					toStdout = true
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

		var (
			fi    os.FileInfo
			piped bool
		)

		fi, err = os.Stdin.Stat()
		if err != nil {
			return errwrap.Wrap(err, "stat of Stdin")
		}

		if fi.Mode()&os.ModeNamedPipe != 0 {
			piped = true
			if !toStdout {
				args = append(args, xbindata.OutputToStdout)
			}
		}

		if len(args) == 0 {
			log.Println("no outputs")
			os.Exit(1)
		}

		if err = unmarshalConfig(&cfg); err != nil {
			return
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

				for _, pth := range args {
					log.Println("translating outlined " + cfg.Pkg + " into `" + pth + "`")
					c.OutputWriter = nil
					if pth == xbindata.OutputToStdout {
						if piped {
							if _, err = io.Copy(os.Stdout, os.Stdin); err != nil {
								return
							}
						}
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
