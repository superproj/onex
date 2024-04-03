// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package conditions

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"

	coreutil "github.com/superproj/onex/internal/pkg/util/core"
	"github.com/superproj/onex/pkg/apis/apps/v1beta1"
)

// UnstructuredGetter return a Getter object that can read conditions from an Unstructured object.
// Important. This method should be used only with types implementing Cluster API conditions.
func UnstructuredGetter(u *unstructured.Unstructured) Getter {
	return &unstructuredWrapper{Unstructured: u}
}

// UnstructuredSetter return a Setter object that can set conditions from an Unstructured object.
// Important. This method should be used only with types implementing Cluster API conditions.
func UnstructuredSetter(u *unstructured.Unstructured) Setter {
	return &unstructuredWrapper{Unstructured: u}
}

type unstructuredWrapper struct {
	*unstructured.Unstructured
}

// GetConditions returns the list of conditions from an Unstructured object.
//
// NOTE: Due to the constraints of JSON-unmarshal, this operation is to be considered best effort.
// In more details:
//   - Errors during JSON-unmarshal are ignored and a empty collection list is returned.
//   - It's not possible to detect if the object has an empty condition list or if it does not implement conditions;
//     in both cases the operation returns an empty slice is returned.
//   - If the object doesn't implement conditions on under status as defined in Cluster API,
//     JSON-unmarshal matches incoming object keys to the keys; this can lead to to conditions values partially set.
func (c *unstructuredWrapper) GetConditions() v1beta1.Conditions {
	conditions := v1beta1.Conditions{}
	if err := coreutil.UnstructuredUnmarshalField(c.Unstructured, &conditions, "status", "conditions"); err != nil {
		return nil
	}
	return conditions
}

// SetConditions set the conditions into an Unstructured object.
//
// NOTE: Due to the constraints of JSON-unmarshal, this operation is to be considered best effort.
// In more details:
//   - Errors during JSON-unmarshal are ignored and a empty collection list is returned.
//   - It's not possible to detect if the object has an empty condition list or if it does not implement conditions;
//     in both cases the operation returns an empty slice is returned.
func (c *unstructuredWrapper) SetConditions(conditions v1beta1.Conditions) {
	v := make([]any, 0, len(conditions))
	for i := range conditions {
		m, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&conditions[i])
		if err != nil {
			log.Log.Error(err, "Failed to convert Condition to unstructured map. This error shouldn't have occurred, please file an issue.", "groupVersionKind", c.GroupVersionKind(), "name", c.GetName(), "namespace", c.GetNamespace())
			continue
		}
		v = append(v, m)
	}
	// unstructured.SetNestedField returns an error only if value cannot be set because one of
	// the nesting levels is not a map[string]any; this is not the case so the error should never happen here.
	err := unstructured.SetNestedField(c.Unstructured.Object, v, "status", "conditions")
	if err != nil {
		log.Log.Error(err, "Failed to set Conditions on unstructured object. This error shouldn't have occurred, please file an issue.", "groupVersionKind", c.GroupVersionKind(), "name", c.GetName(), "namespace", c.GetNamespace())
	}
}
