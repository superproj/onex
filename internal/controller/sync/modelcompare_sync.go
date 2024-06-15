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
	"time"

	"gorm.io/gorm"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	gwmodel "github.com/superproj/onex/internal/gateway/model"
	"github.com/superproj/onex/internal/gateway/store"
	"github.com/superproj/onex/pkg/apis/apps/v1beta1"
)

const modelCompareControllerName = "controller-manager.modelCompareSync"

// ModelCompareSyncReconciler sync a ModelCompare object to database.
type ModelCompareSyncReconciler struct {
	client client.Client

	Store store.IStore
}

// SetupWithManager sets up the controller with the Manager.
func (r *ModelCompareSyncReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options controller.Options) error {
	builder := ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.ModelCompare{}).
		WithOptions(options).
		Named(modelCompareControllerName)

	r.client = mgr.GetClient()

	return builder.Complete(r)
}

func (r *ModelCompareSyncReconciler) Reconcile(ctx context.Context, rq ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	// Fetch the ModelCompare instance
	mc := &v1beta1.ModelCompare{}
	if err := r.client.Get(ctx, rq.NamespacedName, mc); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, r.Store.ModelCompares().Delete(ctx, map[string]any{"namespace": rq.Namespace, "name": rq.Name})
		}
		return ctrl.Result{}, err
	}

	mcr, err := r.Store.ModelCompares().Get(ctx, map[string]any{"namespace": rq.Namespace, "name": rq.Name})
	if err != nil {
		// modelcompare record not exist, create it.
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctrl.Result{}, addModelCompare(ctx, r.Store, mc)
		}

		return ctrl.Result{}, err
	}

	// modelcompare record exist, update it
	originModelCompare := new(gwmodel.WalleModelCompare)
	*originModelCompare = *mcr

	mcr = applyToModelCompare(mcr, mc)
	if !reflect.DeepEqual(mcr, originModelCompare) {
		//nolint: errchkjson
		data, _ := json.Marshal(mcr)
		log.V(4).Info("modelcompare record changed", "newest", string(data))
		return ctrl.Result{}, r.Store.ModelCompares().Update(ctx, mcr)
	}

	return ctrl.Result{}, nil
}

// create chain record.
func addModelCompare(ctx context.Context, dbcli store.IStore, ch *v1beta1.ModelCompare) error {
	return dbcli.ModelCompares().Create(ctx, applyToModelCompare(&gwmodel.WalleModelCompare{}, ch))
}

func applyToModelCompare(mcr *gwmodel.WalleModelCompare, ch *v1beta1.ModelCompare) *gwmodel.WalleModelCompare {
	mcr.Namespace = ch.Namespace
	mcr.Name = ch.Name
	mcr.Creator = ch.Namespace
	mcr.CompareID = time.Now().Unix()
	mcr.CompareName = ch.Spec.DisplayName
	mcr.SceneID = 1
	mcr.SampleID = ch.Spec.Template.Spec.SampleID
	mcr.Status = ch.Status.Phase

	data, _ := json.Marshal(ch.Spec.ModelIDs)
	mcr.ModelIds = ptr.To(string(data))
	mcr.StartedAt = ch.Status.StartedAt.Time
	mcr.EndedAt = ch.Status.EndedAt.Time
	mcr.CreateTime = ch.GetCreationTimestamp().Time
	mcr.UpdateTime = ch.GetCreationTimestamp().Time
	return mcr
}
