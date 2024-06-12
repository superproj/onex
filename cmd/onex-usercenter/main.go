// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

// usercenter is the user center of the onex cloud platform.
package main

import (
	// Importing the package to automatically set GOMAXPROCS.
	_ "go.uber.org/automaxprocs/maxprocs"

	"github.com/superproj/onex/cmd/onex-usercenter/app"
)

func main() {
	// Creating a new instance of the usercenter application and running it
	app.NewApp().Run()
}
