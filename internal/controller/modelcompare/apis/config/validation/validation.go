// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package validation

import (
	"k8s.io/apimachinery/pkg/util/validation/field"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	componentbasevalidation "k8s.io/component-base/config/validation"

	"github.com/superproj/onex/internal/controller/minerset/apis/config"
	"github.com/superproj/onex/internal/pkg/util/validation"
)

// Validate ensures validation of the MinerSetControllerConfiguration struct.
func Validate(cc *config.MinerSetControllerConfiguration) field.ErrorList {
	allErrs := field.ErrorList{}
	newPath := field.NewPath("MinerSetControllerConfiguration")

	effectiveFeatures := utilfeature.DefaultFeatureGate.DeepCopy()
	if err := effectiveFeatures.SetFromMap(cc.FeatureGates); err != nil {
		allErrs = append(allErrs, field.Invalid(newPath.Child("featureGates"), cc.FeatureGates, err.Error()))
	}
	allErrs = append(allErrs, componentbasevalidation.ValidateLeaderElectionConfiguration(&cc.LeaderElection, field.NewPath("leaderElection"))...)

	if cc.HealthzBindAddress != "" {
		allErrs = append(allErrs, validation.ValidateHostPort(cc.HealthzBindAddress, newPath.Child("healthzBindAddress"))...)
	}

	if cc.MetricsBindAddress != "" {
		allErrs = append(allErrs, validation.ValidateHostPort(cc.MetricsBindAddress, newPath.Child("metricsBindAddress"))...)
	}

	return allErrs
}
