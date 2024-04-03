// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

//nolint:gomodguard,gocritic
package contract

import (
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/superproj/onex/pkg/apis/apps/v1beta1"
)

// Metadata provides a helper struct for working with Metadata.
type Metadata struct {
	path Path
}

// Path returns the path of the metadata.
func (m *Metadata) Path() Path {
	return m.path
}

// Get gets the metadata object.
func (m *Metadata) Get(obj *unstructured.Unstructured) (*v1beta1.ObjectMeta, error) {
	labelsPath := append(m.path, "labels")
	labelsValue, ok, err := unstructured.NestedStringMap(obj.UnstructuredContent(), labelsPath...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve control plane metadata.labels")
	}
	if !ok {
		labelsValue = map[string]string{}
	}

	annotationsPath := append(m.path, "annotations")
	annotationsValue, ok, err := unstructured.NestedStringMap(obj.UnstructuredContent(), annotationsPath...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve control plane metadata.annotations")
	}
	if !ok {
		annotationsValue = map[string]string{}
	}

	return &v1beta1.ObjectMeta{
		Labels:      labelsValue,
		Annotations: annotationsValue,
	}, nil
}

// Set sets the metadata value.
// Note: We are blanking out empty label annotations, thus avoiding triggering infinite reconcile
// given that at json level label: {} or annotation: {} is different from no field, which is the
// corresponding value stored in etcd given that those fields are defined as omitempty.
func (m *Metadata) Set(obj *unstructured.Unstructured, metadata *v1beta1.ObjectMeta) error {
	labelsPath := append(m.path, "labels")
	unstructured.RemoveNestedField(obj.UnstructuredContent(), labelsPath...)
	if len(metadata.Labels) > 0 {
		if err := unstructured.SetNestedStringMap(obj.UnstructuredContent(), metadata.Labels, labelsPath...); err != nil {
			return errors.Wrap(err, "failed to set control plane metadata.labels")
		}
	}

	annotationsPath := append(m.path, "annotations")
	unstructured.RemoveNestedField(obj.UnstructuredContent(), annotationsPath...)
	if len(metadata.Annotations) > 0 {
		if err := unstructured.SetNestedStringMap(obj.UnstructuredContent(), metadata.Annotations, annotationsPath...); err != nil {
			return errors.Wrap(err, "failed to set control plane metadata.annotations")
		}
	}
	return nil
}
