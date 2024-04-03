// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package service

import (
	"github.com/google/wire"

	"github.com/superproj/onex/internal/usercenter/biz"
	v1 "github.com/superproj/onex/pkg/api/usercenter/v1"
)

// ProviderSet is a set of service providers, used for dependency injection.
var ProviderSet = wire.NewSet(NewUserCenterService)

// UserCenterService is a struct that implements the v1.UnimplementedUserCenterServer interface
// and holds the business logic, represented by a IBiz instance.
type UserCenterService struct {
	v1.UnimplementedUserCenterServer          // Embeds the generated UnimplementedUserCenterServer struct.
	biz                              biz.IBiz // A factory for creating business logic components.
}

// NewUserCenterService is a constructor function that takes a IBiz instance
// as an input and returns a new UserCenterService instance.
func NewUserCenterService(biz biz.IBiz) *UserCenterService {
	return &UserCenterService{biz: biz}
}
