// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

// +k8s:deepcopy-gen=package

// Package core contains the latest (or "internal") version of the
// OneX API objects. This is the API objects as represented in memory.
// The contract presented to clients is located in the versioned packages,
// which are sub-directories. The first one is "v1".
package core // import "github.com/superproj/onex/pkg/apis/core"
