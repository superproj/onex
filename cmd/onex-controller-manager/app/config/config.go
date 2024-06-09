// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package config

import (
	restclient "k8s.io/client-go/rest"

	ctrlmgrconfig "github.com/superproj/onex/internal/controller/apis/config"
	clientset "github.com/superproj/onex/pkg/generated/clientset/versioned"
)

// Config is the main context object for the controller.
type Config struct {
	ComponentConfig *ctrlmgrconfig.OneXControllerManagerConfiguration

	// the general onex client
	Client clientset.Interface

	// the rest config for the master
	Kubeconfig *restclient.Config
}

// CompletedConfig same as Config, just to swap private object.
type CompletedConfig struct {
	*Config
}

// Complete fills in any fields not set that are required to have valid data. It's mutating the receiver.
func (c *Config) Complete() *CompletedConfig {
	return &CompletedConfig{c}
}
