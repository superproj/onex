// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package secret

import (
	"context"

	"github.com/superproj/onex/internal/pkg/onexx"
	v1 "github.com/superproj/onex/pkg/api/usercenter/v1"
)

// List returns a list of secrets.
func (b *secretBiz) List(ctx context.Context, rq *v1.ListSecretRequest) (*v1.ListSecretResponse, error) {
	count, list, err := b.ds.Secrets().List(ctx, onexx.FromUserID(ctx))
	if err != nil {
		return nil, err
	}

	secrets := make([]*v1.SecretReply, 0)
	for _, item := range list {
		secrets = append(secrets, ModelToReply(item))
	}

	return &v1.ListSecretResponse{TotalCount: count, Secrets: secrets}, nil
}
