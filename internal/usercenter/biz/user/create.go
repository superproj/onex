// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package user

import (
	"context"
	"regexp"

	"github.com/jinzhu/copier"

	"github.com/superproj/onex/internal/usercenter/model"
	v1 "github.com/superproj/onex/pkg/api/usercenter/v1"
)

// Create creates a new user and stores it in the database.
func (b *userBiz) Create(ctx context.Context, rq *v1.CreateUserRequest) (*v1.UserReply, error) {
	var userM model.UserM
	_ = copier.Copy(&userM, rq)
	err := b.ds.TX(ctx, func(ctx context.Context) error {
		if err := b.ds.Users().Create(ctx, &userM); err != nil {
			match, _ := regexp.MatchString("Duplicate entry '.*' for key 'username'", err.Error())
			if match {
				return v1.ErrorUserAlreadyExists("user %q already exist", userM.Username)
			}

			return v1.ErrorUserCreateFailed("create user failed: %s", err.Error())
		}

		secretM := &model.SecretM{
			UserID:      userM.UserID,
			Name:        "generated",
			Expires:     0,
			Description: "automatically generated when user is created",
		}
		if err := b.ds.Secrets().Create(ctx, secretM); err != nil {
			return v1.ErrorSecretCreateFailed("create secret failed: %s", err.Error())
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return ModelToReply(&userM), nil
}
