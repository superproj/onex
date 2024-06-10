// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package exists

import (
	"context"
	"fmt"
	"io"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/admission"
	api "k8s.io/kubernetes/pkg/apis/core"

	"github.com/superproj/onex/internal/admission/initializer"
	clientset "github.com/superproj/onex/pkg/generated/clientset/versioned"
	"github.com/superproj/onex/pkg/generated/informers"
	corev1listers "github.com/superproj/onex/pkg/generated/listers/core/v1"
)

// PluginName indicates name of admission plugin.
const PluginName = "NamespaceExists"

// Register registers a plugin.
func Register(plugins *admission.Plugins) {
	plugins.Register(PluginName, func(config io.Reader) (admission.Interface, error) {
		return NewExists(), nil
	})
}

// Exists is an implementation of admission.Interface.
// It rejects all incoming requests in a namespace context if the namespace does not exist.
// It is useful in deployments that want to enforce pre-declaration of a Namespace resource.
type Exists struct {
	*admission.Handler
	client          clientset.Interface
	namespaceLister corev1listers.NamespaceLister
}

var (
	_ admission.ValidationInterface = &Exists{}
	_                               = initializer.WantsInternalMinerInformerFactory(&Exists{})
	_                               = initializer.WantsExternalMinerClientSet(&Exists{})
)

// Validate makes an admission decision based on the request attributes.
func (e *Exists) Validate(ctx context.Context, a admission.Attributes, o admission.ObjectInterfaces) error {
	// if we're here, then we've already passed authentication, so we're allowed to do what we're trying to do
	// if we're here, then the API server has found a route, which means that if we have a non-empty namespace
	// its a namespaced resource.
	if len(a.GetNamespace()) == 0 || a.GetKind().GroupKind() == api.Kind("Namespace") {
		return nil
	}

	// we need to wait for our caches to warm
	if !e.WaitForReady() {
		return admission.NewForbidden(a, fmt.Errorf("not yet ready to handle request"))
	}
	_, err := e.namespaceLister.Get(a.GetNamespace())
	if err == nil {
		return nil
	}
	if !apierrors.IsNotFound(err) {
		return apierrors.NewInternalError(err)
	}

	// in case of latency in our caches, make a call direct to storage to verify that it truly exists or not
	_, err = e.client.CoreV1().Namespaces().Get(context.TODO(), a.GetNamespace(), metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return err
		}
		return apierrors.NewInternalError(err)
	}

	return nil
}

// NewExists creates a new namespace exists admission control handler.
func NewExists() *Exists {
	return &Exists{
		Handler: admission.NewHandler(admission.Create, admission.Update, admission.Delete),
	}
}

// SetExternalMinerClientSet implements the WantsExternalMinerClientSet interface.
func (e *Exists) SetExternalMinerClientSet(client clientset.Interface) {
	e.client = client
}

// SetInternalMinerInformerFactory implements the WantsInternalMinerInformerFactory interface.
func (e *Exists) SetInternalMinerInformerFactory(f informers.SharedInformerFactory) {
	namespaceInformer := f.Core().V1().Namespaces()
	e.namespaceLister = namespaceInformer.Lister()
	e.SetReadyFunc(namespaceInformer.Informer().HasSynced)
}

// ValidateInitialization implements the InitializationValidator interface.
func (e *Exists) ValidateInitialization() error {
	if e.namespaceLister == nil {
		return fmt.Errorf("missing namespaceLister")
	}
	if e.client == nil {
		return fmt.Errorf("missing client")
	}
	return nil
}
