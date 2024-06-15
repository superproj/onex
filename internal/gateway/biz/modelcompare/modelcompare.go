// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package modelcompare

//go:generate mockgen -self_package github.com/superproj/onex/internal/gateway/biz/modelcompare -destination mock_modelcompare.go -package modelcompare github.com/superproj/onex/internal/gateway/biz/modelcompare ModelCompareBiz

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

// ModelCompareBiz defines functions used to handle modelcompare rquest.
type ModelCompareBiz interface {
	Create(ctx context.Context, namespace string, ms *v1beta1.ModelCompare) error
	List(ctx context.Context, namespace string, rq *v1.ListModelCompareRequest) (*v1.ListModelCompareResponse, error)
	Get(ctx context.Context, namespace, name string) (*v1beta1.ModelCompare, error)
	Update(ctx context.Context, namespace string, ms *v1beta1.ModelCompare) error
	Delete(ctx context.Context, namespace, name string) error
}

type modelCompareBiz struct {
	ds     store.IStore
	client clientset.Interface
	lister listers.ModelCompareLister
}

var _ ModelCompareBiz = (*modelCompareBiz)(nil)

func New(ds store.IStore, client clientset.Interface, f informers.SharedInformerFactory) *modelCompareBiz {
	return &modelCompareBiz{ds, client, f.Apps().V1beta1().ModelCompares().Lister()}
}

func (b *modelCompareBiz) Create(ctx context.Context, namespace string, mc *v1beta1.ModelCompare) error {
	_, err := b.client.AppsV1beta1().ModelCompares("default").Create(ctx, mc, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (b *modelCompareBiz) List(ctx context.Context, namespace string, rq *v1.ListModelCompareRequest) (*v1.ListModelCompareResponse, error) {
	total, list, err := b.ds.ModelCompares().List(ctx, "default", meta.WithOffset(rq.Offset), meta.WithLimit(rq.Limit))
	if err != nil {
		log.C(ctx).Errorw(err, "Failed to list modelcompare")
		return nil, err
	}

	// 夸资源调用
	// b.ds.Evaluates().List()....

	mcs := make([]*v1.ModelCompare, 0, len(list))

	for _, item := range list {
		var mc v1.ModelCompare
		_ = copier.Copy(&mc, &item)
		mc.UpdateTime = timestamppb.New(item.UpdateTime)
		mc.CreateTime = timestamppb.New(item.CreateTime)
		mc.StartedAt = timestamppb.New(item.StartedAt)
		mc.EndedAt = timestamppb.New(item.EndedAt)
		mcs = append(mcs, &mc)
	}

	return &v1.ListModelCompareResponse{TotalCount: total, Compares: mcs}, nil
}

func (b *modelCompareBiz) Get(ctx context.Context, namespace, name string) (*v1beta1.ModelCompare, error) {
	ms, err := b.lister.ModelCompares("default").Get(name)
	if err != nil {
		log.Errorw(err, "Failed to retrieve modelcompare", "modelcompare", klog.KRef(namespace, name))
		return nil, err
	}

	return ms, nil
}

func (b *modelCompareBiz) Update(ctx context.Context, namespace string, ms *v1beta1.ModelCompare) error {
	if _, err := b.client.AppsV1beta1().ModelCompares("default").Update(ctx, ms, metav1.UpdateOptions{}); err != nil {
		log.Errorw(err, "Failed to update modelcompare", "modelcompare", klog.KRef(namespace, ms.Name))
	}

	return nil
}

func (b *modelCompareBiz) Delete(ctx context.Context, namespace, name string) error {
	if err := b.client.AppsV1beta1().ModelCompares("default").Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		log.Errorw(err, "Failed to delete modelcompare", "modelcompare", klog.KRef(namespace, name))
	}

	return nil
}
