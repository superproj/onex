// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package apiserver

import (
	// These imports are the API groups the API server will support.
	_ "k8s.io/kubernetes/pkg/apis/autoscaling/install"

	_ "github.com/superproj/onex/pkg/apis/apps/install"
	_ "github.com/superproj/onex/pkg/apis/coordination/install"
	_ "github.com/superproj/onex/pkg/apis/core/install"
)
