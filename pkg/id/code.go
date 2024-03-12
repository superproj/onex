// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package id

// NewCode can get a unique code by id(You need to ensure that id is unique).
func NewCode(id uint64, options ...func(*CodeOptions)) string {
	ops := getCodeOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	// enlarge and add salt
	id = id*uint64(ops.n1) + ops.salt

	var code []rune
	slIdx := make([]byte, ops.l)

	charLen := len(ops.chars)
	charLenUI := uint64(charLen)

	// diffusion
	for i := 0; i < ops.l; i++ {
		slIdx[i] = byte(id % charLenUI)                          // get each number
		slIdx[i] = (slIdx[i] + byte(i)*slIdx[0]) % byte(charLen) // let units digit affect other digit
		id /= charLenUI                                          // right shift
	}

	// confusion(https://en.wikipedia.org/wiki/Permutation_box)
	for i := 0; i < ops.l; i++ {
		idx := (byte(i) * byte(ops.n2)) % byte(ops.l)
		code = append(code, ops.chars[slIdx[idx]])
	}
	return string(code)
}
