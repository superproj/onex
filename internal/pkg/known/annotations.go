// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package known

const (
	// This exposes compute information based on the miner type.
	CPUAnnotation    = "apps.onex.io/vCPU"
	MemoryAnnotation = "apps.onex.io/memoryMb"
)

const (
	SkipVerifyAnnotation = "apps.onex.io/skip-verify"
)

var AllImmutableAnnotations = []string{
	CPUAnnotation,
	MemoryAnnotation,
}
