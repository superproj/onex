// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/onex.
//

package app

import (
	"context"

	"k8s.io/kubernetes/pkg/controller/clusterroleaggregation"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/onexstack/onex/cmd/onex-controller-manager/names"
)

// newClusterRoleAggregationControllerDescriptor returns a ControllerDescriptor for
// the clusterrole-aggregation controller, which merges the rules of the ClusterRoles
// referenced by a ClusterRole's aggregationRule into that ClusterRole.
func newClusterRoleAggregationControllerDescriptor() *ControllerDescriptor {
	return &ControllerDescriptor{
		name:    names.ClusterRoleAggregationController,
		aliases: []string{"clusterroleaggregation"},
		addFunc: addClusterRoleAggregationController,
	}
}

// addClusterRoleAggregationController starts the clusterrole-aggregation controller.
func addClusterRoleAggregationController(ctx context.Context, _ ctrl.Manager, cctx ControllerContext) (bool, error) {
	clusterRoleInformer := cctx.InformerFactory.Rbac().V1().ClusterRoles()
	clusterRolesClient := cctx.ClientBuilder.ClientOrDie("clusterrole-aggregator").RbacV1()

	controller := clusterroleaggregation.NewClusterRoleAggregation(clusterRoleInformer, clusterRolesClient)
	go controller.Run(ctx, 5)

	return true, nil
}
