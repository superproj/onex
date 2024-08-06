// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package validation

import (
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/superproj/onex/internal/pkg/util/validation"
	"github.com/superproj/onex/pkg/config"
)

// ValidateMySQLConfiguration ensures validation of the MySQLConfiguration struct.
func ValidateMySQLConfiguration(cc *config.MySQLConfiguration, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if cc.MaxIdleConnections <= 0 {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("maxIdleConnections"), cc.MaxIdleConnections, "must be greater than zero"))
	}
	if cc.MaxOpenConnections <= 0 {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("maxOpenConnections"), cc.MaxOpenConnections, "must be greater than zero"))
	}
	if cc.MaxConnectionLifeTime.Duration <= 0 {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("maxConnectionLifeTime"), cc.MaxConnectionLifeTime, "must be greater than zero"))
	}
	if len(cc.Host) == 0 {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("host"), cc.Host, "host is required"))
	}
	if len(cc.Database) == 0 {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("database"), cc.Database, "database is required"))
	}
	if len(cc.Password) == 0 {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("password"), cc.Password, "password is required"))
	}
	return allErrs
}

// ValidateGenericControllerManagerConfiguration ensures validation of the GenericControllerManagerConfiguration struct.
func ValidateGenericControllerManagerConfiguration(cc *config.GenericControllerManagerConfiguration, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if cc.HealthzBindAddress != "" {
		allErrs = append(allErrs, validation.ValidateHostPort(cc.HealthzBindAddress, fldPath.Child("healthzBindAddress"))...)
	}
	if cc.PprofBindAddress != "" {
		allErrs = append(allErrs, validation.ValidateHostPort(cc.PprofBindAddress, fldPath.Child("pprofBindAddress"))...)
	}
	if cc.MetricsBindAddress != "" {
		allErrs = append(allErrs, validation.ValidateHostPort(cc.MetricsBindAddress, fldPath.Child("metricsBindAddress"))...)
	}

	return allErrs
}
