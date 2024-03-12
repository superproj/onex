// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package user

import (
	"context"

	"github.com/superproj/onex/internal/pkg/onexx"
	validationutil "github.com/superproj/onex/internal/pkg/util/validation"
	v1 "github.com/superproj/onex/pkg/api/usercenter/v1"
	"github.com/superproj/onex/pkg/authn"
)

// Update updates a user's information in the database.
func (b *userBiz) Update(ctx context.Context, rq *v1.UpdateUserRequest) error {
	filters := map[string]any{"username": rq.Username}
	if !validationutil.IsAdminUser(onexx.FromUserID(ctx)) {
		filters["user_id"] = onexx.FromUserID(ctx)
	}

	userM, err := b.ds.Users().Fetch(ctx, filters)
	if err != nil {
		return err
	}

	if rq.Nickname != nil {
		userM.Nickname = *rq.Nickname
	}
	if rq.Email != nil {
		userM.Email = *rq.Email
	}
	if rq.Phone != nil {
		userM.Phone = *rq.Phone
	}

	return b.ds.Users().Update(ctx, userM)
}

// UpdatePassword updates a user's password in the database.
// Note that after updating the password, if the JWT Token has not expired, it can
// still be accessed through the token, the token is not deleted synchronously here.
func (b *userBiz) UpdatePassword(ctx context.Context, rq *v1.UpdatePasswordRequest) error {
	userM, err := b.ds.Users().Get(ctx, onexx.FromUserID(ctx), rq.Username)
	if err != nil {
		return err
	}

	if err := authn.Compare(userM.Password, rq.OldPassword); err != nil {
		return v1.ErrorUserLoginFailed("password incorrect")
	}
	userM.Password, _ = authn.Encrypt(rq.NewPassword)

	return b.ds.Users().Update(ctx, userM)
}
