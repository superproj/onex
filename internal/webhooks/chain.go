// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package webhooks

import (
	"context"
	"fmt"

	"github.com/distribution/reference"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	known "github.com/superproj/onex/internal/pkg/known/controllermanager"
	"github.com/superproj/onex/pkg/apis/apps/v1beta1"
)

// Chain implements a validation and defaulting webhook for Chain.
type Chain struct{}

var (
	_ webhook.CustomDefaulter = &Chain{}
	_ webhook.CustomValidator = &Chain{}
)

func (w *Chain) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&v1beta1.Chain{}).
		WithDefaulter(w).
		WithValidator(w).
		Complete()
}

// Default sets default Chain field values.
func (w *Chain) Default(_ context.Context, obj runtime.Object) error {
	ch, ok := obj.(*v1beta1.Chain)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a Chain but got a %T", obj))
	}

	if ch.Labels == nil {
		ch.Labels = make(map[string]string)
	}

	if ch.Spec.Image == "" {
		ch.Spec.Image = known.DefaultChainImage
	}

	return nil
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type.
func (w *Chain) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	ch, ok := obj.(*v1beta1.Chain)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a Chain but got a %T", obj))
	}

	return nil, w.validate(nil, ch)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (w *Chain) ValidateUpdate(_ context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	oldC, ok := oldObj.(*v1beta1.Chain)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a Chain but got a %T", oldObj))
	}
	newC, ok := newObj.(*v1beta1.Chain)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a Chain but got a %T", newObj))
	}

	return nil, w.validate(oldC, newC)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type.
func (w *Chain) ValidateDelete(_ context.Context, _ runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (w *Chain) validate(_, newC *v1beta1.Chain) error {
	var allErrs field.ErrorList

	specPath := field.NewPath("spec")

	if !reference.ReferenceRegexp.MatchString(newC.Spec.Image) {
		allErrs = append(allErrs, field.Invalid(specPath.Child("image"), newC.Spec.Image, "invalid image repository format"))
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(v1beta1.SchemeGroupVersion.WithKind("Chain").GroupKind(), newC.Name, allErrs)
}
