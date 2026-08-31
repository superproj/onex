// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/onex.
//

package v1beta1

// OpenAPIModelName returns the OpenAPI model name for this type.
func (c Chain) OpenAPIModelName() string {
	return "github.com/onexstack/onex/pkg/apis/apps/v1beta1.Chain"
}

// OpenAPIModelName returns the OpenAPI model name for this type.
func (c ChainList) OpenAPIModelName() string {
	return "github.com/onexstack/onex/pkg/apis/apps/v1beta1.ChainList"
}

// OpenAPIModelName returns the OpenAPI model name for this type.
func (m Miner) OpenAPIModelName() string {
	return "github.com/onexstack/onex/pkg/apis/apps/v1beta1.Miner"
}

// OpenAPIModelName returns the OpenAPI model name for this type.
func (m MinerList) OpenAPIModelName() string {
	return "github.com/onexstack/onex/pkg/apis/apps/v1beta1.MinerList"
}

// OpenAPIModelName returns the OpenAPI model name for this type.
func (m MinerSet) OpenAPIModelName() string {
	return "github.com/onexstack/onex/pkg/apis/apps/v1beta1.MinerSet"
}

// OpenAPIModelName returns the OpenAPI model name for this type.
func (m MinerSetList) OpenAPIModelName() string {
	return "github.com/onexstack/onex/pkg/apis/apps/v1beta1.MinerSetList"
}
