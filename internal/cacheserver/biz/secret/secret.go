// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package secret

//go:generate mockgen -destination mock_secret.go -package secret github.com/superproj/onex/internal/cacheserver/biz/secret SecretBiz

import (
	"context"
	"time"

	"github.com/jinzhu/copier"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/superproj/onex/internal/usercenter/model"
	v1 "github.com/superproj/onex/pkg/api/cacheserver/v1"
	"github.com/superproj/onex/pkg/cache"
)

// SecretBiz is the interface for managing secrets in the cache.
type SecretBiz interface {
	Set(ctx context.Context, rq *v1.SetSecretRequest) error
	Get(ctx context.Context, rq *v1.GetSecretRequest) (*v1.GetSecretResponse, error)
	Del(ctx context.Context, rq *v1.DelSecretRequest) error
}

// secretBiz is the implementation of SecretBiz interface.
type secretBiz struct {
	cache *cache.ChainCache[any]
}

// Ensure that secretBiz implements the SecretBiz interface.
var _ SecretBiz = (*secretBiz)(nil)

// New creates a new instance of secretBiz.
func New(cache *cache.ChainCache[any]) *secretBiz {
	return &secretBiz{cache: cache}
}

// Set stores a secret in the cache.
func (b *secretBiz) Set(ctx context.Context, rq *v1.SetSecretRequest) error {
	secret := &model.SecretM{
		Name:        rq.Name,
		SecretID:    rq.Key,
		Description: rq.Description,
	}
	if rq.Expire != nil {
		secret.Expires = time.Now().Add(rq.Expire.AsDuration()).Unix()
	}

	return b.cache.Set(ctx, rq.Key, secret)
}

// Get retrieves a secret from the cache.
func (b *secretBiz) Get(ctx context.Context, rq *v1.GetSecretRequest) (*v1.GetSecretResponse, error) {
	value, err := b.cache.Get(ctx, rq.Key)
	if err != nil {
		return nil, err
	}

	secret := value.(*model.SecretM)

	var rp v1.GetSecretResponse
	_ = copier.Copy(&rp, value)
	rp.CreatedAt = timestamppb.New(secret.CreatedAt)
	rp.UpdatedAt = timestamppb.New(secret.UpdatedAt)
	return &rp, nil
}

// Del deletes a secret from the cache.
func (b *secretBiz) Del(ctx context.Context, rq *v1.DelSecretRequest) error {
	return b.cache.Del(ctx, rq.Key)
}
