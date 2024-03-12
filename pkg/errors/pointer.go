// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package errors

// MinerStatusErrorPtr converts a MinerStatusError to a pointer.
func MinerStatusErrorPtr(v MinerStatusError) *MinerStatusError {
	return &v
}

// MinerSetStatusErrorPtr converts a MinerSetStatusError to a pointer.
func MinerSetStatusErrorPtr(v MinerSetStatusError) *MinerSetStatusError {
	return &v
}
