// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package usercenter

import "time"

const (
	DefaultLRUSize = 1000
	// AccessTokenExpire is the expiration time for the access token.
	AccessTokenExpire = time.Hour * 2
	// RefreshTokenExpire is the expiration time for the refresh token.
	RefreshTokenExpire = time.Hour * 24
)

const (
	TemporaryKeyName = "_onex.io/temporary_key"
)
