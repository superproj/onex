// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package options

import (
	"github.com/spf13/pflag"
)

var _ IOptions = (*InsecureServingOptions)(nil)

// InsecureServingOptions are for creating an unauthenticated, unauthorized, insecure port.
// No one should be using these anymore.
type InsecureServingOptions struct {
	Addr string `json:"addr" mapstructure:"addr"`
}

// NewInsecureServingOptions is for creating an unauthenticated, unauthorized, insecure port.
// No one should be using these anymore.
func NewInsecureServingOptions() *InsecureServingOptions {
	return &InsecureServingOptions{
		Addr: "127.0.0.1:8080",
	}
}

// Validate is used to parse and validate the parameters entered by the user at
// the command line when the program starts.
func (s *InsecureServingOptions) Validate() []error {
	var errors []error

	return errors
}

// AddFlags adds flags related to features for a specific api server to the
// specified FlagSet.
func (s *InsecureServingOptions) AddFlags(fs *pflag.FlagSet, prefixes ...string) {
	fs.StringVar(&s.Addr, "insecure.addr", s.Addr, "Specify the HTTP server bind address and port.")
}
