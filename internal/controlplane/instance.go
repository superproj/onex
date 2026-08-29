// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/onex.
//

package controlplane

import (
	"fmt"
	"net/http"
	"time"

	coordinationv1 "k8s.io/api/coordination/v1"
	apiv1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	flowcontrolv1 "k8s.io/api/flowcontrol/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	apiserverfeatures "k8s.io/apiserver/pkg/features"
	peerreconcilers "k8s.io/apiserver/pkg/reconcilers"
	"k8s.io/apiserver/pkg/registry/generic"
	genericapiserver "k8s.io/apiserver/pkg/server"
	serverstorage "k8s.io/apiserver/pkg/server/storage"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	utilpeerproxy "k8s.io/apiserver/pkg/util/peerproxy"
	"k8s.io/apiserver/pkg/util/proxy"
	kubeinformers "k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/routes"

	"github.com/onexstack/onex/internal/controlplane/controller/systemnamespaces"
	coordinationrest "github.com/onexstack/onex/internal/registry/coordination/rest"
	corerest "github.com/onexstack/onex/internal/registry/core/rest"
	discoveryrest "github.com/onexstack/onex/internal/registry/discovery/rest"
	flowcontrolrest "github.com/onexstack/onex/internal/registry/flowcontrol/rest"
	"github.com/onexstack/onex/pkg/apiserver/storage"
)

const (
	// DefaultEndpointReconcilerInterval is the default amount of time for how often the endpoints for
	// the kubernetes Service are reconciled.
	DefaultEndpointReconcilerInterval = 10 * time.Second
	// DefaultEndpointReconcilerTTL is the default TTL timeout for the storage layer
	DefaultEndpointReconcilerTTL = 15 * time.Second
	// IdentityLeaseComponentLabelKey is used to apply a component label to identity lease objects, indicating:
	//   1. the lease is an identity lease (different from leader election leases)
	//   2. which component owns this lease
	IdentityLeaseComponentLabelKey = "apiserver.kubernetes.io/identity"
	// KubeAPIServer defines variable used internally when referring to kube-apiserver component
	KubeAPIServer = "kube-apiserver"
	// KubeAPIServerIdentityLeaseLabelSelector selects kube-apiserver identity leases
	KubeAPIServerIdentityLeaseLabelSelector = IdentityLeaseComponentLabelKey + "=" + KubeAPIServer
	// repairLoopInterval defines the interval used to run the Services ClusterIP and NodePort repair loops
	repairLoopInterval = 3 * time.Minute
)

// Extra defines extra configuration for the onex-apiserver.
type Extra struct {
	// Place you custom config here.
	APIResourceConfigSource serverstorage.APIResourceConfigSource
	StorageFactory          serverstorage.StorageFactory
	EventTTL                time.Duration
	EnableLogsSupport       bool
	ProxyTransport          *http.Transport

	// PeerProxy, if not nil, sets proxy transport between kube-apiserver peers for requests
	// that can not be served locally
	PeerProxy utilpeerproxy.Interface
	// PeerEndpointLeaseReconciler updates the peer endpoint leases
	PeerEndpointLeaseReconciler peerreconcilers.PeerEndpointLeaseReconciler

	// For external resources and rest storage providers.
	ExternalRESTStorageProviders []storage.RESTStorageProvider
	// ExternalGroupResources       []schema.GroupResource

	// Number of masters running; all masters must be started with the
	// same value for this field. (Numbers > 1 currently untested.)
	MasterCount int

	// KubeVersionedInformers     kubeinformers.SharedInformerFactory
	InternalVersionedInformers kubeinformers.SharedInformerFactory
	ExternalPostStartHooks     map[string]genericapiserver.PostStartHookFunc
}

// Config defines configuration for the onex-apiserver.
type Config struct {
	Generic *genericapiserver.RecommendedConfig
	Extra
}

type completedConfig struct {
	Generic genericapiserver.CompletedConfig
	*Extra
}

// CompletedConfig embeds a private pointer that cannot be instantiated outside of this package.
type CompletedConfig struct {
	*completedConfig
}

// Instance contains state for a onex-apiserver instance.
type Instance struct {
	GenericAPIServer *genericapiserver.GenericAPIServer

	APIResourceConfigSource serverstorage.APIResourceConfigSource
}

// Complete fills in any fields not set that are required to have valid data. It's mutating the receiver.
func (c *Config) Complete() CompletedConfig {
	return CompletedConfig{&completedConfig{
		Generic: c.Generic.Complete(),
		Extra:   &c.Extra,
	}}
}

// New returns a new instance of APIServer from the given config.
// Certain config fields will be set to a default value if unset.
func (c completedConfig) New(delegationTarget genericapiserver.DelegationTarget) (*Instance, error) {
	s, err := c.Generic.New("onex-apiserver", delegationTarget)
	if err != nil {
		return nil, err
	}

	if c.EnableLogsSupport {
		routes.Logs{}.Install(s.Handler.GoRestfulContainer)
	}

	m := &Instance{
		GenericAPIServer:        s,
		APIResourceConfigSource: c.APIResourceConfigSource,
	}

	clientset, err := clientset.NewForConfig(c.Generic.LoopbackClientConfig)
	if err != nil {
		return nil, err
	}

	// Install onex legacy rest storage
	// This part of the code is different from kube-apiserver because
	// we do not need to install all kube-apiserver legacy APIs.
	if err := m.InstallLegacyAPI(&c, c.Generic.RESTOptionsGetter); err != nil {
		return nil, err
	}

	// The order here is preserved in discovery.
	// If resources with identical names exist in more than one of these groups (e.g. "deployments.apps"" and "deployments.extensions"),
	// the order of this list determines which group an unqualified resource name (e.g. "deployments") should prefer.
	// This priority order is used for local discovery, but it ends up aggregated in `k8s.io/kubernetes/cmd/kube-apiserver/app/aggregator.go
	// with specific priorities.
	// TODO: describe the priority all the way down in the RESTStorageProviders and plumb it back through the various discovery
	// handlers that we have.
	restStorageProviders := []storage.RESTStorageProvider{
		// &admissionrest.StorageProvider{LoopbackClientConfig: c.GenericConfig.LoopbackClientConfig},
		coordinationrest.RESTStorageProvider{},
		discoveryrest.RESTStorageProvider{},
		flowcontrolrest.RESTStorageProvider{InformerFactory: c.InternalVersionedInformers},
	}
	restStorageProviders = append(restStorageProviders, c.ExternalRESTStorageProviders...)
	if err := m.InstallAPIs(c.APIResourceConfigSource, c.Generic.RESTOptionsGetter, restStorageProviders...); err != nil {
		return nil, err
	}

	m.GenericAPIServer.AddPostStartHookOrDie("start-system-namespaces-controller", func(hookContext genericapiserver.PostStartHookContext) error {
		go systemnamespaces.NewController(clientset, c.InternalVersionedInformers.Core().V1().Namespaces()).Run(hookContext.Done())
		return nil
	})

	// Here, I removed unused kube-apiserver post start hooks and
	// add post start hooks which onex-apiserver needs

	// TODO: copy from kube-apiserver
	m.GenericAPIServer.AddPostStartHookOrDie(
		"start-internal-informers",
		func(context genericapiserver.PostStartHookContext) error {
			// remove dependence with kube-apiserver
			// c.KubeVersionedInformers.Start(context.StopCh)
			c.InternalVersionedInformers.Start(context.Done())
			return nil
		},
	)

	for name, hook := range c.ExternalPostStartHooks {
		m.GenericAPIServer.AddPostStartHookOrDie(name, hook)
	}

	if utilfeature.DefaultFeatureGate.Enabled(apiserverfeatures.APIServerIdentity) {
		// put some post start hook here
		// refer to: https://github.com/kubernetes/kubernetes/blob/v1.29.3/pkg/controlplane/instance.go#L515
	}
	// Add PostStartHooks for Unknown Version Proxy filter.
	if c.PeerProxy != nil {
		c.Generic.AddPostStartHookOrDie("unknown-version-proxy-filter", func(context genericapiserver.PostStartHookContext) error {
			err := c.PeerProxy.WaitForCacheSync(context.Done())
			return err
		})
	}

	return m, nil
}

// InstallLegacyAPI will install the legacy APIs for the restStorageProviders if they are enabled.
func (m *Instance) InstallLegacyAPI(c *completedConfig, restOptionsGetter generic.RESTOptionsGetter) error {
	// This is different from the implementation of kube-apiserver, where we directly configure the
	// LegacyRESTStorageProvider field. Although it's a bit heavy-handed, it's definitely more convenient.
	legacyRESTStorageProvider := corerest.LegacyRESTStorageProvider{
		EventTTL: c.EventTTL,
		// If necessary in the future, you can uncomment the following comment codes
		// StorageFactory:       c.StorageFactory,
		// LoopbackClientConfig: c.GenericConfig.LoopbackClientConfig,
		// Informers:            c.VersionedInformers,
	}

	// The service proxy needs an EndpointSliceGetter to resolve backend endpoints.
	endpointSliceGetter, err := proxy.NewEndpointSliceListerGetter(
		c.InternalVersionedInformers.Discovery().V1().EndpointSlices().Lister())
	if err != nil {
		return fmt.Errorf("error building endpoint slice getter: %w", err)
	}
	legacyRESTStorageProvider.EndpointSliceGetter = endpointSliceGetter

	apiGroupInfo, err := legacyRESTStorageProvider.NewLegacyRESTStorage(restOptionsGetter)
	if err != nil {
		return fmt.Errorf("error building core storage: %w", err)
	}
	if len(apiGroupInfo.VersionedResourcesStorageMap) == 0 { // if all core storage is disabled, return.
		return nil
	}

	if err := m.GenericAPIServer.InstallLegacyAPIGroup(genericapiserver.DefaultLegacyAPIPrefix, &apiGroupInfo); err != nil {
		return fmt.Errorf("error in registering group versions: %w", err)
	}
	return nil
}

// Instance will install the APIs for the restStorageProviders if they are enabled.
func (m *Instance) InstallAPIs(
	apiResourceConfigSource serverstorage.APIResourceConfigSource,
	restOptionsGetter generic.RESTOptionsGetter,
	restStorageProviders ...storage.RESTStorageProvider,
) error {
	nonLegacy := []*genericapiserver.APIGroupInfo{}

	// used later in the loop to filter the served resource by those that have expired.
	resourceExpirationEvaluator, err := genericapiserver.NewResourceExpirationEvaluator(m.GenericAPIServer.EffectiveVersion.EmulationVersion())
	if err != nil {
		return err
	}

	for _, restStorageBuilder := range restStorageProviders {
		groupName := restStorageBuilder.GroupName()
		if !apiResourceConfigSource.AnyResourceForGroupEnabled(groupName) {
			klog.V(1).InfoS("Skipping disabled API group", "groupName", groupName)
			continue
		}
		apiGroupInfo, err := restStorageBuilder.NewRESTStorage(apiResourceConfigSource, restOptionsGetter)
		if err != nil {
			return fmt.Errorf("problem initializing API group %q: %w", groupName, err)
		}

		if len(apiGroupInfo.VersionedResourcesStorageMap) == 0 {
			// If we have no storage for any resource configured, this API group is effectively disabled.
			// This can happen when an entire API group, version, or development-stage (alpha, beta, GA) is disabled.
			klog.V(1).InfoS("API group is not enabled, skipping.", "groupName", groupName)
			continue
		}

		// Remove resources that serving kinds that are removed or not introduced yet at the current version.
		// We do this here so that we don't accidentally serve versions without resources or openapi information that for kinds we don't serve.
		// This is a spot above the construction of individual storage handlers so that no sig accidentally forgets to check.
		err = resourceExpirationEvaluator.RemoveUnavailableKinds(groupName, apiGroupInfo.Scheme, apiGroupInfo.VersionedResourcesStorageMap, m.APIResourceConfigSource)
		if err != nil {
			return err
		}
		if len(apiGroupInfo.VersionedResourcesStorageMap) == 0 {
			klog.V(1).Infof("Removing API group %v because it is time to stop serving it because it has no versions per APILifecycle.", groupName)
			continue
		}

		klog.V(1).Infof("Enabling API group %q.", groupName)

		if postHookProvider, ok := restStorageBuilder.(genericapiserver.PostStartHookProvider); ok {
			name, hook, err := postHookProvider.PostStartHook()
			if err != nil {
				klog.Fatalf("Error building PostStartHook: %v", err)
			}
			m.GenericAPIServer.AddPostStartHookOrDie(name, hook)
		}

		if len(groupName) == 0 {
			// the legacy group for core APIs is special that it is installed into /api via this special install method.
			if err := m.GenericAPIServer.InstallLegacyAPIGroup(genericapiserver.DefaultLegacyAPIPrefix, &apiGroupInfo); err != nil {
				return fmt.Errorf("error in registering legacy API: %w", err)
			}
		} else {
			// everything else goes to /apis
			nonLegacy = append(nonLegacy, &apiGroupInfo)
		}
	}

	if err := m.GenericAPIServer.InstallAPIGroups(nonLegacy...); err != nil {
		return fmt.Errorf("error in registering group versions: %w", err)
	}
	return nil
}

var (
	// UPDATEME: When add new api group.
	// stableAPIGroupVersionsEnabledByDefault is a list of our stable versions.
	stableAPIGroupVersionsEnabledByDefault = []schema.GroupVersion{
		apiv1.SchemeGroupVersion,
		coordinationv1.SchemeGroupVersion,
		discoveryv1.SchemeGroupVersion,
		flowcontrolv1.SchemeGroupVersion,
		// v1beta1.SchemeGroupVersion, // Migrate to WithOptions
	}

	// legacyBetaEnabledByDefaultResources is the list of beta resources we enable.  You may only add to this list
	// if your resource is already enabled by default in a beta level we still serve AND there is no stable API for it.
	// see https://github.com/kubernetes/enhancements/tree/master/keps/sig-architecture/3136-beta-apis-off-by-default
	// for more details.
	legacyBetaEnabledByDefaultResources = []schema.GroupVersionResource{}

	// betaAPIGroupVersionsDisabledByDefault is for all future beta groupVersions.
	betaAPIGroupVersionsDisabledByDefault = []schema.GroupVersion{}
)

// DefaultAPIResourceConfigSource returns which groupVersion enabled and its
// resources enabled/disabled.
func DefaultAPIResourceConfigSource() *serverstorage.ResourceConfig {
	ret := serverstorage.NewResourceConfig()
	// NOTE: GroupVersions listed here will be enabled by default. Don't put alpha versions in the list.
	ret.EnableVersions(stableAPIGroupVersionsEnabledByDefault...)

	// disable alpha and beta versions explicitly so we have a full list of what's possible to serve
	ret.DisableVersions(betaAPIGroupVersionsDisabledByDefault...)

	// enable the legacy beta resources that were present before stopped serving new beta APIs by default.
	ret.EnableResources(legacyBetaEnabledByDefaultResources...)

	return ret
}

func AddStableAPIGroupVersionsEnabledByDefault(versions ...schema.GroupVersion) {
	stableAPIGroupVersionsEnabledByDefault = append(stableAPIGroupVersionsEnabledByDefault, versions...)
}
