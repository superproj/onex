// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package latest

import (
	"github.com/superproj/onex/internal/controller/minerset/apis/config"
	"github.com/superproj/onex/internal/controller/minerset/apis/config/scheme"
	"github.com/superproj/onex/internal/controller/minerset/apis/config/v1beta1"
)

// Default creates a default configuration of the latest versioned type.
// This function needs to be updated whenever we bump the minerset controller's component config version.
func Default() (*config.MinerSetControllerConfiguration, error) {
	versionedCfg := v1beta1.MinerSetControllerConfiguration{}

	scheme.Scheme.Default(&versionedCfg)
	cfg := config.MinerSetControllerConfiguration{}
	if err := scheme.Scheme.Convert(&versionedCfg, &cfg, nil); err != nil {
		return nil, err
	}
	// We don't set this field in internal/controller/minerset/apis/config/{version}/conversion.go
	// because the field will be cleared later by API machinery during
	// conversion. See MinerSetControllerConfiguration internal type definition for
	// more details.
	cfg.TypeMeta.APIVersion = v1beta1.SchemeGroupVersion.String()
	return &cfg, nil
}
