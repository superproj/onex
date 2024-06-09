// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package biz

//go:generate mockgen -self_package github.com/superproj/onex/internal/gateway/biz -destination mock_biz.go -package biz github.com/superproj/onex/internal/gateway/biz IBiz

import (
	"github.com/google/wire"

	"github.com/superproj/onex/internal/gateway/biz/miner"
	"github.com/superproj/onex/internal/gateway/biz/minerset"
	"github.com/superproj/onex/internal/gateway/store"
	clientset "github.com/superproj/onex/pkg/generated/clientset/versioned"
	"github.com/superproj/onex/pkg/generated/informers"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(NewBiz, wire.Bind(new(IBiz), new(*biz)))

// IBiz defines functions used to return resource interface.
type IBiz interface {
	Miners() miner.MinerBiz
	MinerSets() minerset.MinerSetBiz
}

type biz struct {
	ds store.IStore
	cl clientset.Interface
	f  informers.SharedInformerFactory
}

// NewBiz returns IBiz interface.
func NewBiz(ds store.IStore, cl clientset.Interface, f informers.SharedInformerFactory) *biz {
	return &biz{ds, cl, f}
}

func (b *biz) MinerSets() minerset.MinerSetBiz {
	return minerset.New(b.ds, b.cl, b.f)
}

func (b *biz) Miners() miner.MinerBiz {
	return miner.New(b.ds, b.cl, b.f)
}
