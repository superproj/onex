// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package main

import (
	"fmt"

	"github.com/spf13/pflag"

	"github.com/superproj/onex/internal/pkg/zflag"
)

func main() {
	var accounts map[string]string
	defaultAC := map[string]string{"test": "test"}

	zflag.MapVarP(&accounts, "account", "a", defaultAC, "Authentication username and password set.")

	pflag.Parse()

	for k, v := range accounts {
		fmt.Printf("Username: %q, Password: %q\n", k, v)
	}
}
