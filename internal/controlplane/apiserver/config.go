// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/onex.
//

// Package app does all of the work necessary to create a OneX
// APIServer by binding together the API, master and APIServer infrastructure.
//
//nolint:nakedret
package apiserver

import (
	"fmt"
	"strconv"
	"time"

	"github.com/blang/semver/v4"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	kversion "k8s.io/apimachinery/pkg/version"
	"k8s.io/apiserver/pkg/admission"
	"k8s.io/client-go/rest"

	//"k8s.io/apiserver/pkg/admission"
	"k8s.io/apiserver/pkg/endpoints/discovery/aggregated"
	openapinamer "k8s.io/apiserver/pkg/endpoints/openapi"
	"k8s.io/apiserver/pkg/reconcilers"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/apiserver/pkg/server/filters"
	serverstorage "k8s.io/apiserver/pkg/server/storage"
	"k8s.io/apiserver/pkg/util/openapi"
	utilpeerproxy "k8s.io/apiserver/pkg/util/peerproxy"
	"k8s.io/client-go/informers"
	coordinationv1informers "k8s.io/client-go/informers/coordination/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/transport"
	openapicommon "k8s.io/kube-openapi/pkg/common"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	api "k8s.io/kubernetes/pkg/apis/core"

	"github.com/onexstack/onex/internal/controlplane"
	"github.com/onexstack/onex/internal/controlplane/admission/initializer"
	controlplaneoptions "github.com/onexstack/onex/internal/controlplane/apiserver/options"
	"github.com/onexstack/onex/pkg/apiserver/storage"
	"github.com/onexstack/onexstack/pkg/version"
)

const (
	// DefaultPeerEndpointReconcileInterval is the default amount of time for how often
	// the peer endpoint leases are reconciled.
	DefaultPeerEndpointReconcileInterval = 10 * time.Second
	// DefaultPeerEndpointReconcilerTTL is the default TTL timeout for peer endpoint
	// leases on the storage layer
	DefaultPeerEndpointReconcilerTTL = 15 * time.Second

	// IdentityLeaseComponentLabelKey is used to apply a component label to identity lease objects, indicating:
	//   1. the lease is an identity lease (different from leader election leases)
	//   2. which component owns this lease
	IdentityLeaseComponentLabelKey = "apiserver.kubernetes.io/identity"

	// KubeAPIServer defines variable used internally when referring to kube-apiserver component
	KubeAPIServer = "kube-apiserver"
)

// BuildGenericConfig takes the master server options and produces the genericapiserver.Config associated with it.
func BuildGenericConfig(
	s controlplaneoptions.CompletedOptions,
	schemes []*runtime.Scheme,
	getOpenAPIDefinitions func(ref openapicommon.ReferenceCallback) map[string]openapicommon.OpenAPIDefinition,
) (
	genericConfig *genericapiserver.RecommendedConfig,
	kubeSharedInformers informers.SharedInformerFactory,
	storageFactory *serverstorage.DefaultStorageFactory,
	lastErr error,
) {
	genericConfig = genericapiserver.NewRecommendedConfig(legacyscheme.Codecs)
	genericConfig.MergedResourceConfig = controlplane.DefaultAPIResourceConfigSource()

	if lastErr = s.GenericServerRunOptions.ApplyTo(&genericConfig.Config); lastErr != nil {
		return
	}

	s.RecommendedOptions.ExtraAdmissionInitializers = func(c *genericapiserver.RecommendedConfig) ([]admission.PluginInitializer, error) {
		client, err := clientset.NewForConfig(c.LoopbackClientConfig)
		if err != nil {
			return nil, err
		}
		informerFactory := informers.NewSharedInformerFactory(client, c.LoopbackClientConfig.Timeout)
		s.InternalVersionedInformers = informerFactory
		return []admission.PluginInitializer{initializer.New(informerFactory, client)}, nil
	}

	// RecommendedOptions.ApplyTo must after RecommendedOptions.ExtraAdmissionInitializers.
	// Because RecommendedOptions.ApplyTo need init ExtraAdmissionInitializers.
	if lastErr = s.RecommendedOptions.ApplyTo(genericConfig); lastErr != nil {
		return
	}

	// Use protobufs for self-communication.
	// Since not every generic apiserver has to support protobufs, we
	// cannot default to it in generic apiserver and need to explicitly
	// set it in onex-apiserver.
	genericConfig.LoopbackClientConfig.ContentConfig.ContentType = "application/vnd.kubernetes.protobuf"
	// Disable compression for self-communication, since we are going to be
	// on a fast local network
	genericConfig.LoopbackClientConfig.DisableCompression = true

	loopbackClientConfig := genericConfig.LoopbackClientConfig

	// Build kubernetes client
	// Use onex's config to mock a kubernetes client.
	kubeClient, err := clientset.NewForConfig(loopbackClientConfig)
	if err != nil {
		lastErr = fmt.Errorf("failed to create real external clientset: %v", err)
		return
	}
	kubeSharedInformers = informers.NewSharedInformerFactory(kubeClient, loopbackClientConfig.Timeout)

	if lastErr = s.Features.ApplyTo(&genericConfig.Config, kubeClient, kubeSharedInformers); lastErr != nil {
		return
	}
	if lastErr = s.APIEnablement.ApplyTo(&genericConfig.Config, controlplane.DefaultAPIResourceConfigSource(), legacyscheme.Scheme); lastErr != nil {
		return
	}

	if lastErr = s.Traces.ApplyTo(genericConfig.EgressSelector, &genericConfig.Config); lastErr != nil {
		return
	}

	// wrap the definitions to revert any changes from disabled features
	getOpenAPIDefinitions = openapi.GetOpenAPIDefinitionsWithoutDisabledFeatures(getOpenAPIDefinitions)
	// namer := openapinamer.NewDefinitionNamer(legacyscheme.Scheme)
	namer := openapinamer.NewDefinitionNamer(schemes...)
	genericConfig.OpenAPIConfig = genericapiserver.DefaultOpenAPIConfig(getOpenAPIDefinitions, namer)
	genericConfig.OpenAPIConfig.Info.Title = "OneX"
	genericConfig.OpenAPIConfig.Info.Version = "v0.0.1"
	genericConfig.OpenAPIV3Config = genericapiserver.DefaultOpenAPIV3Config(getOpenAPIDefinitions, namer)
	genericConfig.OpenAPIV3Config.Info.Title = "OneX"
	genericConfig.OpenAPIV3Config.Info.Version = "v0.0.1"
	// Not in use, just serving as a placeholder.
	genericConfig.LongRunningFunc = filters.BasicLongRunningRequestCheck(
		sets.NewString("watch", "proxy"),
		sets.NewString("attach", "exec", "proxy", "log", "portforward"),
	)

	if genericConfig.EgressSelector != nil {
		s.RecommendedOptions.Etcd.StorageConfig.Transport.EgressLookup = genericConfig.EgressSelector.Lookup
	}
	s.RecommendedOptions.Etcd.StorageConfig.Transport.TracerProvider = genericConfig.TracerProvider

	storageFactoryConfig := storage.NewStorageFactoryConfig()
	storageFactoryConfig.APIResourceConfig = genericConfig.MergedResourceConfig
	storageFactoryConfig.DefaultResourceEncoding.SetEffectiveVersion(genericConfig.EffectiveVersion)
	storageFactory, lastErr = storageFactoryConfig.Complete(s.RecommendedOptions.Etcd).New()
	if lastErr != nil {
		return
	}
	if lastErr = s.RecommendedOptions.Etcd.ApplyWithStorageFactoryTo(storageFactory, &genericConfig.Config); lastErr != nil {
		return
	}

	// UPDATEME: Currently authentication and authorization rely on kubernetes cluster. Support in the future.
	/*
		ctx := wait.ContextForChannel(genericConfig.DrainedNotify())

		// Authentication.ApplyTo requires already applied OpenAPIConfig and EgressSelector if present
		if lastErr = s.Authentication.ApplyTo(ctx, &genericConfig.Authentication, genericConfig.SecureServing, genericConfig.EgressSelector, genericConfig.OpenAPIConfig, genericConfig.OpenAPIV3Config, clientgoExternalClient, versionedInformers, genericConfig.APIServerID); lastErr != nil {
			return
		}

		var enablesRBAC bool
		genericConfig.Authorization.Authorizer, genericConfig.RuleResolver, enablesRBAC, err = BuildAuthorizer(
			ctx,
			s,
			genericConfig.EgressSelector,
			genericConfig.APIServerID,
			versionedInformers,
		)
		if err != nil {
			lastErr = fmt.Errorf("invalid authorization config: %v", err)
			return
		}
		if s.Authorization != nil && !enablesRBAC {
			genericConfig.DisabledPostStartHooks.Insert(rbacrest.PostStartHookName)
		}
	*/

	lastErr = s.RecommendedOptions.Audit.ApplyTo(&genericConfig.Config)
	if lastErr != nil {
		return
	}

	genericConfig.AggregatedDiscoveryGroupManager = aggregated.NewResourceManager("apis")
	return
}

// CreatePeerEndpointLeaseReconciler creates a apiserver endpoint lease reconciliation loop
// The peer endpoint leases are used to find network locations of apiservers for peer proxy
func CreatePeerEndpointLeaseReconciler(c genericapiserver.Config, storageFactory serverstorage.StorageFactory) (reconcilers.PeerEndpointLeaseReconciler, error) {
	ttl := DefaultPeerEndpointReconcilerTTL
	config, err := storageFactory.NewConfig(api.Resource("apiServerPeerIPInfo"), &api.Endpoints{})
	if err != nil {
		return nil, fmt.Errorf("error creating storage factory config: %w", err)
	}
	reconciler, err := reconcilers.NewPeerEndpointLeaseReconciler(config, "/peerserverleases/", ttl)
	return reconciler, err
}

func BuildPeerProxy(
	leaseInformer coordinationv1informers.LeaseInformer,
	loopbackClientConfig *rest.Config,
	proxyClientCertFile string,
	proxyClientKeyFile string,
	peerCAFile string,
	peerAdvertiseAddress reconcilers.PeerAdvertiseAddress,
	apiServerID string,
	reconciler reconcilers.PeerEndpointLeaseReconciler,
	serializer runtime.NegotiatedSerializer,
) (utilpeerproxy.Interface, error) {
	if proxyClientCertFile == "" {
		return nil, fmt.Errorf("error building peer proxy handler, proxy-cert-file not specified")
	}
	if proxyClientKeyFile == "" {
		return nil, fmt.Errorf("error building peer proxy handler, proxy-key-file not specified")
	}

	proxyClientConfig := &transport.Config{
		TLS: transport.TLSConfig{
			Insecure:   false,
			CertFile:   proxyClientCertFile,
			KeyFile:    proxyClientKeyFile,
			CAFile:     peerCAFile,
			ServerName: "kubernetes.default.svc",
		},
	}

	return utilpeerproxy.NewPeerProxyHandler(
		apiServerID,
		IdentityLeaseComponentLabelKey+"="+KubeAPIServer,
		leaseInformer,
		reconciler,
		serializer,
		loopbackClientConfig,
		proxyClientConfig,
	)
}

func convertVersion(info version.Info) *kversion.Info {
	v, _ := semver.Make(info.GitVersion)
	return &kversion.Info{
		Major:        strconv.FormatUint(v.Major, 10),
		Minor:        strconv.FormatUint(v.Minor, 10),
		GitVersion:   info.GitVersion,
		GitCommit:    info.GitCommit,
		GitTreeState: info.GitTreeState,
		BuildDate:    info.BuildDate,
		GoVersion:    info.GoVersion,
		Compiler:     info.Compiler,
		Platform:     info.Platform,
	}
}
