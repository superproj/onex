// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

//nolint:dupl
package sync

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"

	"gorm.io/gorm"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	gwmodel "github.com/superproj/onex/internal/gateway/model"
	"github.com/superproj/onex/internal/gateway/store"
	"github.com/superproj/onex/pkg/apis/apps/v1beta1"
)

const chainControllerName = "controller-manager.chainSync"

// ChainSyncReconciler sync a Chain object to database.
type ChainSyncReconciler struct {
	client client.Client
	Store  store.IStore
}

// SetupWithManager sets up the controller with the Manager.
func (r *ChainSyncReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options controller.Options) error {
	builder := ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.Chain{}).
		WithOptions(options).
		Named(chainControllerName)

	r.client = mgr.GetClient()

	return builder.Complete(r)
}

func (r *ChainSyncReconciler) Reconcile(ctx context.Context, rq ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	// Fetch the Chain instance
	ch := &v1beta1.Chain{}
	if err := r.client.Get(ctx, rq.NamespacedName, ch); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, r.Store.Chains().Delete(ctx, map[string]any{"namespace": rq.Namespace, "name": rq.Name})
		}
		return ctrl.Result{}, err
	}

	chr, err := r.Store.Chains().Get(ctx, map[string]any{"namespace": rq.Namespace, "name": rq.Name})
	if err != nil {
		// chain record not exist, create it.
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctrl.Result{}, addChain(ctx, r.Store, ch)
		}

		return ctrl.Result{}, err
	}

	// chain record exist, update it
	originChain := new(gwmodel.ChainM)
	*originChain = *chr

	chr = applyToChain(chr, ch)
	if !reflect.DeepEqual(chr, originChain) {
		//nolint: errchkjson
		data, _ := json.Marshal(chr)
		log.V(4).Info("chain record changed", "newest", string(data))
		return ctrl.Result{}, r.Store.Chains().Update(ctx, chr)
	}

	return ctrl.Result{}, nil
}

// create chain record.
func addChain(ctx context.Context, dbcli store.IStore, ch *v1beta1.Chain) error {
	return dbcli.Chains().Create(ctx, applyToChain(&gwmodel.ChainM{}, ch))
}

func applyToChain(chr *gwmodel.ChainM, ch *v1beta1.Chain) *gwmodel.ChainM {
	chr.Namespace = ch.Namespace
	chr.Name = ch.Name
	chr.DisplayName = ch.Spec.DisplayName
	chr.MinerType = ch.Spec.MinerType
	chr.Image = ch.Spec.Image
	chr.MinMineIntervalSeconds = ch.Spec.MinMineIntervalSeconds
	return chr
}
