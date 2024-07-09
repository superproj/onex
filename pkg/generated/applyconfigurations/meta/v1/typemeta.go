// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.

// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1

// TypeMetaApplyConfiguration represents an declarative configuration of the TypeMeta type for use
// with apply.
type TypeMetaApplyConfiguration struct {
	Kind       *string `json:"kind,omitempty"`
	APIVersion *string `json:"apiVersion,omitempty"`
}

// TypeMetaApplyConfiguration constructs an declarative configuration of the TypeMeta type for use with
// apply.
func TypeMeta() *TypeMetaApplyConfiguration {
	return &TypeMetaApplyConfiguration{}
}

// WithKind sets the Kind field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Kind field is set to the value of the last call.
func (b *TypeMetaApplyConfiguration) WithKind(value string) *TypeMetaApplyConfiguration {
	b.Kind = &value
	return b
}

// WithAPIVersion sets the APIVersion field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the APIVersion field is set to the value of the last call.
func (b *TypeMetaApplyConfiguration) WithAPIVersion(value string) *TypeMetaApplyConfiguration {
	b.APIVersion = &value
	return b
}