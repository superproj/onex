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

	"github.com/superproj/onex/pkg/apis/apps/v1beta1"
)

func TestMetadata(t *testing.T) {
	obj := &unstructured.Unstructured{Object: map[string]any{}}

	t.Run("Manages metadata", func(t *testing.T) {
		g := NewWithT(t)

		metadata := &v1beta1.ObjectMeta{
			Labels: map[string]string{
				"label1": "labelValue1",
			},
			Annotations: map[string]string{
				"annotation1": "annotationValue1",
			},
		}

		m := Metadata{path: Path{"foo"}}
		g.Expect(m.Path()).To(Equal(Path{"foo"}))

		err := m.Set(obj, metadata)
		g.Expect(err).ToNot(HaveOccurred())

		got, err := m.Get(obj)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(got).ToNot(BeNil())
		g.Expect(got).To(Equal(metadata))
	})
	t.Run("Manages empty metadata", func(t *testing.T) {
		g := NewWithT(t)

		metadata := &v1beta1.ObjectMeta{
			Labels:      map[string]string{},
			Annotations: map[string]string{},
		}

		m := Metadata{path: Path{"foo"}}
		g.Expect(m.Path()).To(Equal(Path{"foo"}))

		err := m.Set(obj, metadata)
		g.Expect(err).ToNot(HaveOccurred())

		got, err := m.Get(obj)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(got).ToNot(BeNil())
		g.Expect(got).To(Equal(metadata))
	})
}
