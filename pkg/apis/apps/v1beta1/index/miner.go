// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package index

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/superproj/onex/pkg/apis/apps/v1beta1"
)

const (
	// MinerPodNameField is used by the Miner Controller to index Miners by Pod name, and add a watch on Pods.
	MinerPodNameField = "status.podRef.name"
)

// ByMinerPod adds the miner pod name index to the
// managers cache.
func ByMinerPod(ctx context.Context, mgr ctrl.Manager) error {
	if err := mgr.GetCache().IndexField(ctx, &v1beta1.Miner{}, MinerPodNameField, MinerByPodName); err != nil {
		return err
	}

	return nil
}

// MinerByPodName contains the logic to index Miners by Pod name.
func MinerByPodName(o client.Object) []string {
	miner, ok := o.(*v1beta1.Miner)
	if !ok {
		panic(fmt.Sprintf("Expected a Miner but got a %T", o))
	}
	if miner.Status.PodRef != nil {
		return []string{miner.Status.PodRef.Name}
	}
	return nil
}
