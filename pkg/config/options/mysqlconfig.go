// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package options

import (
	"github.com/spf13/pflag"

	"github.com/superproj/onex/pkg/config"
)

// BindMySQLFlags binds the MySQLConfiguration struct fields to a flagset.
func BindMySQLFlags(m *config.MySQLConfiguration, fs *pflag.FlagSet) {
	fs.StringVar(&m.Host, "mysql-host", m.Host, ""+
		"MySQL service host address. If left blank, the following related mysql options will be ignored.")
	fs.StringVar(&m.Username, "mysql-username", m.Username, ""+
		"Username for access to mysql service.")
	fs.StringVar(&m.Password, "mysql-password", m.Password, ""+
		"Password for access to mysql, should be used pair with password.")
	fs.StringVar(&m.Database, "mysql-database", m.Database, ""+
		"Database name for the server to use.")
	fs.Int32Var(&m.MaxIdleConnections, "mysql-max-idle-connections", m.MaxOpenConnections, ""+
		"Maximum idle connections allowed to connect to mysql.")
	fs.Int32Var(&m.MaxOpenConnections, "mysql-max-open-connections", m.MaxOpenConnections, ""+
		"Maximum open connections allowed to connect to mysql.")
	fs.DurationVar(&m.MaxConnectionLifeTime.Duration, "mysql-max-connection-life-time", m.MaxConnectionLifeTime.Duration, ""+
		"Maximum connection life time allowed to connect to mysql.")
}
