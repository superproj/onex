// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package contract

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestPath_Append(t *testing.T) {
	g := NewWithT(t)

	got0 := Path{}.Append("foo")
	g.Expect(got0).To(Equal(Path{"foo"}))
	g.Expect(got0.String()).To(Equal("foo"))

	got1 := Path{"foo"}.Append("bar")
	g.Expect(got1).To(Equal(Path{"foo", "bar"}))
	g.Expect(got1.String()).To(Equal("foo.bar"))
}

func TestPath_IsParenOf(t *testing.T) {
	tests := []struct {
		name  string
		p     Path
		other Path
		want  bool
	}{
		{
			name:  "True for parent path",
			p:     Path{"foo"},
			other: Path{"foo", "bar"},
			want:  true,
		},
		{
			name:  "False for same path",
			p:     Path{"foo"},
			other: Path{"foo"},
			want:  false,
		},
		{
			name:  "False for child path",
			p:     Path{"foo", "bar"},
			other: Path{"foo"},
			want:  false,
		},
		{
			name:  "False for not overlapping path",
			p:     Path{"foo", "bar"},
			other: Path{"baz"},
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			got := tt.p.IsParentOf(tt.other)
			g.Expect(got).To(Equal(tt.want))
		})
	}
}

func TestPath_Equal(t *testing.T) {
	tests := []struct {
		name  string
		p     Path
		other Path
		want  bool
	}{
		{
			name:  "False for parent path",
			p:     Path{"foo"},
			other: Path{"foo", "bar"},
			want:  false,
		},
		{
			name:  "True for same path",
			p:     Path{"foo"},
			other: Path{"foo"},
			want:  true,
		},
		{
			name:  "False for child path",
			p:     Path{"foo", "bar"},
			other: Path{"foo"},
			want:  false,
		},
		{
			name:  "False for not overlapping path",
			p:     Path{"foo", "bar"},
			other: Path{"baz"},
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			got := tt.p.Equal(tt.other)
			g.Expect(got).To(Equal(tt.want))
		})
	}
}

func TestPath_Overlaps(t *testing.T) {
	tests := []struct {
		name  string
		p     Path
		other Path
		want  bool
	}{
		{
			name:  "True for parent path",
			p:     Path{"foo"},
			other: Path{"foo", "bar"},
			want:  true,
		},
		{
			name:  "True for same path",
			p:     Path{"foo"},
			other: Path{"foo"},
			want:  true,
		},
		{
			name:  "True for child path",
			p:     Path{"foo", "bar"},
			other: Path{"foo"},
			want:  true,
		},
		{
			name:  "False for not overlapping path",
			p:     Path{"foo", "bar"},
			other: Path{"baz"},
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			got := tt.p.Overlaps(tt.other)
			g.Expect(got).To(Equal(tt.want))
		})
	}
}
