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
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/blang/semver/v4"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	kversion "k8s.io/apimachinery/pkg/version"
	"k8s.io/apiserver/pkg/admission"
	admissioninitializer "k8s.io/apiserver/pkg/admission/initializer"
	webhookinitializer "k8s.io/apiserver/pkg/admission/plugin/webhook/initializer"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	utilflowcontrol "k8s.io/apiserver/pkg/util/flowcontrol"
	"k8s.io/apiserver/pkg/util/webhook"
	"k8s.io/client-go/rest"
	aggregatorapiserver "k8s.io/kube-aggregator/pkg/apiserver"

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

		// Webhook admission plugins resolve `service:` references through a
		// service resolver and authenticate outbound calls via an
		// authentication-info resolver wrapper. Both are independent of the
		// onex custom initializer, so wire dedicated initializers for them.
		serviceResolver := aggregatorapiserver.NewClusterIPServiceResolver(
			informerFactory.Core().V1().Services().Lister(),
		)
		if localHost, err := url.Parse(c.LoopbackClientConfig.Host); err == nil {
			serviceResolver = aggregatorapiserver.NewLoopbackServiceResolver(serviceResolver, localHost)
		}

		authInfoResolverWrapper := webhook.NewDefaultAuthenticationInfoResolverWrapper(
			nil, // proxyTransport: not wired yet, outgoing webhook calls use the default transport
			c.EgressSelector,
			c.LoopbackClientConfig,
			c.TracerProvider,
		)
		webhookInitializer := webhookinitializer.NewPluginInitializer(authInfoResolverWrapper, serviceResolver)

		return []admission.PluginInitializer{
			// onex's own plugins (namespace lifecycle/autoprovision/exists) bind
			// through the custom WantsInternalInformerFactory/ClientSet interfaces.
			initializer.New(informerFactory, client),
			// The webhook plugins bind through the standard apiserver
			// WantsExternalKubeInformerFactory/ClientSet interfaces.
			admissioninitializer.New(
				client,                         // externalClient
				nil,                            // dynamicClient
				informerFactory,                // externalInformers
				nil,                            // authorizer is wired separately (BuildAuthorizer)
				utilfeature.DefaultFeatureGate, // featureGates
				nil,                            // effectiveVersion
				nil,                            // drained notification (only needed for manifest-based static webhooks)
				nil,                            // restMapper
			),
			webhookInitializer,
		}, nil
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

	// Wire API Priority and Fairness (APF) into the request-handling filter
	// chain. The flow-control REST storage (FlowSchema/PriorityLevelConfiguration
	// CRUD) and its bootstrap ensurer are already installed via the REST storage
	// provider; setting FlowControl here is what actually routes requests through
	// the APF queue-set dispatch. Until the flow-control informers sync, APF falls
	// back to its built-in catch-all configuration, so requests are not dropped
	// during startup.
	genericConfig.FlowControl = utilflowcontrol.New(
		s.InternalVersionedInformers,
		kubeClient.FlowcontrolV1(),
		genericConfig.MaxRequestsInFlight+genericConfig.MaxMutatingRequestsInFlight,
	)

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

	// Authentication is already wired by s.RecommendedOptions.ApplyTo above:
	// without an explicit client CA or token source, requests fall back to the
	// anonymous user. Install onex's own non-delegating authorizer (default
	// AlwaysAllow) so access can be restricted via --authorization-mode without
	// requiring a kube cluster to serve SubjectAccessReviews.
	genericConfig.Authorization.Authorizer, genericConfig.RuleResolver, err = BuildAuthorizer(
		strings.Split(s.AuthorizationMode, ","),
		s.InternalVersionedInformers,
		webhookAuthorizationOptions{
			ConfigFile:           s.AuthorizationWebhookConfigFile,
			CacheAuthorizedTTL:   s.AuthorizationWebhookCacheAuthorizedTTL,
			CacheUnauthorizedTTL: s.AuthorizationWebhookCacheUnauthorizedTTL,
		},
	)
	if err != nil {
		lastErr = fmt.Errorf("invalid authorization config: %w", err)
		return
	}

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
