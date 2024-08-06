// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package validation

import (
	"k8s.io/apimachinery/pkg/util/validation/field"
	componentbasevalidation "k8s.io/component-base/config/validation"

	"github.com/superproj/onex/internal/controller/apis/config"
	"github.com/superproj/onex/internal/pkg/util/validation"
	genericvalidation "github.com/superproj/onex/pkg/config/validation"
)

// Validate ensures validation of the MinerControllerConfiguration struct.
func Validate(cc *config.OneXControllerManagerConfiguration) field.ErrorList {
	allErrs := field.ErrorList{}

	newPath := field.NewPath("OneXControllerManagerConfiguration")

	allErrs = append(allErrs, componentbasevalidation.ValidateLeaderElectionConfiguration(&cc.Generic.LeaderElection, field.NewPath("generic", "leaderElection"))...)
	allErrs = append(allErrs, genericvalidation.ValidateMySQLConfiguration(&cc.MySQL, field.NewPath("mysql"))...)
	allErrs = append(allErrs, genericvalidation.ValidateGenericControllerManagerConfiguration(&cc.Generic, field.NewPath("generic"))...)

	return allErrs
}
