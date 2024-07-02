// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package app

import (
	cliflag "k8s.io/component-base/cli/flag"
)

// CliOptions abstracts configuration options for reading parameters from the
// command line.
type CliOptions interface {
	// Flags returns flags for a specific server by section name.
	Flags() cliflag.NamedFlagSets

	// Complete completes all the required options.
	Complete() error

	// Validate validates all the required options.
	Validate() error
}
