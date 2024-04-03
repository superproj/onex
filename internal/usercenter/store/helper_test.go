// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// FuzzDefaultLimit 模糊测试用例.
func FuzzDefaultLimit(f *testing.F) {
	testcases := []int{0, 1, 2}
	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, orig int) {
		limit := defaultLimit(orig)
		if orig == 0 {
			assert.Equal(t, defaultLimitValue, limit)
		} else {
			assert.Equal(t, orig, limit)
		}
	})
}
