// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package options

import (
	"github.com/spf13/pflag"
)

var _ IOptions = (*ConsulOptions)(nil)

// ConsulOptions defines options for consul client.
type ConsulOptions struct {
	// Address is the address of the Consul server
	Addr string `json:"addr,omitempty" mapstructure:"addr"`

	// Scheme is the URI scheme for the Consul server
	Scheme string `json:"scheme,omitempty" mapstructure:"scheme"`
}

// NewConsulOptions create a `zero` value instance.
func NewConsulOptions() *ConsulOptions {
	return &ConsulOptions{
		Addr:   "127.0.0.1:8500",
		Scheme: "http",
	}
}

// Validate verifies flags passed to ConsulOptions.
func (o *ConsulOptions) Validate() []error {
	errs := []error{}

	return errs
}

// AddFlags adds flags related to mysql storage for a specific APIServer to the specified FlagSet.
func (o *ConsulOptions) AddFlags(fs *pflag.FlagSet, prefixes ...string) {
	fs.StringVar(&o.Addr, "consul.addr", o.Addr, ""+
		"Addr is the address of the consul server.")

	fs.StringVar(&o.Scheme, "consul.scheme", o.Scheme, ""+
		"Scheme is the URI scheme for the consul server.")
}
