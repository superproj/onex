// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package authn

import (
	"context"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

// IToken defines methods to implement a generic token.
type IToken interface {
	// Get token string.
	GetToken() string
	// Get token type.
	GetTokenType() string
	// Get token expiration timestamp.
	GetExpiresAt() int64
	// JSON encoding
	EncodeToJSON() ([]byte, error)
}

// Authenticator defines methods used for token processing.
type Authenticator interface {
	// Sign is used to generate a token.
	Sign(ctx context.Context, userID string) (IToken, error)

	// Destroy is used to destroy a token.
	Destroy(ctx context.Context, accessToken string) error

	// ParseClaims parse the token and return the claims.
	ParseClaims(ctx context.Context, accessToken string) (*jwt.RegisteredClaims, error)

	// Release used to release the requested resources.
	Release() error
}

// Encrypt encrypts the plain text with bcrypt.
func Encrypt(source string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(source), bcrypt.DefaultCost)
	return string(hashedBytes), err
}

// Compare compares the encrypted text with the plain text if it's the same.
func Compare(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
