// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

//nolint:staticcheck
package lifecycle

import (
	"context"
	"fmt"
	"io"
	"time"

	"k8s.io/klog/v2"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilcache "k8s.io/apimachinery/pkg/util/cache"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apiserver/pkg/admission"
	"k8s.io/utils/clock"

	"github.com/superproj/onex/internal/controlplane/admission/initializer"
	clientset "github.com/superproj/onex/pkg/generated/clientset/versioned"
	"github.com/superproj/onex/pkg/generated/informers"
	corelisters "github.com/superproj/onex/pkg/generated/listers/core/v1"
)

const (
	// PluginName indicates the name of admission plug-in.
	PluginName = "NamespaceLifecycle"
	// how long a namespace stays in the force live lookup cache before expiration.
	forceLiveLookupTTL = 30 * time.Second
	// how long to wait for a missing namespace before re-checking the cache (and then doing a live lookup)
	// this accomplishes two things:
	// 1. It allows a watch-fed cache time to observe a namespace creation event
	// 2. It allows time for a namespace creation to distribute to members of a storage cluster,
	//    so the live lookup has a better chance of succeeding even if it isn't performed against the leader.
	missingNamespaceWait = 50 * time.Millisecond
)

// Register registers a plugin.
func Register(plugins *admission.Plugins) {
	plugins.Register(PluginName, func(config io.Reader) (admission.Interface, error) {
		return NewLifecycle(sets.NewString(metav1.NamespaceDefault, metav1.NamespaceSystem, metav1.NamespacePublic))
	})
}

// Lifecycle is an implementation of admission.Interface.
// It enforces life-cycle constraints around a Namespace depending on its Phase.
type Lifecycle struct {
	*admission.Handler
	client             clientset.Interface
	immortalNamespaces sets.String
	namespaceLister    corelisters.NamespaceLister
	// forceLiveLookupCache holds a list of entries for namespaces that we have a strong reason to believe are stale in our local cache.
	// if a namespace is in this cache, then we will ignore our local state and always fetch latest from api server.
	forceLiveLookupCache *utilcache.LRUExpireCache
}

var (
	_ = initializer.WantsExternalInformerFactory(&Lifecycle{})
	_ = initializer.WantsExternalClientSet(&Lifecycle{})
)

// Admit makes an admission decision based on the request attributes.
func (l *Lifecycle) Admit(ctx context.Context, a admission.Attributes, o admission.ObjectInterfaces) error {
	// prevent deletion of immortal namespaces
	if a.GetOperation() == admission.Delete && a.GetKind().GroupKind() == corev1.SchemeGroupVersion.WithKind("Namespace").GroupKind() && l.immortalNamespaces.Has(a.GetName()) {
		return apierrors.NewForbidden(a.GetResource().GroupResource(), a.GetName(), fmt.Errorf("this namespace may not be deleted"))
	}

	// always allow non-namespaced resources
	if len(a.GetNamespace()) == 0 && a.GetKind().GroupKind() != corev1.SchemeGroupVersion.WithKind("Namespace").GroupKind() {
		return nil
	}

	if a.GetKind().GroupKind() == corev1.SchemeGroupVersion.WithKind("Namespace").GroupKind() {
		// if a namespace is deleted, we want to prevent all further creates into it
		// while it is undergoing termination.  to reduce incidences where the cache
		// is slow to update, we add the namespace into a force live lookup list to ensure
		// we are not looking at stale state.
		if a.GetOperation() == admission.Delete {
			l.forceLiveLookupCache.Add(a.GetName(), true, forceLiveLookupTTL)
		}
		// allow all operations to namespaces
		return nil
	}

	// always allow deletion of other resources
	if a.GetOperation() == admission.Delete {
		return nil
	}

	// we need to wait for our caches to warm
	if !l.WaitForReady() {
		return admission.NewForbidden(a, fmt.Errorf("not yet ready to handle request"))
	}

	var (
		exists bool
		err    error
	)

	namespace, err := l.namespaceLister.Get(a.GetNamespace())
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return apierrors.NewInternalError(err)
		}
	} else {
		exists = true
	}

	if !exists && a.GetOperation() == admission.Create {
		// give the cache time to observe the namespace before rejecting a create.
		// this helps when creating a namespace and immediately creating objects within it.
		time.Sleep(missingNamespaceWait)
		namespace, err = l.namespaceLister.Get(a.GetNamespace())
		switch {
		case apierrors.IsNotFound(err):
			// no-op
		case err != nil:
			return apierrors.NewInternalError(err)
		default:
			exists = true
		}
		if exists {
			klog.V(4).InfoS("Namespace existed in cache after waiting", "namespace", klog.KRef("", a.GetNamespace()))
		}
	}

	// forceLiveLookup if true will skip looking at local cache state and instead always make a live call to server.
	forceLiveLookup := false
	if _, ok := l.forceLiveLookupCache.Get(a.GetNamespace()); ok {
		// we think the namespace was marked for deletion, but our current local cache says otherwise, we will force a live lookup.
		forceLiveLookup = exists && namespace.Status.Phase == corev1.NamespaceActive
	}

	// refuse to operate on non-existent namespaces
	if !exists || forceLiveLookup {
		// as a last resort, make a call directly to storage
		namespace, err = l.client.CoreV1().Namespaces().Get(context.TODO(), a.GetNamespace(), metav1.GetOptions{})
		switch {
		case apierrors.IsNotFound(err):
			return err
		case err != nil:
			return apierrors.NewInternalError(err)
		}

		klog.V(4).InfoS("Found namespace via storage lookup", "namespace", klog.KRef("", a.GetNamespace()))
	}

	// ensure that we're not trying to create objects in terminating namespaces
	if a.GetOperation() == admission.Create {
		if namespace.Status.Phase != corev1.NamespaceTerminating {
			return nil
		}

		err := admission.NewForbidden(a, fmt.Errorf("unable to create new content in namespace %s because it is being terminated", a.GetNamespace()))

		if apierr, ok := err.(*apierrors.StatusError); ok {
			apierr.ErrStatus.Details.Causes = append(apierr.ErrStatus.Details.Causes, metav1.StatusCause{
				Type:    corev1.NamespaceTerminatingCause,
				Message: fmt.Sprintf("namespace %s is being terminated", a.GetNamespace()),
				Field:   "metadata.namespace",
			})
		}
		return err
	}

	return nil
}

// NewLifecycle creates a new namespace Lifecycle admission control handler.
func NewLifecycle(immortalNamespaces sets.String) (*Lifecycle, error) {
	return newLifecycleWithClock(immortalNamespaces, clock.RealClock{})
}

func newLifecycleWithClock(immortalNamespaces sets.String, clock utilcache.Clock) (*Lifecycle, error) {
	forceLiveLookupCache := utilcache.NewLRUExpireCacheWithClock(100, clock)
	return &Lifecycle{
		Handler:              admission.NewHandler(admission.Create, admission.Update, admission.Delete),
		immortalNamespaces:   immortalNamespaces,
		forceLiveLookupCache: forceLiveLookupCache,
	}, nil
}

// SetInternalInformerFactory implements the WantsExternalInformerFactory interface.
func (l *Lifecycle) SetInternalInformerFactory(f informers.SharedInformerFactory) {
	namespaceInformer := f.Core().V1().Namespaces()
	l.namespaceLister = namespaceInformer.Lister()
	l.SetReadyFunc(namespaceInformer.Informer().HasSynced)
}

// SetExternalClientSet implements the WantsExternalClientSet interface.
func (l *Lifecycle) SetExternalClientSet(client clientset.Interface) {
	l.client = client
}

// ValidateInitialization implements the InitializationValidator interface.
func (l *Lifecycle) ValidateInitialization() error {
	if l.namespaceLister == nil {
		return fmt.Errorf("missing namespaceLister")
	}
	if l.client == nil {
		return fmt.Errorf("missing client")
	}
	return nil
}
