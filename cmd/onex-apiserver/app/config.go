// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package app

import (
	"github.com/superproj/onex/cmd/onex-apiserver/app/options"
	"github.com/superproj/onex/internal/apiserver"
)

// NewConfig creates all the resources for running kube-apiserver, but runs none of them.
// onex-apiserver has no extension and aggregator apiserver, so return *apiserver.Config directly.
// If you want to add extension and aggregator apiserver in the future, please refer to
// https://github.com/kubernetes/kubernetes/blob/v1.28.2/cmd/kube-apiserver/app/config.go#L28.
func NewConfig(opts options.CompletedOptions) (*apiserver.Config, error) {
	return CreateOneXAPIServerConfig(opts)
}
