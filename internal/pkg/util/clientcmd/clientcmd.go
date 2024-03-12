// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package clientcmd

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/util/homedir"
)

const (
	RecommendedConfigPathFlag   = "kubeconfig"
	RecommendedConfigPathEnvVar = "KUBECONFIG"
	RecommendedHomeDir          = ".onex"
	RecommendedFileName         = "config"
)

var (
	RecommendedConfigDir = filepath.Join(homedir.HomeDir(), RecommendedHomeDir)
	RecommendedHomeFile  = filepath.Join(RecommendedConfigDir, RecommendedFileName)
)

func DefaultKubeconfig() string {
	defaultKubeconfig := os.Getenv(RecommendedConfigPathEnvVar)
	if defaultKubeconfig == "" {
		defaultKubeconfig = RecommendedHomeFile
	}

	return defaultKubeconfig
}
