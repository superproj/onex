// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package options

import (
	"time"

	"github.com/spf13/pflag"
)

// PolarisOptions defines options for Polaris service.
type PolarisOptions struct {
	Addr         string        `json:"addr" mapstructure:"addr"`
	ReadTimeout  time.Duration `json:"read-timeout" mapstructure:"read-timeout"`
	WriteTimeout time.Duration `json:"write-timeout" mapstructure:"write-timeout"`
}

// NewPolarisOptions create a `zero` value instance.
func NewPolarisOptions() *PolarisOptions {
	return &PolarisOptions{
		Addr:         "127.0.0.1:8080",
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}
}

// Validate verifies flags passed to PolarisOptions.
func (o *PolarisOptions) Validate() []error {
	errs := []error{}
	return errs
}

// AddFlags adds flags related to Polaris service to the specified FlagSet.
func (o *PolarisOptions) AddFlags(fs *pflag.FlagSet, prefixes ...string) {
	fs.StringVar(&o.Addr, "polaris.addr", o.Addr, "Address of your Polaris service(ip:port).")
	fs.DurationVar(&o.ReadTimeout, "polaris.read-timeout", o.ReadTimeout, "Timeout for socket reads.")
	fs.DurationVar(&o.WriteTimeout, "polaris.write-timeout", o.WriteTimeout, "Timeout for socket writes.")
}
