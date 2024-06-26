// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package auth

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/wire"
	lru "github.com/hashicorp/golang-lru"
	"gorm.io/gorm"

	known "github.com/superproj/onex/internal/pkg/known/usercenter"
	"github.com/superproj/onex/internal/usercenter/model"
	v1 "github.com/superproj/onex/pkg/api/usercenter/v1"
	"github.com/superproj/onex/pkg/authn"
	jwtauthn "github.com/superproj/onex/pkg/authn/jwt"
	"github.com/superproj/onex/pkg/log"
)

const (
	// reasonUnauthorized holds the error reason.
	reasonUnauthorized string = "Unauthorized"
)

// AuthnProviderSet is authn providers.
var AuthnProviderSet = wire.NewSet(NewAuthn, wire.Bind(new(AuthnInterface), new(*authnImpl)))

var (
	// ErrMissingKID is returned when the token format is invalid and the kid field is missing in the token header.
	ErrMissingKID = errors.Unauthorized(reasonUnauthorized, "Invalid token format: missing kid field in header")
	// ErrSecretDisabled is returned when the SecretID is disabled.
	ErrSecretDisabled = errors.Unauthorized(reasonUnauthorized, "SecretID is disabled")
)

// AuthnInterface defines the interface for authentication.
type AuthnInterface interface {
	// Sign is used to generate a access token. userID is the jwt identity key.
	Sign(ctx context.Context, userID string) (authn.IToken, error)
	// Verify is used to verify a access token. If the verification
	// is successful, userID will be returned.
	Verify(accessToken string) (string, error)
}

// SecretSetter is used to set or get a temporary secret key pairs.
type TemporarySecretSetter interface {
	Get(ctx context.Context, secretID string) (*model.SecretM, error)
	Set(ctx context.Context, userID string, expires int64) (*model.SecretM, error)
}

type authnImpl struct {
	setter  TemporarySecretSetter
	secrets *lru.Cache
}

// Ensure authnImpl implements AuthnInterface.
var _ AuthnInterface = (*authnImpl)(nil)

// NewAuthn returns a new instance of authn.
func NewAuthn(setter TemporarySecretSetter) (*authnImpl, error) {
	l, err := lru.New(known.DefaultLRUSize)
	if err != nil {
		log.Errorw(err, "Failed to create LRU cache")
		return nil, err
	}

	return &authnImpl{setter: setter, secrets: l}, nil
}

// Verify is used to verify a access token. If the verification
// is successful, userID will be returned.
func (a *authnImpl) Sign(ctx context.Context, userID string) (authn.IToken, error) {
	expires := time.Now().Add(known.AccessTokenExpire).Unix()

	secret, err := a.setter.Set(ctx, userID, expires)
	if err != nil {
		return nil, err
	}

	opts := []jwtauthn.Option{
		jwtauthn.WithSigningMethod(jwt.SigningMethodHS512),
		jwtauthn.WithIssuer("onex-usercenter"),
		jwtauthn.WithTokenHeader(map[string]any{"kid": secret.SecretID}),
		jwtauthn.WithExpired(known.AccessTokenExpire),
		jwtauthn.WithSigningKey([]byte(secret.SecretKey)),
	}

	j, err := jwtauthn.New(nil, opts...).Sign(ctx, userID)
	if err != nil {
		return nil, err
	}

	return j, nil
}

// Verify verifies the given access token and returns the userID associated with the token.
func (a *authnImpl) Verify(accessToken string) (string, error) {
	var secret *model.SecretM
	token, err := jwt.ParseWithClaims(accessToken, &jwt.RegisteredClaims{}, func(token *jwt.Token) (any, error) {
		// Validate the alg is HMAC signature
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return "", jwtauthn.ErrUnSupportSigningMethod
		}

		kid, ok := token.Header["kid"].(string)
		if !ok {
			return "", ErrMissingKID
		}

		var err error
		secret, err = a.GetSecret(kid)
		if err != nil {
			return "", err
		}

		if secret.Status == known.SecretStatusDisabled {
			return "", ErrSecretDisabled
		}

		return []byte(secret.SecretKey), nil
	})
	if err != nil {
		ve, ok := err.(*jwt.ValidationError)
		if !ok {
			return "", errors.Unauthorized(reasonUnauthorized, err.Error())
		}
		if ve.Errors&jwt.ValidationErrorMalformed != 0 {
			return "", jwtauthn.ErrTokenInvalid
		}
		if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
			return "", jwtauthn.ErrTokenExpired
		}
		return "", err
	}

	if !token.Valid {
		return "", jwtauthn.ErrTokenInvalid
	}

	if keyExpired(secret.Expires) {
		return "", jwtauthn.ErrTokenExpired
	}

	// you can return claims if you need
	// claims := token.Claims.(*jwt.RegisteredClaims)
	return secret.UserID, nil
}

// GetSecret returns the secret associated with the given key.
func (a *authnImpl) GetSecret(key string) (*model.SecretM, error) {
	s, ok := a.secrets.Get(key)
	if ok {
		return s.(*model.SecretM), nil
	}

	secret, err := a.setter.Get(context.Background(), key)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, v1.ErrorSecretNotFound(err.Error())
		}

		return nil, err
	}

	a.secrets.Add(key, secret)
	return secret, nil
}
