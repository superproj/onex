// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package i18n

import (
	"context"
)

type translator struct{}

func NewContext(ctx context.Context, i *I18n) context.Context {
	return context.WithValue(ctx, translator{}, i)
}

func FromContext(ctx context.Context) *I18n {
	if i, ok := ctx.Value(translator{}).(*I18n); ok {
		return i
	}

	return New()
}
