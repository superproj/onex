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

func TestBootstrap(t *testing.T) {
	obj := &unstructured.Unstructured{Object: map[string]any{}}

	t.Run("Manages status.ready", func(t *testing.T) {
		g := NewWithT(t)

		g.Expect(Bootstrap().Ready().Path()).To(Equal(Path{"status", "ready"}))

		err := Bootstrap().Ready().Set(obj, true)
		g.Expect(err).ToNot(HaveOccurred())

		got, err := Bootstrap().Ready().Get(obj)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(got).ToNot(BeNil())
		g.Expect(*got).To(BeTrue())
	})
	t.Run("Manages status.dataSecretName", func(t *testing.T) {
		g := NewWithT(t)

		g.Expect(Bootstrap().DataSecretName().Path()).To(Equal(Path{"status", "dataSecretName"}))

		err := Bootstrap().DataSecretName().Set(obj, "fake-data-secret-name")
		g.Expect(err).ToNot(HaveOccurred())

		got, err := Bootstrap().DataSecretName().Get(obj)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(got).ToNot(BeNil())
		g.Expect(*got).To(Equal("fake-data-secret-name"))
	})
	t.Run("Manages optional status.failureReason", func(t *testing.T) {
		g := NewWithT(t)

		g.Expect(Bootstrap().FailureReason().Path()).To(Equal(Path{"status", "failureReason"}))

		err := Bootstrap().FailureReason().Set(obj, "fake-reason")
		g.Expect(err).ToNot(HaveOccurred())

		got, err := Bootstrap().FailureReason().Get(obj)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(got).ToNot(BeNil())
		g.Expect(*got).To(Equal("fake-reason"))
	})
	t.Run("Manages optional status.failureMessage", func(t *testing.T) {
		g := NewWithT(t)

		g.Expect(Bootstrap().FailureMessage().Path()).To(Equal(Path{"status", "failureMessage"}))

		err := Bootstrap().FailureMessage().Set(obj, "fake-message")
		g.Expect(err).ToNot(HaveOccurred())

		got, err := Bootstrap().FailureMessage().Get(obj)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(got).ToNot(BeNil())
		g.Expect(*got).To(Equal("fake-message"))
	})
}
