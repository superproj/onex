// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package secret

import (
	"context"

	"github.com/jinzhu/copier"

	"github.com/superproj/onex/internal/pkg/onexx"
	"github.com/superproj/onex/internal/usercenter/model"
	v1 "github.com/superproj/onex/pkg/api/usercenter/v1"
)

// Create creates a new secret.
func (b *secretBiz) Create(ctx context.Context, rq *v1.CreateSecretRequest) (*v1.SecretReply, error) {
	var secretM model.SecretM
	_ = copier.Copy(&secretM, rq)
	secretM.UserID = onexx.FromUserID(ctx)

	if err := b.ds.Secrets().Create(ctx, &secretM); err != nil {
		return nil, v1.ErrorSecretCreateFailed("create secret failed: %s", err.Error())
	}

	return ModelToReply(&secretM), nil
}
