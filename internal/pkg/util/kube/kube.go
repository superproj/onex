// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package kube

import (
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
)

func SetClientOptionsForMinerController(config *rest.Config) *rest.Config {
	return SetDefaultClientOptions(config)
}

func SetClientOptionsForLifecycleController(config *rest.Config) *rest.Config {
	return SetDefaultClientOptions(config)
}

func SetClientOptionsForMinerSetController(config *rest.Config) *rest.Config {
	return SetDefaultClientOptions(config)
}

func SetClientOptionsForController(config *rest.Config) *rest.Config {
	return SetDefaultClientOptions(config)
}

func SetDefaultClientOptions(config *rest.Config) *rest.Config {
	config.DisableCompression = true
	config.QPS = float32(2000)
	config.Burst = 4000
	// Set ContentType to application/json, otherwise configmap will report
	// `the body of the request was in an unknown format - accepted media types include: application/json, application/yaml` error
	config.ContentType = runtime.ContentTypeJSON

	return config
}

func IsLocalEnv() bool {
	return os.Getenv("NODEAPIENV") == "local"
}
