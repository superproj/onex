// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package minerset

//go:generate mockgen -self_package github.com/superproj/onex/internal/gateway/biz/minerset -destination mock_minerset.go -package minerset github.com/superproj/onex/internal/gateway/biz/minerset MinerSetBiz

import (
	"context"

	"github.com/jinzhu/copier"
	"google.golang.org/protobuf/types/known/timestamppb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	"github.com/superproj/onex/internal/gateway/store"
	v1 "github.com/superproj/onex/pkg/api/gateway/v1"
	"github.com/superproj/onex/pkg/apis/apps/v1beta1"
	clientset "github.com/superproj/onex/pkg/generated/clientset/versioned"
	"github.com/superproj/onex/pkg/generated/informers"
	listers "github.com/superproj/onex/pkg/generated/listers/apps/v1beta1"
	"github.com/superproj/onex/pkg/log"
	"github.com/superproj/onex/pkg/store/where"
)

// MinerSetBiz defines the interface for handling miner set requests.
type MinerSetBiz interface {
	// Create creates a new miner set in the specified namespace.
	Create(ctx context.Context, namespace string, ms *v1beta1.MinerSet) error

	// Update updates an existing miner set in the specified namespace.
	Update(ctx context.Context, namespace string, ms *v1beta1.MinerSet) error

	// Delete removes a miner set by name from the specified namespace.
	Delete(ctx context.Context, namespace, name string) error

	// Get retrieves a miner set by name from the specified namespace.
	Get(ctx context.Context, namespace, name string) (*v1beta1.MinerSet, error)

	// List retrieves a list of miner sets in the specified namespace based on the request parameters.
	List(ctx context.Context, namespace string, rq *v1.ListMinerSetRequest) (*v1.ListMinerSetResponse, error)

	MinerSetExpansion
}

// MinerSetExpansion defines additional methods for miner set operations.
type MinerSetExpansion interface {
	// Scale adjusts the number of replicas for a miner set.
	Scale(ctx context.Context, namespace, name string, replicas int32) error
}

// minerSetBiz implements the MinerSetBiz interface.
type minerSetBiz struct {
	ds     store.IStore           // Data store interface for accessing miner set data.
	client clientset.Interface    // Kubernetes client interface for interacting with the API.
	lister listers.MinerSetLister // Lister interface for retrieving miner sets from the cache.
}

// Ensure minerSetBiz implements the MinerSetBiz interface.
var _ MinerSetBiz = (*minerSetBiz)(nil)

// New creates a new instance of minerSetBiz with the provided data store, client, and informer factory.
func New(ds store.IStore, client clientset.Interface, f informers.SharedInformerFactory) *minerSetBiz {
	return &minerSetBiz{ds, client, f.Apps().V1beta1().MinerSets().Lister()}
}

// Create creates a new miner set in the specified namespace.
func (b *minerSetBiz) Create(ctx context.Context, namespace string, ms *v1beta1.MinerSet) error {
	_, err := b.client.AppsV1beta1().MinerSets(namespace).Create(ctx, ms, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

// List retrieves a list of miner sets in the specified namespace based on the request parameters.
func (b *minerSetBiz) List(ctx context.Context, namespace string, rq *v1.ListMinerSetRequest) (*v1.ListMinerSetResponse, error) {
	total, minerSetList, err := b.ds.MinerSets().List(ctx, where.F("namespace", namespace).P(int(rq.Offset), int(rq.Limit)))
	if err != nil {
		log.C(ctx).Errorw(err, "Failed to list miner set")
		return nil, err
	}

	mss := make([]*v1.MinerSet, 0, len(minerSetList))

	for _, item := range minerSetList {
		var ms v1.MinerSet
		_ = copier.Copy(&ms, &item)
		ms.CreatedAt = timestamppb.New(item.CreatedAt)
		ms.UpdatedAt = timestamppb.New(item.UpdatedAt)
		mss = append(mss, &ms)
	}

	return &v1.ListMinerSetResponse{TotalCount: total, MinerSets: mss}, nil
}

// Get retrieves a miner set by name from the specified namespace.
func (b *minerSetBiz) Get(ctx context.Context, namespace, name string) (*v1beta1.MinerSet, error) {
	ms, err := b.lister.MinerSets(namespace).Get(name)
	if err != nil {
		log.Errorw(err, "Failed to retrieve miner set", "minerset", klog.KRef(namespace, name))
		return nil, err
	}

	return ms, nil
}

// Update updates an existing miner set in the specified namespace.
func (b *minerSetBiz) Update(ctx context.Context, namespace string, ms *v1beta1.MinerSet) error {
	if _, err := b.client.AppsV1beta1().MinerSets(namespace).Update(ctx, ms, metav1.UpdateOptions{}); err != nil {
		log.Errorw(err, "Failed to update miner set", "minerset", klog.KRef(namespace, ms.Name))
	}

	return nil
}

// Delete removes a miner set by name from the specified namespace.
func (b *minerSetBiz) Delete(ctx context.Context, namespace, name string) error {
	if err := b.client.AppsV1beta1().MinerSets(namespace).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		log.Errorw(err, "Failed to delete miner set", "minerset", klog.KRef(namespace, name))
	}

	return nil
}

// Scale adjusts the number of replicas for a miner set.
func (b *minerSetBiz) Scale(ctx context.Context, namespace, name string, replicas int32) error {
	// Retrieve the current scale configuration for the miner set.
	scale, err := b.client.AppsV1beta1().MinerSets(namespace).GetScale(ctx, name, metav1.GetOptions{})
	if err != nil {
		log.Errorw(err, "Failed to get scale", "minerset", klog.KRef(namespace, name))
		return err
	}

	// Update the replicas count in the scale specification.
	scale.Spec.Replicas = replicas
	if _, err := b.client.AppsV1beta1().MinerSets(namespace).UpdateScale(ctx, name, scale, metav1.UpdateOptions{}); err != nil {
		log.Errorw(err, "Failed to scale miner set", "minerset", klog.KRef(namespace, name))
		return err
	}

	return nil
}
