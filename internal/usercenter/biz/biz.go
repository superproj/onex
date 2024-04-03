// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package biz

//go:generate mockgen -self_package github.com/superproj/onex/internal/usercenter/biz -destination mock_biz.go -package biz github.com/superproj/onex/internal/usercenter/biz IBiz

import (
	"github.com/google/wire"

	"github.com/superproj/onex/internal/usercenter/auth"
	authbiz "github.com/superproj/onex/internal/usercenter/biz/auth"
	"github.com/superproj/onex/internal/usercenter/biz/secret"
	"github.com/superproj/onex/internal/usercenter/biz/user"
	"github.com/superproj/onex/internal/usercenter/store"
	"github.com/superproj/onex/pkg/authn"
)

// ProviderSet contains providers for creating instances of the biz struct.
var ProviderSet = wire.NewSet(NewBiz, wire.Bind(new(IBiz), new(*biz)))

// IBiz defines a set of methods for returning interfaces that the biz struct implements.
type IBiz interface {
	Secrets() secret.SecretBiz
	Users() user.UserBiz
	Auths() authbiz.AuthBiz
}

type biz struct {
	ds    store.IStore
	authn authn.Authenticator
	auth  auth.AuthProvider
}

// NewBiz returns a pointer to a new instance of the biz struct.
func NewBiz(ds store.IStore, authn authn.Authenticator, auth auth.AuthProvider) *biz {
	return &biz{ds: ds, authn: authn, auth: auth}
}

// Auths returns a new instance of the AuthBiz interface.
func (b *biz) Auths() authbiz.AuthBiz {
	return authbiz.New(b.ds, b.authn, b.auth)
}

// Users returns a new instance of the UserBiz interface.
func (b *biz) Users() user.UserBiz {
	return user.New(b.ds)
}

// Secrets returns a new instance of the SecretBiz interface.
func (b *biz) Secrets() secret.SecretBiz {
	return secret.New(b.ds)
}
