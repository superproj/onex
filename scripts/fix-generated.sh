#!/usr/bin/env bash

ONEX_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
fixDir="${ONEX_ROOT}/pkg/generated/clientset/versioned/typed/core/v1"

function replace_generated_expansion() {
  cat << 'EOF' > $1
// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.

// Code generated by client-gen. DO NOT EDIT.

package v1

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	scheme "k8s.io/client-go/kubernetes/scheme"
)

type ConfigMapExpansion any

type EventExpansion any

type SecretExpansion any

// The NamespaceExpansion interface allows manually adding extra methods to the NamespaceInterface.
type NamespaceExpansion interface {
	Finalize(ctx context.Context, item *v1.Namespace, opts metav1.UpdateOptions) (*v1.Namespace, error)
}

// Finalize takes the representation of a namespace to update.  Returns the server's representation of the namespace, and an error, if it occurs.
func (c *namespaces) Finalize(ctx context.Context, namespace *v1.Namespace, opts metav1.UpdateOptions) (result *v1.Namespace, err error) {
	result = &v1.Namespace{}
	err = c.client.Put().Resource("namespaces").Name(namespace.Name).VersionedParams(&opts, scheme.ParameterCodec).SubResource("finalize").Body(namespace).Do(ctx).Into(result)
	return
}
EOF
}

function add_fake_finalize_method() {
  if egrep -q 'Finalize' $1;then
    return
  fi

  cat << 'EOF' >> $1

// Finalize takes the representation of a namespace to update.  Returns the server's representation of the namespace, and an error, if it occurs.
func (c *FakeNamespaces) Finalize(ctx context.Context, namespace *v1.Namespace, opts metav1.UpdateOptions) (result *v1.Namespace, err error) {
	return nil, nil
}
EOF
}

replace_generated_expansion ${fixDir}/generated_expansion.go
add_fake_finalize_method ${fixDir}/fake/fake_namespace.go
