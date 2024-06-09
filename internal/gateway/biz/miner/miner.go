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
	"github.com/superproj/onex/internal/pkg/meta"
	v1 "github.com/superproj/onex/pkg/api/gateway/v1"
	"github.com/superproj/onex/pkg/apis/apps/v1beta1"
	clientset "github.com/superproj/onex/pkg/generated/clientset/versioned"
	"github.com/superproj/onex/pkg/generated/informers"
	listers "github.com/superproj/onex/pkg/generated/listers/apps/v1beta1"
	"github.com/superproj/onex/pkg/log"
)

// MinerBiz defines functions used to handle miner rquest.
type MinerBiz interface {
	Create(ctx context.Context, namespace string, m *v1beta1.Miner) error
	List(ctx context.Context, namespace string, rq *v1.ListMinerRequest) (*v1.ListMinerResponse, error)
	Get(ctx context.Context, namespace, name string) (*v1beta1.Miner, error)
	Update(ctx context.Context, namespace string, m *v1beta1.Miner) error
	Delete(ctx context.Context, namespace, name string) error
}

type minerBiz struct {
	ds     store.IStore
	client clientset.Interface
	lister listers.MinerLister
}

var _ MinerBiz = (*minerBiz)(nil)

func New(ds store.IStore, client clientset.Interface, f informers.SharedInformerFactory) *minerBiz {
	return &minerBiz{ds, client, f.Apps().V1beta1().Miners().Lister()}
}

func (b *minerBiz) Create(ctx context.Context, namespace string, m *v1beta1.Miner) error {
	_, err := b.client.AppsV1beta1().Miners(namespace).Create(ctx, m, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (b *minerBiz) List(ctx context.Context, namespace string, rq *v1.ListMinerRequest) (*v1.ListMinerResponse, error) {
	total, list, err := b.ds.Miners().List(ctx, namespace, meta.WithOffset(rq.Offset), meta.WithLimit(rq.Limit))
	if err != nil {
		log.Errorw(err, "Failed to list miner")
		return nil, err
	}

	miners := make([]*v1.Miner, 0, len(list))
	for _, item := range list {
		var m v1.Miner
		_ = copier.Copy(&m, &item)
		m.CreatedAt = timestamppb.New(item.CreatedAt)
		m.UpdatedAt = timestamppb.New(item.UpdatedAt)
		miners = append(miners, &m)
	}

	return &v1.ListMinerResponse{TotalCount: total, Miners: miners}, nil
}

func (b *minerBiz) Get(ctx context.Context, namespace, name string) (*v1beta1.Miner, error) {
	m, err := b.lister.Miners(namespace).Get(name)
	if err != nil {
		log.Errorw(err, "Failed to retrieve miner")
		return nil, err
	}

	return m, nil
}

func (b *minerBiz) Update(ctx context.Context, namespace string, m *v1beta1.Miner) error {
	if _, err := b.client.AppsV1beta1().Miners(namespace).Update(ctx, m, metav1.UpdateOptions{}); err != nil {
		log.Errorw(err, "Failed to update miner")
	}

	return nil
}

func (b *minerBiz) Delete(ctx context.Context, namespace, name string) error {
	if err := b.client.AppsV1beta1().Miners(namespace).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		log.Errorw(err, "Failed to delete miner")
	}

	return nil
}
