// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package util

import (
	"hash/fnv"
	"time"
)

// Check panics if the given error is not nil.
func Check(e error) {
	if e != nil {
		panic(e)
	}
}

// NowNano returns UnixNano in UTC.
func NowNano() int64 {
	return time.Now().UTC().UnixNano()
}

// HashCode returns a uint32 hash for the given byte array.
func HashCode(b []byte) uint32 {
	h := fnv.New32a()
	_, err := h.Write(b)
	Check(err)
	return h.Sum32()
}
