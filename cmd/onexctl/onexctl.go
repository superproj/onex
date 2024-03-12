// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package main

import (
	"k8s.io/component-base/cli"

	"github.com/superproj/onex/internal/onexctl/cmd"
	"github.com/superproj/onex/internal/onexctl/cmd/util"
)

func main() {
	command := cmd.NewDefaultOneXCtlCommand()
	if err := cli.RunNoErrOutput(command); err != nil {
		// Pretty-print the error and exit with an error.
		util.CheckErr(err)
	}
}
