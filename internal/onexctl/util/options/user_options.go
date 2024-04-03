// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package options

import (
	"fmt"

	"github.com/spf13/pflag"
)

// UserOptions defines options for user client.
type UserOptions struct {
	BearerToken string `json:"token" mapstructure:"token"`
	Username    string `json:"username" mapstructure:"username"`
	Password    string `json:"password" mapstructure:"password"`
	SecretID    string `json:"secret-id" mapstructure:"secret-id"`
	SecretKey   string `json:"secret-key" mapstructure:"secret-key"`
	CertFile    string `json:"client-certificate" mapstructure:"client-certificate"`
	KeyFile     string `json:"client-key" mapstructure:"client-key"`
}

// NewUserOptions create a `zero` value instance.
func NewUserOptions() *UserOptions {
	return &UserOptions{}
}

// Validate verifies flags passed to UserOptions.
func (o *UserOptions) Validate() []error {
	errs := []error{}

	if (o.Username == "" && o.Password != "") || (o.Username != "" && o.Password == "") {
		errs = append(errs, fmt.Errorf("both username and password must be set or empty"))
	}

	if (o.SecretID == "" && o.SecretKey != "") || (o.SecretID != "" && o.SecretKey == "") {
		errs = append(errs, fmt.Errorf("both secretID and secretKey must be set or empty"))
	}

	return errs
}

// AddFlags adds flags related to mysql storage for a specific APIServer to the specified FlagSet.
func (o *UserOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.BearerToken, "user.token", o.BearerToken, "Bearer token for authentication to the API server")
	fs.StringVar(&o.Username, "user.username", o.Username, "Username for basic authentication to the API server")
	fs.StringVar(&o.Password, "user.password", o.Password, "Password for basic authentication to the API server")
	fs.StringVar(&o.SecretID, "user.secret-id", o.SecretID, "SecretID for JWT authentication to the API server")
	fs.StringVar(&o.SecretKey, "user.secret-key", o.SecretKey, "SecretKey for jwt authentication to the API server")
	fs.StringVar(&o.CertFile, "user.client-certificate", o.CertFile, "Path to a client certificate file for TLS")
	fs.StringVar(&o.KeyFile, "user.client-key", o.KeyFile, "Path to a client key file for TLS")
}
