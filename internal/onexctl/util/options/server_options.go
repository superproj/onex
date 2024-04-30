// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package options

import (
	"strings"
	"time"

	"github.com/spf13/pflag"
)

// ServerOptions defines options for server client.
type ServerOptions struct {
	Insecure      bool          `json:"insecure-skip-tls-verify" mapstructure:"insecure-skip-tls-verify"`
	CAFile        string        `json:"certificate-authority" mapstructure:"certificate-authority"`
	Addr          string        `json:"addr" mapstructure:"addr"`
	Timeout       time.Duration `json:"timeout" mapstructure:"timeout"`
	MaxRetries    int           `json:"max-retries" mapstructure:"max-retries"`
	RetryInterval time.Duration `json:"retry-interval" mapstructure:"retry-interval"`
}

// NewServerOptions create a `zero` value instance.
func NewServerOptions() *ServerOptions {
	return &ServerOptions{}
}

// Validate verifies flags passed to ServerOptions.
func (o *ServerOptions) Validate() []error {
	errs := []error{}

	return errs
}

// AddFlags adds flags related to mysql storage for a specific APIServer to the specified FlagSet.
func (o *ServerOptions) AddFlags(fs *pflag.FlagSet, prefixes ...string) {
	fs.BoolVar(&o.Insecure, join(prefixes...)+"insecure-skip-tls-verify", o.Insecure, ""+
		"If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure")
	fs.StringVar(&o.CAFile, join(prefixes...)+"certificate-authority", o.CAFile, "Path to a cert file for the certificate authority")
	fs.StringVar(&o.Addr, join(prefixes...)+"address", o.Addr, "The address and port of the OneX API server")
	fs.DurationVar(&o.Timeout, join(prefixes...)+"timeout", o.Timeout, "The length of time to wait before giving up on a single "+
		"server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests.")
	fs.IntVar(&o.MaxRetries, join(prefixes...)+"max-retries", o.MaxRetries, "Maximum number of retries.")
	fs.DurationVar(&o.RetryInterval, join(prefixes...)+"retry-interval", o.RetryInterval, "The interval time between each attempt.")
}

func join(prefixes ...string) string {
	joined := strings.Join(prefixes, ".")
	if joined != "" {
		joined += "."
	}

	return joined
}
