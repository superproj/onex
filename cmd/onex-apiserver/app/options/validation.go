// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package options

// Validate checks ServerRunOptions and return a slice of found errs.
func (o CompletedOptions) Validate() []error {
	errs := []error{}
	errs = append(errs, o.RecommendedOptions.Validate()...)
	errs = append(errs, o.GenericServerRunOptions.Validate()...)
	errs = append(errs, o.Metrics.Validate()...)
	// errs = append(errs, o.CloudOptions.Validate()...)

	return errs
}
