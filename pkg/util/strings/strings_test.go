// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package strings

import "testing"

func TestDiff(t *testing.T) {
	testCase := [][]string{
		{"foo", "bar", "hello"},
		{"foo", "bar", "world"},
	}
	result := Diff(testCase[0], testCase[1])
	if len(result) != 1 || result[0] != "hello" {
		t.Fatalf("Diff failed")
	}
}
