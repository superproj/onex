// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

// cacheserver is a standard, specification-compliant demo example of the onex service.
// cacheserver is also a gRPC and HTTP server.
package main

import (
	_ "go.uber.org/automaxprocs/maxprocs"

	"github.com/superproj/onex/cmd/onex-cacheserver/app"
)

func main() {
	app.NewApp("onex-cacheserver").Run()
}
