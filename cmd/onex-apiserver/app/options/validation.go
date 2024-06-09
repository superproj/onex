// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package options

import "fmt"

// Validate checks ServerRunOptions and return a slice of found errs.
func (o CompletedOptions) Validate() []error {
	errs := []error{}
	errs = append(errs, o.CompletedOptions.Validate()...)
	//errs = append(errs, s.CloudProvider.Validate()...)

	if o.MasterCount <= 0 {
		errs = append(errs, fmt.Errorf("--apiserver-count should be a positive number, but value '%d' provided", o.MasterCount))
	}

	return errs
}
