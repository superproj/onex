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
	"reflect"
)

// AllPtrFieldsNil tests whether all pointer fields in a struct are nil.  This is useful when,
// for example, an API struct is handled by plugins which need to distinguish
// "no plugin accepted this spec" from "this spec is empty".
//
// This function is only valid for structs and pointers to structs.  Any other
// type will cause a panic.  Passing a typed nil pointer will return true.
func AllPtrFieldsNil(obj interface{}) bool {
	v := reflect.ValueOf(obj)
	if !v.IsValid() {
		panic(fmt.Sprintf("reflect.ValueOf() produced a non-valid Value for %#v", obj))
	}
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return true
		}
		v = v.Elem()
	}
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).Kind() == reflect.Ptr && !v.Field(i).IsNil() {
			return false
		}
	}
	return true
}

// To returns a pointer to the given value.
func To[T any](v T) *T {
	return &v
}

// From returns the value pointed to by the pointer p.
// If the pointer is nil, returns the zero value of T instead.
func From[T any](v *T) T {
	var zero T
	if v != nil {
		return *v
	}

	return zero
}

// FromOr dereferences ptr and returns the value it points to if no nil, or else
// returns def.
func FromOr[T any](ptr *T, def T) T {
	if ptr != nil {
		return *ptr
	}
	return def
}

// IsNil returns whether the given pointer v is nil.
func IsNil[T any](p *T) bool {
	return p == nil
}

// IsNotNil is negation of [IsNil].
func IsNotNil[T any](p *T) bool {
	return p != nil
}

// Clone returns a shallow copy of the slice.
// If the given pointer is nil, nil is returned.
//
// HINT: The element is copied using assignment (=), so this is a shallow clone.
// If you want to do a deep clone, use [CloneBy] with an appropriate element
// clone function.
//
// AKA: Copy
func Clone[T any](p *T) *T {
	if p == nil {
		return nil
	}
	clone := *p
	return &clone
}

// CloneBy is variant of [Clone], it returns a copy of the map.
// Element is copied using function f.
// If the given pointer is nil, nil is returned.
func CloneBy[T any](p *T, f func(T) T) *T {
	return Map(p, f)
}

// Equal returns true if both arguments are nil or both arguments
// dereference to the same value.
func Equal[T comparable](a, b *T) bool {
	if (a == nil) != (b == nil) {
		return false
	}
	if a == nil {
		return true
	}
	return *a == *b
}

// EqualTo returns whether the value of pointer p is equal to value v.
// It a shortcut of "x != nil && *x == y".
//
// EXAMPLE:
//
//	x, y := 1, 2
//	Equal(&x, 1)   ⏩  true
//	Equal(&y, 1)   ⏩n false
//	Equal(nil, 1)  ⏩  false
func EqualTo[T comparable](p *T, v T) bool {
	return p != nil && *p == v
}

// Map applies function f to element of pointer p.
// If p is nil, f will not be called and nil is returned, otherwise,
// result of f are returned as a new pointer.
//
// EXAMPLE:
//
//	i := 1
//	Map(&i, strconv.Itoa)       ⏩  (*string)("1")
//	Map[int](nil, strconv.Itoa) ⏩  (*string)(nil)
func Map[F, T any](p *F, f func(F) T) *T {
	if p == nil {
		return nil
	}
	return To(f(*p))
}
