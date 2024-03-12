// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package options

import (
	"fmt"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
)

// Defines flag for onexctl.
const (
	FlagConfig = "config"
)

// Options composes the set of values necessary for obtaining onex service config.
type Options struct {
	Config string

	WrapConfigFn      func() error
	UserOptions       *UserOptions   `json:"user" mapstructure:"user"`
	UserCenterOptions *ServerOptions `json:"usercenter" mapstructure:"usercenter"`
	GatewayOptions    *ServerOptions `json:"gateway" mapstructure:"gateway"`
}

func (o *Options) Complete() {
	if err := viper.Unmarshal(&o); err != nil {
		panic(err)
	}
}

// AddFlags binds client configuration flags to a given flagset.
func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Config, FlagConfig, o.Config, fmt.Sprintf("Path to the %s file to use for CLI.", FlagConfig))
	o.UserOptions.AddFlags(fs)
	o.UserCenterOptions.AddFlags(fs, "usercenter")
	o.GatewayOptions.AddFlags(fs, "gateway")
}

// Validate validates ServerRunOptions.
func (o *Options) Validate() error {
	errors := []error{}
	errors = append(errors, o.UserOptions.Validate()...)
	errors = append(errors, o.UserCenterOptions.Validate()...)
	errors = append(errors, o.GatewayOptions.Validate()...)
	return utilerrors.NewAggregate(errors)
}

// NewOptions returns ConfigFlags with default values set.
func NewOptions() *Options {
	return &Options{
		UserOptions:       NewUserOptions(),
		UserCenterOptions: NewServerOptions(),
		GatewayOptions:    NewServerOptions(),
	}
}
