// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package v1beta1

import (
	"k8s.io/apimachinery/pkg/runtime"

	genericconfigv1beta1 "github.com/superproj/onex/pkg/config/v1beta1"
)

func addDefaultingFuncs(scheme *runtime.Scheme) error {
	return RegisterDefaults(scheme)
}

func SetDefaults_OneXControllerManagerConfiguration(obj *OneXControllerManagerConfiguration) {
	genericconfigv1beta1.RecommendedDefaultGenericControllerManagerConfiguration(&obj.Generic)
	genericconfigv1beta1.RecommendedDefaultGarbageCollectorControllerConfiguration(&obj.GarbageCollectorController)
	RecommendedDefaultChainControllerConfiguration(&obj.ChainController)
}

func RecommendedDefaultChainControllerConfiguration(obj *ChainControllerConfiguration) {
	if obj.Image == "" {
		obj.Image = "ccr.ccs.tencentyun.com/superproj/onex-toyblc-amd64:v0.1.0"
	}
}
