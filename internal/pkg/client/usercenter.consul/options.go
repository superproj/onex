// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package usercenter

import (
	"github.com/spf13/pflag"
)

// Options is a list of options for the specific client.
type Options struct {
	Server string `json:"server" mapstructure:"server"`
}

// NewOptions returns initialized Options.
func NewOptions() *Options {
	return &Options{
		Server: "http://127.0.0.1:8080",
	}
}

// Validate validates all the required options.
func (o *Options) Validate() []error {
	if o == nil {
		return nil
	}

	allErrs := []error{}
	return allErrs
}

// AddFlags adds flags for a specific APIServer to the specified FlagSet.
func (o *Options) AddFlags(fs *pflag.FlagSet) {
	if o == nil {
		return
	}

	fs.StringVar(&o.Server, "usercenter.server", o.Server, ""+
		"usercenter server to request with (scheme://ip:port).")
}
