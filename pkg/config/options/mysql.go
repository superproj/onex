// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package options

import (
	"github.com/spf13/pflag"

	genericconfig "github.com/superproj/onex/pkg/config"
)

// MySQLOptions holds the MySQL options.
type MySQLOptions struct {
	*genericconfig.MySQLConfiguration
}

func NewMySQLOptions(cfg *genericconfig.MySQLConfiguration) *MySQLOptions {
	return &MySQLOptions{
		MySQLConfiguration: cfg,
	}
}

// AddFlags adds flags related to MySQL for controller manager to the specified FlagSet.
func (o *MySQLOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Host, "mysql-host", o.Host, ""+
		"MySQL service host address. If left blank, the following related mysql options will be ignored.")
	fs.StringVar(&o.Username, "mysql-username", o.Username, ""+
		"Username for access to mysql service.")
	fs.StringVar(&o.Password, "mysql-password", o.Password, ""+
		"Password for access to mysql, should be used pair with password.")
	fs.StringVar(&o.Database, "mysql-database", o.Database, ""+
		"Database name for the server to use.")
	fs.Int32Var(&o.MaxIdleConnections, "mysql-max-idle-connections", o.MaxOpenConnections, ""+
		"Maximum idle connections allowed to connect to mysql.")
	fs.Int32Var(&o.MaxOpenConnections, "mysql-max-open-connections", o.MaxOpenConnections, ""+
		"Maximum open connections allowed to connect to mysql.")
	fs.DurationVar(&o.MaxConnectionLifeTime.Duration, "mysql-max-connection-life-time", o.MaxConnectionLifeTime.Duration, ""+
		"Maximum connection life time allowed to connect to mysql.")

}

// ApplyTo fills up MySQL config with options.
func (o *MySQLOptions) ApplyTo(cfg *genericconfig.MySQLConfiguration) error {
	if o == nil || cfg == nil {
		return nil
	}

	*cfg = *o.MySQLConfiguration
	return nil
}

// Validate checks validation of MySQL.
func (o *MySQLOptions) Validate() []error {
	if o == nil {
		return nil
	}

	errs := []error{}
	return errs
}
