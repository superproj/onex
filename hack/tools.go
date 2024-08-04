// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

//go:build tools
// +build tools

// This package imports things required by build hack, to force `go mod` to see them as dependencies
package tools

import _ "k8s.io/code-generator"
