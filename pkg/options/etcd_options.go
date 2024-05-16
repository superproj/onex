// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package options

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"
)

var _ IOptions = (*EtcdOptions)(nil)

// EtcdOptions defines options for etcd cluster.
type EtcdOptions struct {
	Endpoints   []string      `json:"endpoints"               mapstructure:"endpoints"`
	DialTimeout time.Duration `json:"dial-timeout"         mapstructure:"dial-timeout"`
	Username    string        `json:"username"                mapstructure:"username"`
	Password    string        `json:"password"                mapstructure:"password"`
	TLSOptions  *TLSOptions   `json:"tls"               mapstructure:"tls"`
}

// NewEtcdOptions create a `zero` value instance.
func NewEtcdOptions() *EtcdOptions {
	return &EtcdOptions{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
		TLSOptions:  NewTLSOptions(),
	}
}

// Validate verifies flags passed to EtcdOptions.
func (o *EtcdOptions) Validate() []error {
	errs := []error{}

	if len(o.Endpoints) == 0 {
		errs = append(errs, fmt.Errorf("--etcd.endpoints can not be empty"))
	}

	if o.DialTimeout <= 0 {
		errs = append(errs, fmt.Errorf("--etcd.dial-timeout cannot be negative"))
	}

	errs = append(errs, o.TLSOptions.Validate()...)

	return errs
}

// AddFlags adds flags related to redis storage for a specific APIServer to the specified FlagSet.
func (o *EtcdOptions) AddFlags(fs *pflag.FlagSet, prefixes ...string) {
	o.TLSOptions.AddFlags(fs, "etcd")

	fs.StringSliceVar(&o.Endpoints, "etcd.endpoints", o.Endpoints, "Endpoints of etcd cluster.")
	fs.StringVar(&o.Username, "etcd.username", o.Username, "Username of etcd cluster.")
	fs.StringVar(&o.Password, "etcd.password", o.Password, "Password of etcd cluster.")
	fs.DurationVar(&o.DialTimeout, "etcd.dial-timeout", o.DialTimeout, "Etcd dial timeout in seconds.")
}
