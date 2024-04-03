// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package main

import (
	"flag"

	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/superproj/onex/internal/lint/kubelistcheck"
)

func main() {
	flag.Bool("unsafeptr", false, "")

	singlechecker.Main(kubelistcheck.NewAnalyzer(false))
}
