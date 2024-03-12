// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MySQLConfiguration defines the configuration of mysql
// clients for components that can run with mysql database.
type MySQLConfiguration struct {
	// MySQL service host address. If left blank, the following related mysql options will be ignored.
	Host string `json:"host"`
	// Username for access to mysql service.
	Username string `json:"username"`
	// Password for access to mysql, should be used pair with password.
	Password string `json:"password"`
	// Database name for the server to use.
	Database string `json:"database"`
	// Maximum idle connections allowed to connect to mysql.
	MaxIdleConnections int32 `json:"maxIdleConnections"`
	// Maximum open connections allowed to connect to mysql.
	MaxOpenConnections int32 `json:"maxOpenConnections"`
	// Maximum connection life time allowed to connect to mysql.
	MaxConnectionLifeTime metav1.Duration `json:"maxConnectionLifeTime"`
}

// RedisConfiguration defines the configuration of redis
// clients for components that can run with redis key-value database.
type RedisConfiguration struct {
	// Address of your Redis server(ip:port).
	Addr string `json:"addr"`
	// Username for access to redis service.
	Username string `json:"username"`
	// Optional auth password for Redis db.
	Password string `json:"password"`
	// Database to be selected after connecting to the server.
	Database int `json:"database"`
	// Maximum number of retries before giving up.
	MaxRetries int `json:"maxRetries"`
	// Timeout when connecting to redis service.
	Timeout metav1.Duration `json:"timeout"`
}
