// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

// Package ssa contains utils related to ssa.
package ssa

import (
	"context"
	"encoding/json"
	"testing"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/superproj/onex/internal/pkg/contract"
)

func TestDropManagedFields(t *testing.T) {
	ctx := context.Background()

	ssaManager := "ssa-manager"

	fieldV1Map := map[string]any{
		"f:metadata": map[string]any{
			"f:name":        map[string]any{},
			"f:labels":      map[string]any{},
			"f:annotations": map[string]any{},
			"f:finalizers":  map[string]any{},
		},
	}
	fieldV1, err := json.Marshal(fieldV1Map)
	if err != nil {
		panic(err)
	}

	objWithoutSSAManager := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cm-1",
			Namespace: "default",
			ManagedFields: []metav1.ManagedFieldsEntry{{
				Manager:    classicManager,
				Operation:  metav1.ManagedFieldsOperationUpdate,
				FieldsType: "FieldsV1",
				FieldsV1:   &metav1.FieldsV1{Raw: fieldV1},
			}},
			Labels: map[string]string{
				"label-1": "value-1",
			},
			Annotations: map[string]string{
				"annotation-1": "value-1",
			},
			Finalizers: []string{"test-finalizer"},
		},
		Data: map[string]string{
			"test-key": "test-value",
		},
	}

	objectWithSSAManager := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cm-2",
			Namespace: "default",
			ManagedFields: []metav1.ManagedFieldsEntry{
				{
					Manager:    classicManager,
					Operation:  metav1.ManagedFieldsOperationUpdate,
					FieldsType: "FieldsV1",
					FieldsV1:   &metav1.FieldsV1{Raw: fieldV1},
				},
				{
					Manager:   ssaManager,
					Operation: metav1.ManagedFieldsOperationApply,
				},
			},
			Labels: map[string]string{
				"label-1": "value-1",
			},
			Annotations: map[string]string{
				"annotation-1": "value-1",
			},
			Finalizers: []string{"test-finalizer"},
		},
		Data: map[string]string{
			"test-key": "test-value",
		},
	}

	tests := []struct {
		name                string
		obj                 client.Object
		wantOwnershipToDrop bool
	}{
		{
			name:                "should drop ownership of fields if there is no entry for ssaManager",
			obj:                 objWithoutSSAManager,
			wantOwnershipToDrop: true,
		},
		{
			name:                "should not drop ownership of fields if there is an entry for ssaManager",
			obj:                 objectWithSSAManager,
			wantOwnershipToDrop: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			fakeClient := fake.NewClientBuilder().WithObjects(tt.obj).Build()
			labelsAndAnnotationsManagedFieldPaths := []contract.Path{
				{"f:metadata", "f:annotations"},
				{"f:metadata", "f:labels"},
			}
			g.Expect(DropManagedFields(ctx, fakeClient, tt.obj, ssaManager, labelsAndAnnotationsManagedFieldPaths)).To(Succeed())
			if tt.wantOwnershipToDrop {
				g.Expect(tt.obj.GetManagedFields()).ShouldNot(MatchFieldOwnership(
					classicManager,
					metav1.ManagedFieldsOperationUpdate,
					contract.Path{"f:metadata", "f:labels"},
				))
				g.Expect(tt.obj.GetManagedFields()).ShouldNot(MatchFieldOwnership(
					classicManager,
					metav1.ManagedFieldsOperationUpdate,
					contract.Path{"f:metadata", "f:annotations"},
				))
			} else {
				g.Expect(tt.obj.GetManagedFields()).Should(MatchFieldOwnership(
					classicManager,
					metav1.ManagedFieldsOperationUpdate,
					contract.Path{"f:metadata", "f:labels"},
				))
				g.Expect(tt.obj.GetManagedFields()).Should(MatchFieldOwnership(
					classicManager,
					metav1.ManagedFieldsOperationUpdate,
					contract.Path{"f:metadata", "f:annotations"},
				))
			}
			// Verify ownership of other fields is not affected.
			g.Expect(tt.obj.GetManagedFields()).Should(MatchFieldOwnership(
				classicManager,
				metav1.ManagedFieldsOperationUpdate,
				contract.Path{"f:metadata", "f:finalizers"},
			))
		})
	}
}

func TestCleanUpManagedFieldsForSSAAdoption(t *testing.T) {
	ctx := context.Background()

	ssaManager := "ssa-manager"

	objWithoutAnyManager := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cm-1",
			Namespace: "default",
		},
		Data: map[string]string{
			"test-key": "test-value",
		},
	}
	objWithOnlyClassicManager := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cm-1",
			Namespace: "default",
			ManagedFields: []metav1.ManagedFieldsEntry{{
				Manager:   classicManager,
				Operation: metav1.ManagedFieldsOperationUpdate,
			}},
		},
		Data: map[string]string{
			"test-key": "test-value",
		},
	}
	objWithOnlySSAManager := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cm-1",
			Namespace: "default",
			ManagedFields: []metav1.ManagedFieldsEntry{{
				Manager:   ssaManager,
				Operation: metav1.ManagedFieldsOperationApply,
			}},
		},
		Data: map[string]string{
			"test-key": "test-value",
		},
	}
	objWithClassicManagerAndSSAManager := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cm-1",
			Namespace: "default",
			ManagedFields: []metav1.ManagedFieldsEntry{
				{
					Manager:   classicManager,
					Operation: metav1.ManagedFieldsOperationUpdate,
				},
				{
					Manager:   ssaManager,
					Operation: metav1.ManagedFieldsOperationApply,
				},
			},
		},
		Data: map[string]string{
			"test-key": "test-value",
		},
	}

	tests := []struct {
		name                       string
		obj                        client.Object
		wantEntryForClassicManager bool
	}{
		{
			name:                       "should add an entry for ssaManager if it does not have one",
			obj:                        objWithoutAnyManager,
			wantEntryForClassicManager: false,
		},
		{
			name:                       "should add an entry for ssaManager and drop entry for classic manager if it exists",
			obj:                        objWithOnlyClassicManager,
			wantEntryForClassicManager: false,
		},
		{
			name:                       "should keep the entry for ssa-manager if it already has one (no-op)",
			obj:                        objWithOnlySSAManager,
			wantEntryForClassicManager: false,
		},
		{
			name:                       "should keep the entry with ssa-manager if it already has one - should not drop other manager entries (no-op)",
			obj:                        objWithClassicManagerAndSSAManager,
			wantEntryForClassicManager: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			fakeClient := fake.NewClientBuilder().WithObjects(tt.obj).Build()
			g.Expect(CleanUpManagedFieldsForSSAAdoption(ctx, fakeClient, tt.obj, ssaManager)).Should(Succeed())
			g.Expect(tt.obj.GetManagedFields()).Should(
				ContainElement(MatchManagedFieldsEntry(ssaManager, metav1.ManagedFieldsOperationApply)))
			if tt.wantEntryForClassicManager {
				g.Expect(tt.obj.GetManagedFields()).Should(
					ContainElement(MatchManagedFieldsEntry(classicManager, metav1.ManagedFieldsOperationUpdate)))
			} else {
				g.Expect(tt.obj.GetManagedFields()).ShouldNot(
					ContainElement(MatchManagedFieldsEntry(classicManager, metav1.ManagedFieldsOperationUpdate)))
			}
		})
	}
}
