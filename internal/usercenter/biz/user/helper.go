// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package user

import (
	"github.com/jinzhu/copier"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/superproj/onex/internal/usercenter/model"
	v1 "github.com/superproj/onex/pkg/api/usercenter/v1"
)

// ModelToReply converts a model.UserM to a v1.UserReply. It copies the data from
// userM to user and sets the CreatedAt and UpdatedAt fields to their respective timestamps.
func ModelToReply(userM *model.UserM) *v1.UserReply {
	var user v1.UserReply
	_ = copier.Copy(&user, userM)
	user.CreatedAt = timestamppb.New(userM.CreatedAt)
	user.UpdatedAt = timestamppb.New(userM.UpdatedAt)
	return &user
}
