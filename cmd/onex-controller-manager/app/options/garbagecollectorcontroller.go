// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package options

import (
	"github.com/spf13/pflag"

	garbagecollectorconfig "github.com/superproj/onex/internal/controller/apis/config"
)

// GarbageCollectorControllerOptions holds the GarbageCollectorController options.
type GarbageCollectorControllerOptions struct {
	*garbagecollectorconfig.GarbageCollectorControllerConfiguration
}

// AddFlags adds flags related to GarbageCollectorController for controller manager to the specified FlagSet.
func (o *GarbageCollectorControllerOptions) AddFlags(fs *pflag.FlagSet) {
	if o == nil {
		return
	}

	fs.Int32Var(&o.ConcurrentGCSyncs, "concurrent-gc-syncs", o.ConcurrentGCSyncs, "The number of garbage collector workers that are allowed to sync concurrently."+
		"This parameter is ignored if a config file is specified by --config.")
	fs.BoolVar(&o.EnableGarbageCollector, "enable-garbage-collector", o.EnableGarbageCollector, "Enables the generic garbage collector. MUST be synced with "+
		"the corresponding flag of the kube-apiserver. This parameter is ignored if a config file is specified by --config.")
}
