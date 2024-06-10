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
	"github.com/superproj/onex/internal/pkg/meta"
	v1 "github.com/superproj/onex/pkg/api/gateway/v1"
	"github.com/superproj/onex/pkg/apis/apps/v1beta1"
	clientset "github.com/superproj/onex/pkg/generated/clientset/versioned"
	"github.com/superproj/onex/pkg/generated/informers"
	listers "github.com/superproj/onex/pkg/generated/listers/apps/v1beta1"
	"github.com/superproj/onex/pkg/log"
)

// MinerSetBiz defines functions used to handle minerset rquest.
type MinerSetBiz interface {
	Create(ctx context.Context, namespace string, ms *v1beta1.MinerSet) error
	List(ctx context.Context, namespace string, rq *v1.ListMinerSetRequest) (*v1.ListMinerSetResponse, error)
	Get(ctx context.Context, namespace, name string) (*v1beta1.MinerSet, error)
	Update(ctx context.Context, namespace string, ms *v1beta1.MinerSet) error
	Delete(ctx context.Context, namespace, name string) error
	Scale(ctx context.Context, namespace, name string, replicas int32) error
}

type minerSetBiz struct {
	ds     store.IStore
	client clientset.Interface
	lister listers.MinerSetLister
}

var _ MinerSetBiz = (*minerSetBiz)(nil)

func New(ds store.IStore, client clientset.Interface, f informers.SharedInformerFactory) *minerSetBiz {
	return &minerSetBiz{ds, client, f.Apps().V1beta1().MinerSets().Lister()}
}

func (b *minerSetBiz) Create(ctx context.Context, namespace string, ms *v1beta1.MinerSet) error {
	_, err := b.client.AppsV1beta1().MinerSets(namespace).Create(ctx, ms, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (b *minerSetBiz) List(ctx context.Context, namespace string, rq *v1.ListMinerSetRequest) (*v1.ListMinerSetResponse, error) {
	total, list, err := b.ds.MinerSets().List(ctx, namespace, meta.WithOffset(rq.Offset), meta.WithLimit(rq.Limit))
	if err != nil {
		log.C(ctx).Errorw(err, "Failed to list minerset")
		return nil, err
	}

	mss := make([]*v1.MinerSet, 0, len(list))

	for _, item := range list {
		var ms v1.MinerSet
		_ = copier.Copy(&ms, &item)
		ms.CreatedAt = timestamppb.New(item.CreatedAt)
		ms.UpdatedAt = timestamppb.New(item.UpdatedAt)
		mss = append(mss, &ms)
	}

	return &v1.ListMinerSetResponse{TotalCount: total, MinerSets: mss}, nil
}

func (b *minerSetBiz) Get(ctx context.Context, namespace, name string) (*v1beta1.MinerSet, error) {
	ms, err := b.lister.MinerSets(namespace).Get(name)
	if err != nil {
		log.Errorw(err, "Failed to retrieve minerset", "minerset", klog.KRef(namespace, name))
		return nil, err
	}

	return ms, nil
}

func (b *minerSetBiz) Update(ctx context.Context, namespace string, ms *v1beta1.MinerSet) error {
	if _, err := b.client.AppsV1beta1().MinerSets(namespace).Update(ctx, ms, metav1.UpdateOptions{}); err != nil {
		log.Errorw(err, "Failed to update minerset", "minerset", klog.KRef(namespace, ms.Name))
	}

	return nil
}

func (b *minerSetBiz) Delete(ctx context.Context, namespace, name string) error {
	if err := b.client.AppsV1beta1().MinerSets(namespace).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		log.Errorw(err, "Failed to delete minerset", "minerset", klog.KRef(namespace, name))
	}

	return nil
}

func (b *minerSetBiz) Scale(ctx context.Context, namespace, name string, replicas int32) error {
	scale, err := b.client.AppsV1beta1().MinerSets(namespace).GetScale(ctx, name, metav1.GetOptions{})
	if err != nil {
		log.Errorw(err, "Failed to get scale", "minerset", klog.KRef(namespace, name))
		return err
	}

	scale.Spec.Replicas = replicas
	if _, err := b.client.AppsV1beta1().MinerSets(namespace).UpdateScale(ctx, name, scale, metav1.UpdateOptions{}); err != nil {
		log.Errorw(err, "Failed to scale minerset", "minerset", klog.KRef(namespace, name))
		return err
	}

	return nil
}
