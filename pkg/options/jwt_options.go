// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package options

import (
	"fmt"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/spf13/pflag"
)

var _ IOptions = (*JWTOptions)(nil)

// JWTOptions contains configuration items related to API server features.
type JWTOptions struct {
	Key           string        `json:"key" mapstructure:"key"`
	Expired       time.Duration `json:"expired" mapstructure:"expired"`
	MaxRefresh    time.Duration `json:"max-refresh" mapstructure:"max-refresh"`
	SigningMethod string        `json:"signing-method" mapstructure:"signing-method"`
}

// NewJWTOptions creates a JWTOptions object with default parameters.
func NewJWTOptions() *JWTOptions {
	return &JWTOptions{
		// Realm:         "",
		Key:           "onex(#)666",
		Expired:       2 * time.Hour,
		MaxRefresh:    2 * time.Hour,
		SigningMethod: "HS512",
	}
}

// Validate is used to parse and validate the parameters entered by the user at
// the command line when the program starts.
func (s *JWTOptions) Validate() []error {
	var errs []error

	if !govalidator.StringLength(s.Key, "6", "32") {
		errs = append(errs, fmt.Errorf("--jwt.key must larger than 5 and little than 33"))
	}

	return errs
}

// AddFlags adds flags related to features for a specific api server to the
// specified FlagSet.
func (s *JWTOptions) AddFlags(fs *pflag.FlagSet, prefixes ...string) {
	if fs == nil {
		return
	}

	// fs.StringVar(&s.Realm, "jwt.realm", s.Realm, "Realm name to display to the user.")
	fs.StringVar(&s.Key, "jwt.key", s.Key, "Private key used to sign jwt token.")
	fs.DurationVar(&s.Expired, "jwt.expired", s.Expired, "JWT token expiration time.")
	fs.DurationVar(&s.MaxRefresh, "jwt.max-refresh", s.MaxRefresh, ""+
		"This field allows clients to refresh their token until MaxRefresh has passed.")
	fs.StringVar(&s.SigningMethod, "jwt.signing-method", s.SigningMethod, "JWT token signature method.")
}
