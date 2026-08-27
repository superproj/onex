// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/onex.
//

package webhooks

import (
	"context"

	"github.com/distribution/reference"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	known "github.com/onexstack/onex/internal/pkg/known/controllermanager"
	"github.com/onexstack/onex/pkg/apis/apps/v1beta1"
)

// Chain implements a validation and defaulting webhook for Chain.
type Chain struct{}

var (
	_ admission.Defaulter[*v1beta1.Chain] = &Chain{}
	_ admission.Validator[*v1beta1.Chain] = &Chain{}
)

func (w *Chain) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &v1beta1.Chain{}).
		WithDefaulter(w).
		WithValidator(w).
		Complete()
}

// Default sets default Chain field values.
func (w *Chain) Default(_ context.Context, ch *v1beta1.Chain) error {
	if ch.Labels == nil {
		ch.Labels = make(map[string]string)
	}

	if ch.Spec.Image == "" {
		ch.Spec.Image = known.DefaultChainImage
	}

	return nil
}

// ValidateCreate implements admission.Validator so a webhook will be registered for the type.
func (w *Chain) ValidateCreate(_ context.Context, ch *v1beta1.Chain) (admission.Warnings, error) {
	return nil, w.validate(nil, ch)
}

// ValidateUpdate implements admission.Validator so a webhook will be registered for the type.
func (w *Chain) ValidateUpdate(_ context.Context, oldC, newC *v1beta1.Chain) (admission.Warnings, error) {
	return nil, w.validate(oldC, newC)
}

// ValidateDelete implements admission.Validator so a webhook will be registered for the type.
func (w *Chain) ValidateDelete(_ context.Context, _ *v1beta1.Chain) (admission.Warnings, error) {
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
