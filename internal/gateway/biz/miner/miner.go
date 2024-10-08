// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package miner

//go:generate mockgen -self_package github.com/superproj/onex/internal/gateway/biz/miner -destination mock_miner.go -package miner github.com/superproj/onex/internal/gateway/biz/miner MinerBiz

import (
	"context"

	"github.com/jinzhu/copier"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/superproj/onex/internal/gateway/store"
	v1 "github.com/superproj/onex/pkg/api/gateway/v1"
	"github.com/superproj/onex/pkg/apis/apps/v1beta1"
	clientset "github.com/superproj/onex/pkg/generated/clientset/versioned"
	"github.com/superproj/onex/pkg/generated/informers"
	listers "github.com/superproj/onex/pkg/generated/listers/apps/v1beta1"
	"github.com/superproj/onex/pkg/log"
	"github.com/superproj/onex/pkg/store/where"
)

// MinerBiz defines the interface for handling miner requests.
type MinerBiz interface {
	// Create creates a new miner in the specified namespace.
	Create(ctx context.Context, namespace string, m *v1beta1.Miner) error

	// Update updates an existing miner in the specified namespace.
	Update(ctx context.Context, namespace string, m *v1beta1.Miner) error

	// Delete removes a miner by name from the specified namespace.
	Delete(ctx context.Context, namespace, name string) error

	// Get retrieves a miner by name from the specified namespace.
	Get(ctx context.Context, namespace, name string) (*v1beta1.Miner, error)

	// List retrieves a list of miners in the specified namespace based on the request parameters.
	List(ctx context.Context, namespace string, rq *v1.ListMinerRequest) (*v1.ListMinerResponse, error)

	MinerExpansion
}

// MinerExpansion defines additional methods for miner operations.
type MinerExpansion interface{}

// minerBiz implements the MinerBiz interface.
type minerBiz struct {
	ds     store.IStore        // Data store interface for accessing miner data.
	client clientset.Interface // Kubernetes client interface for interacting with the API.
	lister listers.MinerLister // Lister interface for retrieving miners from the cache.
}

// Ensure minerBiz implements the MinerBiz interface.
var _ MinerBiz = (*minerBiz)(nil)

// New creates a new instance of minerBiz with the provided data store, client, and informer factory.
func New(ds store.IStore, client clientset.Interface, f informers.SharedInformerFactory) *minerBiz {
	return &minerBiz{ds, client, f.Apps().V1beta1().Miners().Lister()}
}

// Create creates a new miner in the specified namespace.
func (b *minerBiz) Create(ctx context.Context, namespace string, m *v1beta1.Miner) error {
	_, err := b.client.AppsV1beta1().Miners(namespace).Create(ctx, m, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

// List retrieves a list of miners in the specified namespace based on the request parameters.
func (b *minerBiz) List(ctx context.Context, namespace string, rq *v1.ListMinerRequest) (*v1.ListMinerResponse, error) {
	total, minerList, err := b.ds.Miners().List(ctx, where.F("namespace", namespace).P(int(rq.Offset), int(rq.Limit)))
	if err != nil {
		log.Errorw(err, "Failed to list miner")
		return nil, err
	}

	miners := make([]*v1.Miner, 0, len(minerList))
	for _, item := range minerList {
		var m v1.Miner
		_ = copier.Copy(&m, &item)
		m.CreatedAt = timestamppb.New(item.CreatedAt)
		m.UpdatedAt = timestamppb.New(item.UpdatedAt)
		miners = append(miners, &m)
	}

	return &v1.ListMinerResponse{TotalCount: total, Miners: miners}, nil
}

// Get retrieves a miner by name from the specified namespace.
func (b *minerBiz) Get(ctx context.Context, namespace, name string) (*v1beta1.Miner, error) {
	m, err := b.lister.Miners(namespace).Get(name)
	if err != nil {
		log.Errorw(err, "Failed to retrieve miner")
		return nil, err
	}

	return m, nil
}

// Update updates an existing miner in the specified namespace.
func (b *minerBiz) Update(ctx context.Context, namespace string, m *v1beta1.Miner) error {
	if _, err := b.client.AppsV1beta1().Miners(namespace).Update(ctx, m, metav1.UpdateOptions{}); err != nil {
		log.Errorw(err, "Failed to update miner")
	}

	return nil
}

// Delete removes a miner by name from the specified namespace.
func (b *minerBiz) Delete(ctx context.Context, namespace, name string) error {
	if err := b.client.AppsV1beta1().Miners(namespace).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		log.Errorw(err, "Failed to delete miner")
	}

	return nil
}
