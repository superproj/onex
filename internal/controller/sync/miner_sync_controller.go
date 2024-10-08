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
	"k8s.io/apimachinery/pkg/api/resource"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	gwmodel "github.com/superproj/onex/internal/gateway/model"
	"github.com/superproj/onex/internal/gateway/store"
	"github.com/superproj/onex/internal/pkg/known"
	"github.com/superproj/onex/pkg/apis/apps/v1beta1"
	"github.com/superproj/onex/pkg/store/where"
)

const minerControllerName = "controller-manager.minerSync"

// MinerSyncReconciler sync a Miner object to database.
type MinerSyncReconciler struct {
	client client.Client

	Store store.IStore
}

// SetupWithManager sets up the controller with the Manager.
func (r *MinerSyncReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options controller.Options) error {
	builder := ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.Miner{}).
		WithOptions(options).
		Named(minerControllerName)

	r.client = mgr.GetClient()

	return builder.Complete(r)
}

func (r *MinerSyncReconciler) Reconcile(ctx context.Context, rq ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	// Fetch the Miner instance
	m := &v1beta1.Miner{}
	if err := r.client.Get(ctx, rq.NamespacedName, m); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, r.Store.Miners().Delete(ctx, where.F("namespace", rq.Namespace, "name", rq.Name))
		}
		return ctrl.Result{}, err
	}

	mr, err := r.Store.Miners().Get(ctx, where.F("namespace", rq.Namespace, "name", rq.Name))
	if err != nil {
		// miner record not exist, create it.
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctrl.Result{}, addMiner(ctx, r.Store, m)
		}

		return ctrl.Result{}, err
	}

	// miner record exist, update it
	originMiner := new(gwmodel.MinerM)
	*originMiner = *mr

	mr = applyToMiner(mr, m)
	if !reflect.DeepEqual(mr, originMiner) {
		//nolint: errchkjson
		data, _ := json.Marshal(mr)
		log.V(4).Info("miner record changed", "newest", string(data))
		return ctrl.Result{}, r.Store.Miners().Update(ctx, mr)
	}

	return ctrl.Result{}, nil
}

// create miner record.
func addMiner(ctx context.Context, dbcli store.IStore, m *v1beta1.Miner) error {
	return dbcli.Miners().Create(ctx, applyToMiner(&gwmodel.MinerM{}, m))
}

func applyToMiner(mr *gwmodel.MinerM, m *v1beta1.Miner) *gwmodel.MinerM {
	mr.Namespace = m.Namespace
	mr.Name = m.Name
	mr.DisplayName = m.Spec.DisplayName
	mr.Phase = m.Status.Phase
	mr.MinerType = m.Spec.MinerType
	mr.ChainName = m.Spec.ChainName

	if mr.CPU == 0 || mr.Memory == 0 {
		mr.CPU, mr.Memory = GetMinerConfig(m.Annotations)
	}

	return mr
}

func GetMinerConfig(annotations map[string]string) (cpu int32, mem int32) {
	if annotations == nil {
		return 0, 0
	}

	if v, ok := annotations[known.CPUAnnotation]; ok {
		quantity, _ := resource.ParseQuantity(v)
		if val, ok := quantity.AsInt64(); ok {
			cpu = int32(val)
		}
	}
	if v, ok := annotations[known.MemoryAnnotation]; ok {
		quantity, _ := resource.ParseQuantity(v)
		if val, ok := quantity.AsInt64(); ok {
			mem = int32(val)
		}
	}

	return cpu, mem
}
