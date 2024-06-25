// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package app

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"k8s.io/client-go/util/homedir"
	cliflag "k8s.io/component-base/cli/flag"

	"github.com/superproj/onex/pkg/log"
)

const configFlagName = "config"

// CliOptions abstracts configuration options for reading parameters from the
// command line.
type CliOptions interface {
	// Flags returns flags for a specific server by section name.
	Flags() cliflag.NamedFlagSets

	// Complete completes all the required options.
	Complete() error

	// Validate validates all the required options.
	Validate() error
}

var cfgFile string

// AddConfigFlag adds flags for a specific server to the specified FlagSet object.
// It also sets a passed functions to read values from configuration file into viper
// when each cobra command's Execute method is called.
func AddConfigFlag(fs *pflag.FlagSet, name string, watch bool) {
	fs.AddFlag(pflag.Lookup(configFlagName))

	// Enable viper's automatic environment variable parsing. This means
	// that viper will automatically read values corresponding to viper
	// variables from environment variables.
	viper.AutomaticEnv()
	// Set the environment variable prefix. Use the strings.ReplaceAll function
	// to replace hyphens with underscores in the name, and use strings.ToUpper
	// to convert the name to uppercase, then set it as the prefix for environment variables.
	viper.SetEnvPrefix(strings.ReplaceAll(strings.ToUpper(name), "-", "_"))
	// Set the replacement rules for environment variable keys. Use the
	// strings.NewReplacer function to specify replacing periods and hyphens with underscores.
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	cobra.OnInitialize(func() {
		if cfgFile != "" {
			viper.SetConfigFile(cfgFile)
		} else {
			viper.AddConfigPath(".")

			if names := strings.Split(name, "-"); len(names) > 1 {
				viper.AddConfigPath(filepath.Join(homedir.HomeDir(), "."+names[0]))
				viper.AddConfigPath(filepath.Join("/etc", names[0]))
			}

			viper.SetConfigName(name)
		}

		if err := viper.ReadInConfig(); err != nil {
			log.Debugw("Failed to read configuration file", "file", cfgFile, "err", err)
		}
		log.Debugw("Success to read configuration file", "file", viper.ConfigFileUsed())

		if watch {
			viper.WatchConfig()
			viper.OnConfigChange(func(e fsnotify.Event) {
				log.Debugw("Config file changed", "name", e.Name)
			})
		}
	})
}

func PrintConfig() {
	for _, key := range viper.AllKeys() {
		log.Debugw(fmt.Sprintf("CFG: %s=%v", key, viper.Get(key)))
	}
}

func init() {
	pflag.StringVarP(&cfgFile, configFlagName, "c", cfgFile, "Read configuration from specified `FILE`, "+
		"support JSON, TOML, YAML, HCL, or Java properties formats.")
}
