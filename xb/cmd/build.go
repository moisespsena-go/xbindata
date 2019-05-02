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
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strconv"

	"github.com/mitchellh/mapstructure"

	"github.com/moisespsena-go/xbindata"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func unmarshalConfig(dest interface{}) error {
	return viper.Unmarshal(dest, func(config *mapstructure.DecoderConfig) {
		oldHook := config.DecodeHook
		config.DecodeHook = mapstructure.ComposeDecodeHookFunc(oldHook, func(from reflect.Type, to reflect.Type, v interface{}) (interface{}, error) {
			if to.Kind() == reflect.Struct {
				unmarshalerType := reflect.TypeOf((*xbindata.MapUnmarshaler)(nil)).Elem()
				if reflect.PtrTo(to).Implements(unmarshalerType) {
					unmh := reflect.New(to).Interface().(xbindata.MapUnmarshaler)
					if err := unmh.UnmarshalMap(v); err != nil {
						return nil, err
					}
					return unmh, nil
				}
			}
			return v, nil
		})
	})
}

// buildCmd represents the build command
var (
	cfgFile string

	buildCmd = &cobra.Command{
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			initConfig()
		},
		Use:   "build [PKG...]",
		Short: "build all or specified PKG from config file",
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
				if err = os.Chdir(filepath.Dir(cfgFile)); err != nil {
					return
				}
			}

			if err = cfg.Validate(); err != nil {
				return
			}

			if len(args) > 0 {
				var (
					embedded []xbindata.ManyConfigEmbedded
					outline  []xbindata.ManyConfigOutlined
					accepts  = func(pkg string) bool {
						for _, arg := range args {
							if arg == pkg {
								return true
							}
						}
						return false
					}
				)

				for _, cfg := range cfg.Outlined {
					if accepts(cfg.Pkg) {
						outline = append(outline, cfg)
					}
				}
				for _, cfg := range cfg.Embedded {
					if accepts(cfg.Pkg) {
						embedded = append(embedded, cfg)
					}
				}
				cfg.Outlined, cfg.Embedded = outline, embedded
			}

			ctx := context.Background()

			for i, cfg := range cfg.Outlined {
				log.Println("==== cfg config #"+strconv.Itoa(i)+":", cfg.Pkg, " ====")
				var (
					c     *xbindata.Config
					count int
				)
				if c, err = cfg.Config(ctx); err != nil {
					return fmt.Errorf("cfg #%d [%s]: create config failed: %v", i, cfg.Pkg, err)
				}
				if count, err = xbindata.Translate(c); err != nil {
					return fmt.Errorf("cfg #%d [%s]: translate failed: %v", i, cfg.Pkg, err)
				}
				log.Printf("done with %d assets.\n", count)
			}

			for i, cfg := range cfg.Embedded {
				log.Println("==== embeded config #"+strconv.Itoa(i)+":", cfg.Pkg, " ====")
				var (
					c     *xbindata.Config
					count int
				)
				if c, err = cfg.Config(ctx); err != nil {
					return fmt.Errorf("cfg #%d [%s]: create config failed: %v", i, cfg.Pkg, err)
				}
				if count, err = xbindata.Translate(c); err != nil {
					return fmt.Errorf("cfg #%d [%s]: translate failed: %v", i, cfg.Pkg, err)
				}
				log.Printf("done with %d assets.\n", count)
			}

			return
		},
	}
)

func init() {
	rootCmd.AddCommand(buildCmd)
	flag := buildCmd.Flags()
	flag.BoolP("program", "P", false, "build outlined and append contents into program")
	flag.StringP("outlined-output-dir", "d", "_assets", "The outlined output root dir")
	flag.StringP("outlined-output-local-dir", "D", "_assets", "The outlined Local FS root dir")

	buildCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./.xb.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigName(".xb")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		cfgFile = viper.ConfigFileUsed()
		fmt.Println("Using config file:", cfgFile)
	} else if !os.IsNotExist(err) {
		panic(fmt.Errorf("load config `%v` failed: %v", viper.ConfigFileUsed(), err))
	}
}
