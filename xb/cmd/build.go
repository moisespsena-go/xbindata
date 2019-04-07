// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
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
	"log"
	"os"

	"github.com/moisespsena-go/xbindata"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "build from config file",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		var cfg xbindata.ManyConfig
		if err = viper.Unmarshal(&cfg); err != nil {
			return
		}

		if err = cfg.Validate(); err != nil {
			return
		}

		for i, outlined := range cfg.Outlined {
			log.Println("==== outlined:", outlined.Pkg, " ====")
			yaml.NewEncoder(os.Stderr).Encode(&outlined)
			var c *xbindata.Config
			if c, err = outlined.Config(); err != nil {
				return fmt.Errorf("outlined #%d [%s]: create config failed: %v", i, outlined.Pkg, err)
			}
			if err = xbindata.Translate(c); err != nil {
				return fmt.Errorf("outlined #%d [%s]: translate failed: %v", i, outlined.Pkg, err)
			}
		}

		return
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
