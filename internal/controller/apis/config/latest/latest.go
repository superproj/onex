// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package latest

import (
	"github.com/superproj/onex/internal/controller/apis/config"
	"github.com/superproj/onex/internal/controller/apis/config/scheme"
	"github.com/superproj/onex/internal/controller/apis/config/v1beta1"
)

// Default creates a default configuration of the latest versioned type.
// This function needs to be updated whenever we bump the miner controller's component config version.
func Default() (*config.OneXControllerManagerConfiguration, error) {
	versioned := v1beta1.OneXControllerManagerConfiguration{}

	scheme.Scheme.Default(&versioned)
	cfg := config.OneXControllerManagerConfiguration{}
	if err := scheme.Scheme.Convert(&versioned, &cfg, nil); err != nil {
		return nil, err
	}

	// We don't set this field in internal/controller/apis/config/{version}/conversion.go
	// because the field will be cleared later by API machinery during
	// conversion. See MinerControllerConfiguration internal type definition for
	// more details.
	cfg.TypeMeta.APIVersion = v1beta1.SchemeGroupVersion.String()
	return &cfg, nil
}
