// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package validation

import (
	apimachineryvalidation "k8s.io/apimachinery/pkg/api/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	corevalidation "k8s.io/kubernetes/pkg/apis/core/validation"

	"github.com/superproj/onex/pkg/apis/apps"
)

// ValidateEvaluateName validates that the given name can be used as a chain name.
var ValidateEvaluateName = apimachineryvalidation.NameIsDNSSubdomain

// ValidateEvaluate validates a given Evaluate.
func ValidateEvaluate(obj *apps.Evaluate) field.ErrorList {
	allErrs := field.ErrorList{}

	allErrs = append(allErrs, corevalidation.ValidateObjectMeta(&obj.ObjectMeta, true, ValidateEvaluateName, field.NewPath("metadata"))...)

	return allErrs
}

// ValidateEvaluateSpec validates given chain spec.
func ValidateEvaluateSpec(spec *apps.EvaluateSpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	return allErrs
}

// ValidateEvaluateUpdate tests if an update to a Evaluate is valid.
func ValidateEvaluateUpdate(update, old *apps.Evaluate) field.ErrorList {
	allErrs := field.ErrorList{}

	return allErrs
}

// ValidateEvaluateSpecUpdate tests if an update to a EvaluateSpec is valid.
func ValidateEvaluateSpecUpdate(newSpec, oldSpec *apps.EvaluateSpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	return allErrs
}

// ValidateEvaluateStatus validates given chain status.
func ValidateEvaluateStatus(status *apps.EvaluateStatus, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	return allErrs
}

// ValidateEvaluateStatusUpdate tests if a an update to a Evaluate status
// is valid.
func ValidateEvaluateStatusUpdate(update, old *apps.Evaluate) field.ErrorList {
	allErrs := corevalidation.ValidateObjectMetaUpdate(&update.ObjectMeta, &old.ObjectMeta, field.NewPath("metadata"))
	fldPath := field.NewPath("status")
	allErrs = append(allErrs, ValidateEvaluateStatus(&update.Status, fldPath)...)
	return allErrs
}
