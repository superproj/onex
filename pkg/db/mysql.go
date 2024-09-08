// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package db

import (
	"fmt"
	"time"

	"github.com/superproj/onex/pkg/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// MySQLOptions defines options for mysql database.
type MySQLOptions struct {
	Host                  string
	Username              string
	Password              string
	Database              string
	MaxIdleConnections    int
	MaxOpenConnections    int
	MaxConnectionLifeTime time.Duration
	LogLevel              int
	// +optional
	Logger logger.Interface
}

// DSN return DSN from MySQLOptions.
func (o *MySQLOptions) DSN() string {
	return fmt.Sprintf(`%s:%s@tcp(%s)/%s?charset=utf8&parseTime=%t&loc=%s`,
		o.Username,
		o.Password,
		o.Host,
		o.Database,
		true,
		"Local")
}

// NewMySQL create a new gorm db instance with the given options.
func NewMySQL(opts *MySQLOptions) (*gorm.DB, error) {
	// Set default values to ensure all fields in opts are available.
	setDefaults(opts)

	db, err := gorm.Open(mysql.Open(opts.DSN()), &gorm.Config{
		// PrepareStmt executes the given query in cached statement.
		// This can improve performance.
		PrepareStmt: true,
		Logger:      opts.Logger,
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(opts.MaxOpenConnections)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(opts.MaxConnectionLifeTime)

	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(opts.MaxIdleConnections)

	return db, nil
}

// setDefaults set available default values for some fields.
func setDefaults(opts *MySQLOptions) {
	if opts.Host == "" {
		opts.Host = "127.0.0.1:3306"
	}
	if opts.MaxIdleConnections == 0 {
		opts.MaxIdleConnections = 100
	}
	if opts.MaxOpenConnections == 0 {
		opts.MaxOpenConnections = 100
	}
	if opts.MaxConnectionLifeTime == 0 {
		opts.MaxConnectionLifeTime = time.Duration(10) * time.Second
	}
	if 1 <= opts.LogLevel && opts.LogLevel <= 4 {
		opts.Logger = log.Default().LogMode(logger.LogLevel(opts.LogLevel))
	}
	if opts.Logger == nil {
		opts.Logger = logger.Default
	}
}
