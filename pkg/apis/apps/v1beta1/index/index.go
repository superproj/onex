// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

// Package index provides indexes for the api.
package index

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
)

// AddDefaultIndexes registers the default list of indexes.
func AddDefaultIndexes(ctx context.Context, mgr ctrl.Manager) error {
	if err := ByMinerPod(ctx, mgr); err != nil {
		return err
	}

	return nil
}
