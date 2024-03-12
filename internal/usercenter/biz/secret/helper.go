// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package secret

import (
	"github.com/jinzhu/copier"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/superproj/onex/internal/usercenter/model"
	v1 "github.com/superproj/onex/pkg/api/usercenter/v1"
)

// ModelToReply converts a model.SecretM to a v1.SecretReply. It copies the data from
// secretM to secret and sets the CreatedAt and UpdatedAt fields to their respective timestamps.
func ModelToReply(secretM *model.SecretM) *v1.SecretReply {
	var secret v1.SecretReply
	_ = copier.Copy(&secret, secretM)
	secret.CreatedAt = timestamppb.New(secretM.CreatedAt)
	secret.UpdatedAt = timestamppb.New(secretM.UpdatedAt)
	return &secret
}
