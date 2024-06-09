// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

// The nightwatch binary is responsible for executing some async tasks like linux crontab.
package main

import (
	_ "go.uber.org/automaxprocs/maxprocs"

	"github.com/superproj/onex/cmd/onex-nightwatch/app"
)

func main() {
	app.NewApp().Run()
}
