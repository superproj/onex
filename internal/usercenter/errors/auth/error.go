// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package auth

import "github.com/go-kratos/kratos/v2/errors"

// ErrAuthFail is a predefined error that represents an authentication failure.
// This error occurs when a token is missing or incorrect.
var ErrAuthFail = errors.New(401, "Authentication failed", "Missing token or token incorrect")
