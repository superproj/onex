// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package id

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCode(t *testing.T) {
	type args struct {
		id      uint64
		options []func(*CodeOptions)
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "default", args: args{id: 1}, want: "VHB4JX86"},
		{
			name: "with-options",
			args: args{
				id: 1,
				options: []func(*CodeOptions){
					WithCodeChars([]rune{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}),
					WithCodeN1(9),
					WithCodeN2(3),
					WithCodeL(5),
					WithCodeSalt(56789),
				},
			},
			want: "80773",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NewCode(tt.args.id, tt.args.options...))
		})
	}
}

func BenchmarkNewCode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewCode(1)
	}
}

func BenchmarkNewCodeTimeConsuming(b *testing.B) {
	b.StopTimer() // 调用该函数停止压力测试的时间计数

	id := NewCode(1)
	assert.Equal(b, "VHB4JX86", id)

	b.StartTimer() // 重新开始时间

	for i := 0; i < b.N; i++ {
		NewCode(1)
	}
}
