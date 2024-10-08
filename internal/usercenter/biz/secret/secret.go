// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package secret

//go:generate mockgen -self_package github.com/superproj/onex/internal/usercenter/biz/secret -destination mock_secret.go -package secret github.com/superproj/onex/internal/usercenter/biz/secret SecretBiz

import (
	"context"
	"errors"
	"gorm.io/gorm"

	"github.com/jinzhu/copier"

	"github.com/superproj/onex/internal/pkg/onexx"
	"github.com/superproj/onex/internal/usercenter/conversion"
	"github.com/superproj/onex/internal/usercenter/model"
	"github.com/superproj/onex/internal/usercenter/store"
	v1 "github.com/superproj/onex/pkg/api/usercenter/v1"
	"github.com/superproj/onex/pkg/store/where"
)

// SecretBiz defines the interface for managing secrets.
type SecretBiz interface {
	// Create creates a new secret based on the provided request.
	Create(ctx context.Context, rq *v1.CreateSecretRequest) (*v1.SecretReply, error)

	// Update updates an existing secret based on the provided request.
	Update(ctx context.Context, rq *v1.UpdateSecretRequest) error

	// Delete removes a secret based on the provided request.
	Delete(ctx context.Context, rq *v1.DeleteSecretRequest) error

	// Get retrieves a secret by name based on the provided request.
	Get(ctx context.Context, rq *v1.GetSecretRequest) (*v1.SecretReply, error)

	// List retrieves a list of all secrets based on the provided request.
	List(ctx context.Context, rq *v1.ListSecretRequest) (*v1.ListSecretResponse, error)

	SecretExpansion
}

// SecretExpansion defines additional methods for secret operations.
type SecretExpansion interface {
	// Additional methods can be defined here in the future.
}

// secretBiz is the concrete implementation of the SecretBiz interface.
type secretBiz struct {
	ds store.IStore // Data store for persistent storage operations.
}

// Ensure secretBiz implements the SecretBiz interface.
var _ SecretBiz = (*secretBiz)(nil)

// New creates a new instance of secretBiz with the provided data store.
func New(ds store.IStore) *secretBiz {
	return &secretBiz{ds: ds}
}

// Create creates a new secret based on the provided request.
func (b *secretBiz) Create(ctx context.Context, rq *v1.CreateSecretRequest) (*v1.SecretReply, error) {
	var secretM model.SecretM
	_ = copier.Copy(&secretM, rq)          // Copy request data to the Secret model.
	secretM.UserID = onexx.FromUserID(ctx) // Set the user ID from the context.

	// Attempt to create the secret in the data store.
	if err := b.ds.Secrets().Create(ctx, &secretM); err != nil {
		return nil, v1.ErrorSecretCreateFailed("create secret failed: %s", err.Error()) // Handle creation error.
	}

	return conversion.ConvertToV1SecretReply(&secretM), nil // Convert and return the created secret.
}

// Update updates an existing secret based on the provided request.
func (b *secretBiz) Update(ctx context.Context, rq *v1.UpdateSecretRequest) error {
	// Retrieve the existing secret by name.
	secret, err := b.ds.Secrets().Get(ctx, where.T(ctx).F("name", rq.Name))
	if err != nil {
		return err // Return any error encountered.
	}

	// Update the fields if provided in the request.
	if rq.Expires != nil {
		secret.Expires = *rq.Expires
	}
	if rq.Status != nil {
		secret.Status = *rq.Status
	}
	if rq.Description != nil {
		secret.Description = *rq.Description
	}

	return b.ds.Secrets().Update(ctx, secret) // Update the secret in the data store.
}

// Delete removes a secret based on the provided request.
func (b *secretBiz) Delete(ctx context.Context, rq *v1.DeleteSecretRequest) error {
	// Delete the secret by name from the data store.
	return b.ds.Secrets().Delete(ctx, where.T(ctx).F("name", rq.Name))
}

// Get retrieves a secret by name based on the provided request.
func (b *secretBiz) Get(ctx context.Context, rq *v1.GetSecretRequest) (*v1.SecretReply, error) {
	// Retrieve the secret from the data store.
	secretM, err := b.ds.Secrets().Get(ctx, where.T(ctx).F("name", rq.Name))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, v1.ErrorSecretNotFound(err.Error()) // Return an error if secret is not found.
		}
		return nil, err // Return any other error encountered.
	}

	return conversion.ConvertToV1SecretReply(secretM), nil // Convert and return the found secret.
}

// List retrieves a list of all secrets from the data store.
func (b *secretBiz) List(ctx context.Context, rq *v1.ListSecretRequest) (*v1.ListSecretResponse, error) {
	// Retrieve the total count and list of secrets.
	count, secretList, err := b.ds.Secrets().List(ctx, where.T(ctx).P(int(rq.Offset), int(rq.Limit)))
	if err != nil {
		return nil, err // Return any error encountered.
	}

	// Convert the list of secrets to the response format.
	secrets := make([]*v1.SecretReply, 0)
	for _, item := range secretList {
		secrets = append(secrets, conversion.ConvertToV1SecretReply(item))
	}

	return &v1.ListSecretResponse{TotalCount: count, Secrets: secrets}, nil // Return the list of secrets.
}
