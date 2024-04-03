// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package options

import (
	"github.com/spf13/pflag"

	// ensure libs have a chance to globally register their flags.
	_ "k8s.io/apiserver/pkg/admission"
)

// AddCustomGlobalFlags explicitly registers flags that internal packages register
// against the global flagsets from "flag". We do this in order to prevent
// unwanted flags from leaking into the kube-apiserver's flagset.
func AddCustomGlobalFlags(fs *pflag.FlagSet) {
	// Lookup flags in global flag set and re-register the values with our flagset.

	// Adds flags from k8s.io/apiserver/pkg/admission.
}
