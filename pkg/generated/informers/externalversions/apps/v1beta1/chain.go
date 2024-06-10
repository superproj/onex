// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.

// Code generated by informer-gen. DO NOT EDIT.

package v1beta1

import (
	"context"
	time "time"

	appsv1beta1 "github.com/superproj/onex/pkg/apis/apps/v1beta1"
	versioned "github.com/superproj/onex/pkg/generated/clientset/versioned"
	internalinterfaces "github.com/superproj/onex/pkg/generated/informers/externalversions/internalinterfaces"
	v1beta1 "github.com/superproj/onex/pkg/generated/listers/apps/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// ChainInformer provides access to a shared informer and lister for
// Chains.
type ChainInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1beta1.ChainLister
}

type chainInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewChainInformer constructs a new informer for Chain type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewChainInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredChainInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredChainInformer constructs a new informer for Chain type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredChainInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.AppsV1beta1().Chains(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.AppsV1beta1().Chains(namespace).Watch(context.TODO(), options)
			},
		},
		&appsv1beta1.Chain{},
		resyncPeriod,
		indexers,
	)
}

func (f *chainInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredChainInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *chainInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&appsv1beta1.Chain{}, f.defaultInformer)
}

func (f *chainInformer) Lister() v1beta1.ChainLister {
	return v1beta1.NewChainLister(f.Informer().GetIndexer())
}
