// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package server

import (
	genericoptions "github.com/superproj/onex/pkg/options"
)

// Config represents the configuration options for a service, including HTTP, GRPC, and TLS settings.
type Config struct {
	HTTP genericoptions.HTTPOptions
	GRPC genericoptions.GRPCOptions
	TLS  genericoptions.TLSOptions
}
