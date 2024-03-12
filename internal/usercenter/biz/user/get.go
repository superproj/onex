// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package user

import (
	"context"

	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gorm"

	"github.com/superproj/onex/internal/pkg/onexx"
	validationutil "github.com/superproj/onex/internal/pkg/util/validation"
	v1 "github.com/superproj/onex/pkg/api/usercenter/v1"
)

// Get retrieves a single user from the database.
func (b *userBiz) Get(ctx context.Context, rq *v1.GetUserRequest) (*v1.UserReply, error) {
	filters := map[string]any{"username": rq.Username}
	if !validationutil.IsAdminUser(onexx.FromUserID(ctx)) {
		filters["user_id"] = onexx.FromUserID(ctx)
	}

	userM, err := b.ds.Users().Fetch(ctx, filters)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, v1.ErrorUserNotFound(err.Error())
		}

		return nil, err
	}

	return ModelToReply(userM), nil
}
