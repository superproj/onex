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

// ValidateModelCompareName validates that the given name can be used as a chain name.
var ValidateModelCompareName = apimachineryvalidation.NameIsDNSSubdomain

// ValidateModelCompare validates a given ModelCompare.
func ValidateModelCompare(obj *apps.ModelCompare) field.ErrorList {
	allErrs := field.ErrorList{}

	allErrs = append(allErrs, corevalidation.ValidateObjectMeta(&obj.ObjectMeta, true, ValidateModelCompareName, field.NewPath("metadata"))...)

	return allErrs
}

// ValidateModelCompareSpec validates given chain spec.
func ValidateModelCompareSpec(spec *apps.ModelCompareSpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	return allErrs
}

// ValidateModelCompareUpdate tests if an update to a ModelCompare is valid.
func ValidateModelCompareUpdate(update, old *apps.ModelCompare) field.ErrorList {
	allErrs := field.ErrorList{}

	return allErrs
}

// ValidateModelCompareSpecUpdate tests if an update to a ModelCompareSpec is valid.
func ValidateModelCompareSpecUpdate(newSpec, oldSpec *apps.ModelCompareSpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	return allErrs
}

// ValidateModelCompareStatus validates given chain status.
func ValidateModelCompareStatus(status *apps.ModelCompareStatus, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	return allErrs
}

// ValidateModelCompareStatusUpdate tests if a an update to a ModelCompare status
// is valid.
func ValidateModelCompareStatusUpdate(update, old *apps.ModelCompare) field.ErrorList {
	allErrs := corevalidation.ValidateObjectMetaUpdate(&update.ObjectMeta, &old.ObjectMeta, field.NewPath("metadata"))
	fldPath := field.NewPath("status")
	allErrs = append(allErrs, ValidateModelCompareStatus(&update.Status, fldPath)...)
	return allErrs
}
