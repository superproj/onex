// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package contract

import (
	"testing"

	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var fooRefBuilder = func() *unstructured.Unstructured {
	refObj := &unstructured.Unstructured{}
	refObj.SetAPIVersion("fooApiVersion")
	refObj.SetKind("fooKind")
	refObj.SetNamespace("fooNamespace")
	refObj.SetName("fooName")
	return refObj
}

func TestGetNestedRef(t *testing.T) {
	t.Run("Gets a nested ref if defined", func(t *testing.T) {
		g := NewWithT(t)

		refObj := fooRefBuilder()
		obj := &unstructured.Unstructured{Object: map[string]any{}}

		err := SetNestedRef(obj, refObj, "spec", "machineTemplate", "infrastructureRef")
		g.Expect(err).To(BeNil())

		ref, err := GetNestedRef(obj, "spec", "machineTemplate", "infrastructureRef")
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(ref).ToNot(BeNil())
		g.Expect(ref.APIVersion).To(Equal(refObj.GetAPIVersion()))
		g.Expect(ref.Kind).To(Equal(refObj.GetKind()))
		g.Expect(ref.Name).To(Equal(refObj.GetName()))
		g.Expect(ref.Namespace).To(Equal(refObj.GetNamespace()))
	})
	t.Run("getNestedRef fails if the nested ref does not exist", func(t *testing.T) {
		g := NewWithT(t)

		obj := &unstructured.Unstructured{Object: map[string]any{}}

		ref, err := GetNestedRef(obj, "spec", "machineTemplate", "infrastructureRef")
		g.Expect(err).To(HaveOccurred())
		g.Expect(ref).To(BeNil())
	})
	t.Run("getNestedRef fails if the nested ref exist but it is incomplete", func(t *testing.T) {
		g := NewWithT(t)

		obj := &unstructured.Unstructured{Object: map[string]any{}}

		err := unstructured.SetNestedField(obj.UnstructuredContent(), "foo", "spec", "machineTemplate", "infrastructureRef", "kind")
		g.Expect(err).ToNot(HaveOccurred())
		err = unstructured.SetNestedField(obj.UnstructuredContent(), "bar", "spec", "machineTemplate", "infrastructureRef", "namespace")
		g.Expect(err).ToNot(HaveOccurred())
		err = unstructured.SetNestedField(obj.UnstructuredContent(), "baz", "spec", "machineTemplate", "infrastructureRef", "apiVersion")
		g.Expect(err).ToNot(HaveOccurred())
		// Reference name missing

		ref, err := GetNestedRef(obj, "spec", "machineTemplate", "infrastructureRef")
		g.Expect(err).To(HaveOccurred())
		g.Expect(ref).To(BeNil())
	})
}

func TestSetNestedRef(t *testing.T) {
	t.Run("Sets a nested ref", func(t *testing.T) {
		g := NewWithT(t)

		refObj := fooRefBuilder()
		obj := &unstructured.Unstructured{Object: map[string]any{}}

		err := SetNestedRef(obj, refObj, "spec", "machineTemplate", "infrastructureRef")
		g.Expect(err).To(BeNil())

		ref, err := GetNestedRef(obj, "spec", "machineTemplate", "infrastructureRef")
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(ref).ToNot(BeNil())
		g.Expect(ref.APIVersion).To(Equal(refObj.GetAPIVersion()))
		g.Expect(ref.Kind).To(Equal(refObj.GetKind()))
		g.Expect(ref.Name).To(Equal(refObj.GetName()))
		g.Expect(ref.Namespace).To(Equal(refObj.GetNamespace()))
	})
}

func TestObjToRef(t *testing.T) {
	t.Run("Gets a ref from an obj", func(t *testing.T) {
		g := NewWithT(t)

		refObj := fooRefBuilder()
		ref := ObjToRef(refObj)

		g.Expect(ref).ToNot(BeNil())
		g.Expect(ref.APIVersion).To(Equal(refObj.GetAPIVersion()))
		g.Expect(ref.Kind).To(Equal(refObj.GetKind()))
		g.Expect(ref.Name).To(Equal(refObj.GetName()))
		g.Expect(ref.Namespace).To(Equal(refObj.GetNamespace()))
	})
}
