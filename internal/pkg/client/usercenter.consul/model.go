// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package usercenter

import (
	usercenterv1 "github.com/superproj/onex/pkg/api/usercenter/v1"
)

type GetSecretRequest struct {
	Username string
	Name     string
}

type GetSecretResponse = usercenterv1.SecretReply
