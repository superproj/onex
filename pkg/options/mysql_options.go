// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package options

import (
	"time"

	"github.com/spf13/pflag"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/superproj/onex/pkg/db"
	"github.com/superproj/onex/pkg/log"
)

var _ IOptions = (*MySQLOptions)(nil)

// MySQLOptions defines options for mysql database.
type MySQLOptions struct {
	Host                  string        `json:"host,omitempty" mapstructure:"host"`
	Username              string        `json:"username,omitempty" mapstructure:"username"`
	Password              string        `json:"-" mapstructure:"password"`
	Database              string        `json:"database" mapstructure:"database"`
	MaxIdleConnections    int           `json:"max-idle-connections,omitempty" mapstructure:"max-idle-connections,omitempty"`
	MaxOpenConnections    int           `json:"max-open-connections,omitempty" mapstructure:"max-open-connections"`
	MaxConnectionLifeTime time.Duration `json:"max-connection-life-time,omitempty" mapstructure:"max-connection-life-time"`
	LogLevel              int           `json:"log-level" mapstructure:"log-level"`
}

// NewMySQLOptions create a `zero` value instance.
func NewMySQLOptions() *MySQLOptions {
	return &MySQLOptions{
		Host:                  "127.0.0.1:3306",
		Username:              "onex",
		Password:              "onex(#)666",
		Database:              "onex",
		MaxIdleConnections:    100,
		MaxOpenConnections:    100,
		MaxConnectionLifeTime: time.Duration(10) * time.Second,
		LogLevel:              1, // Silent
	}
}

// Validate verifies flags passed to MySQLOptions.
func (o *MySQLOptions) Validate() []error {
	errs := []error{}

	return errs
}

// AddFlags adds flags related to mysql storage for a specific APIServer to the specified FlagSet.
func (o *MySQLOptions) AddFlags(fs *pflag.FlagSet, prefixes ...string) {
	fs.StringVar(&o.Host, join(prefixes...)+"db.host", o.Host, ""+
		"MySQL service host address. If left blank, the following related mysql options will be ignored.")
	fs.StringVar(&o.Username, join(prefixes...)+"db.username", o.Username, "Username for access to mysql service.")
	fs.StringVar(&o.Password, join(prefixes...)+"db.password", o.Password, ""+
		"Password for access to mysql, should be used pair with password.")
	fs.StringVar(&o.Database, join(prefixes...)+"db.database", o.Database, ""+
		"Database name for the server to use.")
	fs.IntVar(&o.MaxIdleConnections, join(prefixes...)+"db.max-idle-connections", o.MaxOpenConnections, ""+
		"Maximum idle connections allowed to connect to mysql.")
	fs.IntVar(&o.MaxOpenConnections, join(prefixes...)+"db.max-open-connections", o.MaxOpenConnections, ""+
		"Maximum open connections allowed to connect to mysql.")
	fs.DurationVar(&o.MaxConnectionLifeTime, join(prefixes...)+"db.max-connection-life-time", o.MaxConnectionLifeTime, ""+
		"Maximum connection life time allowed to connect to mysql.")
	fs.IntVar(&o.LogLevel, join(prefixes...)+"db.log-mode", o.LogLevel, ""+
		"Specify gorm log level.")
}

// NewDB create mysql store with the given config.
func (o *MySQLOptions) NewDB() (*gorm.DB, error) {
	opts := &db.MySQLOptions{
		Host:                  o.Host,
		Username:              o.Username,
		Password:              o.Password,
		Database:              o.Database,
		MaxIdleConnections:    o.MaxIdleConnections,
		MaxOpenConnections:    o.MaxOpenConnections,
		MaxConnectionLifeTime: o.MaxConnectionLifeTime,
		Logger:                log.Default().LogMode(gormlogger.LogLevel(o.LogLevel)),
	}

	return db.NewMySQL(opts)
}
