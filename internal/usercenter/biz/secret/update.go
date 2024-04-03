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

// Update updates a secret.
func (b *secretBiz) Update(ctx context.Context, rq *v1.UpdateSecretRequest) error {
	secret, err := b.ds.Secrets().Get(ctx, onexx.FromUserID(ctx), rq.Name)
	if err != nil {
		return err
	}

	if rq.Expires != nil {
		secret.Expires = *rq.Expires
	}
	if rq.Status != nil {
		secret.Status = *rq.Status
	}
	if rq.Description != nil {
		secret.Description = *rq.Description
	}

	return b.ds.Secrets().Update(ctx, secret)
}
