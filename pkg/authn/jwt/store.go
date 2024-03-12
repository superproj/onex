// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package jwt

import (
	"context"
	"time"
)

// Storer token storage interface.
type Storer interface {
	// Store token data and specify expiration time.
	Set(ctx context.Context, accessToken string, expiration time.Duration) error

	// Delete token data from storage.
	Delete(ctx context.Context, accessToken string) (bool, error)

	// Check if token exists.
	Check(ctx context.Context, accessToken string) (bool, error)

	// Close the storage.
	Close() error
}
