// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.

// Code generated by informer-gen. DO NOT EDIT.

package v1

import (
	"context"
	time "time"

	versioned "github.com/superproj/onex/pkg/generated/clientset/versioned"
	internalinterfaces "github.com/superproj/onex/pkg/generated/informers/externalversions/internalinterfaces"
	v1 "github.com/superproj/onex/pkg/generated/listers/core/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// NamespaceInformer provides access to a shared informer and lister for
// Namespaces.
type NamespaceInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1.NamespaceLister
}

type namespaceInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// NewNamespaceInformer constructs a new informer for Namespace type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewNamespaceInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredNamespaceInformer(client, resyncPeriod, indexers, nil)
}

// NewFilteredNamespaceInformer constructs a new informer for Namespace type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredNamespaceInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.CoreV1().Namespaces().List(context.TODO(), options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.CoreV1().Namespaces().Watch(context.TODO(), options)
			},
		},
		&corev1.Namespace{},
		resyncPeriod,
		indexers,
	)
}

func (f *namespaceInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredNamespaceInformer(client, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *namespaceInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&corev1.Namespace{}, f.defaultInformer)
}

func (f *namespaceInformer) Lister() v1.NamespaceLister {
	return v1.NewNamespaceLister(f.Informer().GetIndexer())
}
