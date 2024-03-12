// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package validation

import (
	apimachineryvalidation "k8s.io/apimachinery/pkg/api/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	corevalidation "k8s.io/kubernetes/pkg/apis/core/validation"

	"github.com/superproj/onex/pkg/apis/apps"
)

// ValidateChainName validates that the given name can be used as a chain name.
var ValidateChainName = apimachineryvalidation.NameIsDNSSubdomain

// ValidateChain validates a given Chain.
func ValidateChain(obj *apps.Chain) field.ErrorList {
	allErrs := field.ErrorList{}

	allErrs = append(allErrs, corevalidation.ValidateObjectMeta(&obj.ObjectMeta, true, ValidateChainName, field.NewPath("metadata"))...)

	if obj.ObjectMeta.Namespace != metav1.NamespaceSystem {
		allErrs = append(allErrs, field.Forbidden(field.NewPath("metadata", "namespace"), "must be set to `kube-system`"))
	}

	return allErrs
}

// ValidateChainSpec validates given chain spec.
func ValidateChainSpec(spec *apps.ChainSpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	return allErrs
}

// ValidateChainUpdate tests if an update to a Chain is valid.
func ValidateChainUpdate(update, old *apps.Chain) field.ErrorList {
	allErrs := field.ErrorList{}

	return allErrs
}

// ValidateChainSpecUpdate tests if an update to a ChainSpec is valid.
func ValidateChainSpecUpdate(newSpec, oldSpec *apps.ChainSpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	return allErrs
}

// ValidateChainStatus validates given chain status.
func ValidateChainStatus(status *apps.ChainStatus, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	return allErrs
}

// ValidateChainStatusUpdate tests if a an update to a Chain status
// is valid.
func ValidateChainStatusUpdate(update, old *apps.Chain) field.ErrorList {
	allErrs := corevalidation.ValidateObjectMetaUpdate(&update.ObjectMeta, &old.ObjectMeta, field.NewPath("metadata"))
	fldPath := field.NewPath("status")
	allErrs = append(allErrs, ValidateChainStatus(&update.Status, fldPath)...)
	return allErrs
}
