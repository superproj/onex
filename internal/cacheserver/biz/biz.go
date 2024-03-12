// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package biz

//go:generate mockgen -destination mock_biz.go -package biz github.com/superproj/onex/internal/cacheserver/biz IBiz

import (
	"github.com/golang/protobuf/ptypes/any"
	"github.com/google/wire"

	"github.com/superproj/onex/internal/cacheserver/biz/namespaced"
	"github.com/superproj/onex/internal/cacheserver/biz/secret"
	"github.com/superproj/onex/internal/cacheserver/store"
	"github.com/superproj/onex/pkg/cache"
)

// ProviderSet contains providers for creating instances of the biz struct.
var ProviderSet = wire.NewSet(NewBiz, wire.Bind(new(IBiz), new(*biz)))

// IBiz defines the methods that need to be implemented by the Biz layer.
type IBiz interface {
	Namespace(namespace string) namespaced.NamespacedBiz
	Secrets() secret.SecretBiz
}

// biz is a concrete implementation of IBiz.
type biz struct {
	cache cache.Cache[*any.Any]
	store store.IStore
}

// Ensure that biz implements the IBiz interface.
var _ IBiz = (*biz)(nil)

// NewBiz creates an instance of IBiz.
func NewBiz(cache cache.Cache[*any.Any], store store.IStore) *biz {
	return &biz{cache: cache, store: store}
}

// Namespace returns a NamespacedBiz instance for the specified namespace.
func (b *biz) Namespace(namespace string) namespaced.NamespacedBiz {
	return namespaced.New(b.cache, namespace)
}

// Secrets returns a SecretBiz instance.
func (b *biz) Secrets() secret.SecretBiz {
	return secret.New(b.store.Secrets())
}
