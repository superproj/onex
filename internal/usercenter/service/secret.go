// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package service

import (
	"context"

	emptypb "google.golang.org/protobuf/types/known/emptypb"

	v1 "github.com/superproj/onex/pkg/api/usercenter/v1"
)

// CreateSecret is a method for creating a new secret.
// It takes a CreateSecretRequest as input and returns an Empty message or an error.
func (s *UserCenterService) CreateSecret(ctx context.Context, rq *v1.CreateSecretRequest) (*v1.SecretReply, error) {
	return s.biz.Secrets().Create(ctx, rq)
}

// ListSecret is a method for listing secrets.
// It takes a ListSecretRequest as input and returns a ListSecretResponse with the secrets or an error.
func (s *UserCenterService) ListSecret(ctx context.Context, rq *v1.ListSecretRequest) (*v1.ListSecretResponse, error) {
	r, err := s.biz.Secrets().List(ctx, rq)
	return r, err
}

// GetSecret is a method for retrieving a specific secret.
// It takes a GetSecretRequest as input and returns a SecretReply with the secret or an error.
func (s *UserCenterService) GetSecret(ctx context.Context, rq *v1.GetSecretRequest) (*v1.SecretReply, error) {
	return s.biz.Secrets().Get(ctx, rq)
}

// UpdateSecret is a method for updating a secret.
// It takes an UpdateSecretRequest as input and returns an Empty message or an error.
func (s *UserCenterService) UpdateSecret(ctx context.Context, rq *v1.UpdateSecretRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, s.biz.Secrets().Update(ctx, rq)
}

// DeleteSecret is a method for deleting a secret.
// It takes a DeleteSecretRequest as input and returns an Empty message or an error.
func (s *UserCenterService) DeleteSecret(ctx context.Context, rq *v1.DeleteSecretRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, s.biz.Secrets().Delete(ctx, rq)
}
