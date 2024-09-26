// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

//go:build tools
// +build tools

// This package imports things required by build scripts and test packages of submodules, to force `go mod` to see them as dependencies
package tools

import (
	// build script dependencies
	_ "github.com/onsi/ginkgo/v2/ginkgo"
	_ "k8s.io/code-generator/cmd/go-to-protobuf"
	_ "k8s.io/code-generator/cmd/go-to-protobuf/protoc-gen-gogo"
	_ "k8s.io/gengo/examples/deepcopy-gen/generators"
	_ "k8s.io/gengo/examples/defaulter-gen/generators"
	_ "k8s.io/gengo/examples/import-boss/generators"
	_ "k8s.io/gengo/examples/set-gen/generators"
	_ "k8s.io/kube-openapi/cmd/openapi-gen"

	// submodule test dependencies
	_ "github.com/armon/go-socks5" // for staging/src/k8s.io/apimachinery/pkg/util/httpstream/spdy/roundtripper_test.go
)
