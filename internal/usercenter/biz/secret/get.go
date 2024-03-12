// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package secret

import (
	"context"

	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gorm"

	"github.com/superproj/onex/internal/pkg/onexx"
	v1 "github.com/superproj/onex/pkg/api/usercenter/v1"
)

// Get returns a single secret.
func (b *secretBiz) Get(ctx context.Context, rq *v1.GetSecretRequest) (*v1.SecretReply, error) {
	secretM, err := b.ds.Secrets().Get(ctx, onexx.FromUserID(ctx), rq.Name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, v1.ErrorSecretNotFound(err.Error())
		}

		return nil, err
	}

	return ModelToReply(secretM), nil
}
