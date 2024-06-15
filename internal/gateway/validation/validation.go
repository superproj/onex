// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package validation

import (
	"context"

	"github.com/google/wire"

	"github.com/superproj/onex/internal/gateway/store"
	v1 "github.com/superproj/onex/pkg/api/gateway/v1"
)

// ProviderSet is a set of validator providers, used for dependency injection.
var ProviderSet = wire.NewSet(New, wire.Bind(new(any), new(*validator)))

// validator is a struct that implements the validate.IValidator interface.
type validator struct {
	ds store.IStore // Data store instance.
}

// New is a factory function that creates and initializes the custom validator.
// It takes a store.IStore instance as input and returns *validator.
func New(ds store.IStore) (*validator, error) {
	vd := &validator{ds: ds}

	return vd, nil
}

// ValidateListModelCompareRequest is a method that validates the ListModelCompareRequest input.
// In this particular case, no validation is performed and it always returns nil.
func (vd *validator) ValidateListModelCompareRequest(ctx context.Context, rq *v1.ListModelCompareRequest) error {
	return nil
}
