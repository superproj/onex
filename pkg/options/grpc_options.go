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

var _ IOptions = (*GRPCOptions)(nil)

// GRPCOptions are for creating an unauthenticated, unauthorized, insecure port.
// No one should be using these anymore.
type GRPCOptions struct {
	// Network with server network.
	Network string `json:"network" mapstructure:"network"`

	// Address with server address.
	Addr string `json:"addr" mapstructure:"addr"`

	// Timeout with server timeout. Used by grpc client side.
	Timeout time.Duration `json:"timeout" mapstructure:"timeout"`
}

// NewGRPCOptions is for creating an unauthenticated, unauthorized, insecure port.
// No one should be using these anymore.
func NewGRPCOptions() *GRPCOptions {
	return &GRPCOptions{
		Network: "tcp",
		Addr:    "0.0.0.0:39090",
		Timeout: 30 * time.Second,
	}
}

// Validate is used to parse and validate the parameters entered by the user at
// the command line when the program starts.
func (o *GRPCOptions) Validate() []error {
	var errors []error

	if err := ValidateAddress(o.Addr); err != nil {
		errors = append(errors, err)
	}

	return errors
}

// AddFlags adds flags related to features for a specific api server to the
// specified FlagSet.
func (o *GRPCOptions) AddFlags(fs *pflag.FlagSet, prefixes ...string) {
	fs.StringVar(&o.Network, "grpc.network", o.Network, "Specify the network for the gRPC server.")
	fs.StringVar(&o.Addr, "grpc.addr", o.Addr, "Specify the gRPC server bind address and port.")
	fs.DurationVar(&o.Timeout, "grpc.timeout", o.Timeout, "Timeout for server connections.")
}
