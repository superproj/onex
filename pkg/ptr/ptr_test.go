/*
Copyright 2023 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ptr

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAllPtrFieldsNil(t *testing.T) {
	testCases := []struct {
		obj      interface{}
		expected bool
	}{
		{struct{}{}, true},
		{struct{ Foo int }{12345}, true},
		{&struct{ Foo int }{12345}, true},
		{struct{ Foo *int }{nil}, true},
		{&struct{ Foo *int }{nil}, true},
		{struct {
			Foo int
			Bar *int
		}{12345, nil}, true},
		{&struct {
			Foo int
			Bar *int
		}{12345, nil}, true},
		{struct {
			Foo *int
			Bar *int
		}{nil, nil}, true},
		{&struct {
			Foo *int
			Bar *int
		}{nil, nil}, true},
		{struct{ Foo *int }{new(int)}, false},
		{&struct{ Foo *int }{new(int)}, false},
		{struct {
			Foo *int
			Bar *int
		}{nil, new(int)}, false},
		{&struct {
			Foo *int
			Bar *int
		}{nil, new(int)}, false},
		{(*struct{})(nil), true},
	}
	for i, tc := range testCases {
		name := fmt.Sprintf("case[%d]", i)
		t.Run(name, func(t *testing.T) {
			if actual := AllPtrFieldsNil(tc.obj); actual != tc.expected {
				t.Errorf("%s: expected %t, got %t", name, tc.expected, actual)
			}
		})
	}
}

func TestRef(t *testing.T) {
	type T int

	val := T(0)
	pointer := To(val)
	if *pointer != val {
		t.Errorf("expected %d, got %d", val, *pointer)
	}

	val = T(1)
	pointer = To(val)
	if *pointer != val {
		t.Errorf("expected %d, got %d", val, *pointer)
	}
}

func TestFrom(t *testing.T) {
	assert.Equal(t, 543, From(To(543)))
	assert.Equal(t, "Alice", From(To("Alice")))
	assert.Zero(t, From[int](nil))
	assert.Nil(t, From[interface{}](nil))
	assert.Nil(t, From(To[fmt.Stringer](nil)))
}

func TestFromOr(t *testing.T) {
	type T int

	var val, def T = 1, 0

	out := FromOr(&val, def)
	if out != val {
		t.Errorf("expected %d, got %d", val, out)
	}

	out = FromOr(nil, def)
	if out != def {
		t.Errorf("expected %d, got %d", def, out)
	}
}

func TestIsNil(t *testing.T) {
	assert.False(t, IsNil(To(1)))
	assert.True(t, IsNil[int](nil))
}

func TestClone(t *testing.T) {
	assert.True(t, IsNil(Clone(((*int)(nil)))))

	v := 1
	assert.True(t, Clone(&v) != &v)
	assert.True(t, Equal(Clone(&v), &v))

	src := To(1)
	dst := Clone(&src)
	assert.Equal(t, &src, dst)
	assert.True(t, src == *dst)
}

func TestCloneBy(t *testing.T) {
	assert.True(t, IsNil(CloneBy(((**int)(nil)), Clone[int])))

	src := To(1)
	dst := CloneBy(&src, Clone[int])
	assert.Equal(t, &src, dst)
	assert.False(t, src == *dst)
}

func TestEqual(t *testing.T) {
	type T int

	if !Equal[T](nil, nil) {
		t.Errorf("expected true (nil == nil)")
	}
	if !Equal(To(T(123)), To(T(123))) {
		t.Errorf("expected true (val == val)")
	}
	if Equal(nil, To(T(123))) {
		t.Errorf("expected false (nil != val)")
	}
	if Equal(To(T(123)), nil) {
		t.Errorf("expected false (val != nil)")
	}
	if Equal(To(T(123)), To(T(456))) {
		t.Errorf("expected false (val != val)")
	}
}

func TestEqualTo(t *testing.T) {
	assert.True(t, EqualTo(To(1), 1))
	assert.False(t, EqualTo(To(2), 1))
	assert.False(t, EqualTo(nil, 0))
}

func TestMap(t *testing.T) {
	i := 1
	assert.Equal(t, To("1"), Map(&i, strconv.Itoa))
	assert.True(t, Map(nil, strconv.Itoa) == nil)

	assert.NotPanics(t, func() {
		_ = Map(nil, func(int) string {
			panic("Q_Q")
		})
	})

	assert.Panics(t, func() {
		_ = Map(&i, func(int) string {
			panic("Q_Q")
		})
	})
}
