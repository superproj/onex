// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

// Package collections implements collection utilities.
package collections

import (
	"context"

	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/superproj/onex/pkg/apis/apps/v1beta1"
)

// GetFilteredMinersForMinerSet returns a list of miners that can be filtered or not.
// If no filter is supplied then all miners associated with the target minerset are returned.
func GetFilteredMinersForMinerSet(ctx context.Context, c client.Reader, ms *v1beta1.MinerSet, filters ...Func) (Miners, error) {
	ml := &v1beta1.MinerList{}
	if err := c.List(
		ctx,
		ml,
		client.InNamespace(ms.Namespace),
		client.MatchingLabels{
			v1beta1.MinerSetNameLabel: ms.Name,
		},
	); err != nil {
		return nil, errors.Wrap(err, "failed to list miners")
	}

	miners := FromMinerList(ml)
	return miners.Filter(filters...), nil
}
