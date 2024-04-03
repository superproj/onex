// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package jwt

import (
	"encoding/json"
)

// tokenInfo contains token information.
type tokenInfo struct {
	// Token string.
	Token string `json:"token"`

	// Token type.
	Type string `json:"type"`

	// Token expiration time
	ExpiresAt int64 `json:"expiresAt"`
}

func (t *tokenInfo) GetToken() string {
	return t.Token
}

func (t *tokenInfo) GetTokenType() string {
	return t.Type
}

func (t *tokenInfo) GetExpiresAt() int64 {
	return t.ExpiresAt
}

func (t *tokenInfo) EncodeToJSON() ([]byte, error) {
	return json.Marshal(t)
}
