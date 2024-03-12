// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package secret

import (
	"context"

	"github.com/superproj/onex/internal/pkg/onexx"
	v1 "github.com/superproj/onex/pkg/api/usercenter/v1"
)

// Delete deletes a secret.
func (b *secretBiz) Delete(ctx context.Context, rq *v1.DeleteSecretRequest) error {
	return b.ds.Secrets().Delete(ctx, onexx.FromUserID(ctx), rq.Name)
}
