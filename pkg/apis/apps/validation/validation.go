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

// ValidateMinerName validates that the given name can be used as a miner name.
var ValidateMinerName = apimachineryvalidation.NameIsDNSSubdomain

// ValidateMinerSetName validates that the given name can be used as a minerset name.
var ValidateMinerSetName = apimachineryvalidation.NameIsDNSSubdomain

// ValidateMiner validates a given Miner.
func ValidateMiner(obj *apps.Miner) field.ErrorList {
	allErrs := field.ErrorList{}
	return allErrs
}

// ValidateMinerSpec validates given miner spec.
func ValidateMinerSpec(spec *apps.MinerSpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	return allErrs
}

// ValidateMinerStatus validates given miner status.
func ValidateMinerStatus(status *apps.MinerStatus, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	return allErrs
}

// ValidateMinerUpdate tests if an update to a Miner is valid.
func ValidateMinerUpdate(update, old *apps.Miner) field.ErrorList {
	allErrs := field.ErrorList{}

	return allErrs
}

// ValidateMinerSpecUpdate tests if an update to a MinerSpec is valid.
func ValidateMinerSpecUpdate(newSpec, oldSpec *apps.MinerSpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	return allErrs
}

// ValidateMinerStatusUpdate tests if a an update to a Miner status
// is valid.
func ValidateMinerStatusUpdate(update, old *apps.Miner) field.ErrorList {
	allErrs := corevalidation.ValidateObjectMetaUpdate(&update.ObjectMeta, &old.ObjectMeta, field.NewPath("metadata"))
	fldPath := field.NewPath("status")
	allErrs = append(allErrs, ValidateMinerStatus(&update.Status, fldPath)...)
	return allErrs
}

// ValidateMinerSet validates a given MinerSet.
func ValidateMinerSet(obj *apps.MinerSet) field.ErrorList {
	allErrs := field.ErrorList{}
	return allErrs
}

// ValidateMinerSetSpec validates given minerset spec.
func ValidateMinerSetSpec(spec *apps.MinerSetSpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	return allErrs
}

// ValidateMinerSetStatus validates given minerset status.
func ValidateMinerSetStatus(status *apps.MinerSetStatus, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	return allErrs
}

// ValidateMinerSetUpdate tests if an update to a MinerSet is valid.
func ValidateMinerSetUpdate(update, old *apps.MinerSet) field.ErrorList {
	allErrs := field.ErrorList{}

	return allErrs
}

// ValidateMinerSetSpecUpdate tests if an update to a MinerSetSpec is valid.
func ValidateMinerSetSpecUpdate(newSpec, oldSpec *apps.MinerSetSpec, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	return allErrs
}

// ValidateMinerSetStatusUpdate tests if a an update to a MinerSet status
// is valid.
func ValidateMinerSetStatusUpdate(update, old *apps.MinerSet) field.ErrorList {
	allErrs := corevalidation.ValidateObjectMetaUpdate(&update.ObjectMeta, &old.ObjectMeta, field.NewPath("metadata"))
	fldPath := field.NewPath("status")
	allErrs = append(allErrs, ValidateMinerSetStatus(&update.Status, fldPath)...)
	return allErrs
}
