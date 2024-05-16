// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package usercenter

import (
	"time"

	"github.com/spf13/pflag"

	genericoptions "github.com/superproj/onex/pkg/options"
)

var _ genericoptions.IOptions = (*UserCenterOptions)(nil)

// UserCenterOptions is a list of options for the specific client.
type UserCenterOptions struct {
	Server string `json:"server" mapstructure:"server"`

	// Timeout with server timeout.
	Timeout time.Duration `json:"timeout" mapstructure:"timeout"`
}

// NewUserCenterOptions returns initialized UserCenterOptions.
func NewUserCenterOptions() *UserCenterOptions {
	return &UserCenterOptions{
		Server:  "127.0.0.1:8081",
		Timeout: 30 * time.Second,
	}
}

// Validate validates all the required options.
func (o *UserCenterOptions) Validate() []error {
	if o == nil {
		return nil
	}

	allErrs := []error{}
	return allErrs
}

// AddFlags adds flags for a specific APIServer to the specified FlagSet.
func (o *UserCenterOptions) AddFlags(fs *pflag.FlagSet, prefixes ...string) {
	if o == nil {
		return
	}

	fs.StringVar(&o.Server, "usercenter.server", o.Server, "UserCenter server to request with (ip:port).")
	fs.DurationVar(&o.Timeout, "usercenter.timeout", o.Timeout, "Timeout for usercenter server connections.")
}
