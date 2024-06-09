// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

// Package app implements a server that runs a set of active
// components. This includes sync controllers, chains and namespace.
package app

import (
	"context"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/jinzhu/copier"
	"github.com/spf13/cobra"
	apiv1 "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/selection"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	genericapiserver "k8s.io/apiserver/pkg/server"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	cacheddiscovery "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/metadata"
	"k8s.io/client-go/metadata/metadatainformer"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/cli/globalflag"
	"k8s.io/component-base/configz"
	"k8s.io/component-base/featuregate"
	"k8s.io/component-base/logs"
	logsapi "k8s.io/component-base/logs/api/v1"
	"k8s.io/component-base/metrics/features"
	controllersmetrics "k8s.io/component-base/metrics/prometheus/controllers"
	"k8s.io/component-base/term"
	genericcontrollermanager "k8s.io/controller-manager/app"
	"k8s.io/controller-manager/pkg/clientbuilder"
	"k8s.io/controller-manager/pkg/informerfactory"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/controller/garbagecollector"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	ctrlmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/superproj/onex/cmd/onex-controller-manager/app/config"
	"github.com/superproj/onex/cmd/onex-controller-manager/app/options"
	"github.com/superproj/onex/cmd/onex-controller-manager/names"
	configv1beta1 "github.com/superproj/onex/internal/controller/apis/config/v1beta1"
	"github.com/superproj/onex/internal/gateway/store"
	"github.com/superproj/onex/internal/pkg/metrics"
	"github.com/superproj/onex/internal/pkg/util/ratelimiter"
	"github.com/superproj/onex/internal/webhooks"
	v1beta1 "github.com/superproj/onex/pkg/apis/apps/v1beta1"
	"github.com/superproj/onex/pkg/db"
	"github.com/superproj/onex/pkg/record"
	"github.com/superproj/onex/pkg/version"
)

const appName = "onex-controller-manager"

var scheme = runtime.NewScheme()

func init() {
	utilruntime.Must(logsapi.AddFeatureGates(utilfeature.DefaultMutableFeatureGate))
	utilruntime.Must(features.AddFeatureGates(utilfeature.DefaultMutableFeatureGate))

	// applies all the stored functions to the scheme created by controller-runtime
	_ = apiv1.AddToScheme(scheme)
	_ = v1beta1.AddToScheme(scheme)
	_ = configv1beta1.AddToScheme(scheme)
	// _ = corev1.AddToScheme(scheme)
}

const (
	// ControllerStartJitter is the Jitter used when starting controller managers
	ControllerStartJitter = 1.0
	// ConfigzName is the name used for register onex-controller manager /configz, same with GroupName.
	ConfigzName = "onexcontrollermanager.config.onex.io"
)

// NewControllerManagerCommand creates a *cobra.Command object with default parameters.
func NewControllerManagerCommand() *cobra.Command {
	o, err := options.NewOptions()
	if err != nil {
		klog.Background().Error(err, "Unable to initialize command options")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	cmd := &cobra.Command{
		Use: appName,
		Long: `The onex controller manager is a daemon that embeds
the core control loops. In applications of robotics and
automation, a control loop is a non-terminating loop that regulates the state of
the system. In OneX , a controller is a control loop that watches the shared
state of the miner through the onex-apiserver and makes changes attempting to move the
current state towards the desired state.`,
		PersistentPreRunE: func(*cobra.Command, []string) error {
			// silence client-go warnings.
			// onex-controller-manager generically watches APIs (including deprecated ones),
			// and CI ensures it works properly against matching onex-apiserver versions.
			restclient.SetDefaultWarningHandler(restclient.NoWarnings{})
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			version.PrintAndExitIfRequested(appName)

			// Activate logging as soon as possible, after that
			// show flags with the final logging configuration.
			if err := logsapi.ValidateAndApply(o.Logs, utilfeature.DefaultFeatureGate); err != nil {
				return err
			}
			// klog.Background will automatically use the right logger. Here use the
			// global klog.logging initialized by `logsapi.ValidateAndApply`.
			ctrl.SetLogger(klog.Background())

			cliflag.PrintFlags(cmd.Flags())

			if err := o.Complete(); err != nil {
				return err
			}

			allControllers, disabledControllers, controllerAliases := KnownControllers(), DisabledControllers(), ControllerAliases()
			if err := o.Validate(allControllers, disabledControllers, controllerAliases); err != nil {
				return err
			}

			c, err := o.Config(allControllers, disabledControllers, controllerAliases)
			if err != nil {
				return err
			}

			cc := c.Complete()
			if err := options.LogOrWriteConfig(o.WriteConfigTo, cc.ComponentConfig); err != nil {
				return err
			}

			// add feature enablement metrics
			utilfeature.DefaultMutableFeatureGate.AddMetrics()
			return Run(genericapiserver.SetupSignalContext(), cc)
		},
		Args: func(cmd *cobra.Command, args []string) error {
			for _, arg := range args {
				if len(arg) > 0 {
					return fmt.Errorf("%q does not take any arguments, got %q", cmd.CommandPath(), args)
				}
			}
			return nil
		},
	}

	fs := cmd.Flags()
	namedFlagSets := o.Flags(KnownControllers(), DisabledControllers(), ControllerAliases())
	version.AddFlags(namedFlagSets.FlagSet("global"))
	globalflag.AddGlobalFlags(namedFlagSets.FlagSet("global"), cmd.Name(), logs.SkipLoggingConfigurationFlags())
	for _, f := range namedFlagSets.FlagSets {
		fs.AddFlagSet(f)
	}

	cols, _, _ := term.TerminalSize(cmd.OutOrStdout())
	cliflag.SetUsageAndHelpFunc(cmd, namedFlagSets, cols)

	if err := cmd.MarkFlagFilename("config", "yaml", "yml", "json"); err != nil {
		klog.Background().Error(err, "Failed to mark flag filename")
	}

	return cmd
}

// Run runs the controller manager options. This should never exit.
func Run(ctx context.Context, c *config.CompletedConfig) error {
	// To help debugging, immediately log version
	klog.InfoS("Starting controller manager", "version", version.Get().String())

	klog.InfoS("Golang settings", "GOGC", os.Getenv("GOGC"), "GOMAXPROCS", os.Getenv("GOMAXPROCS"), "GOTRACEBACK", os.Getenv("GOTRACEBACK"))

	// Store controller configs
	cfgz, err := configz.New(ConfigzName)
	if err != nil {
		klog.ErrorS(err, "Unable to register configz")
		return err
	}
	cfgz.Set(c.ComponentConfig)

	// Do some initialization here
	var mysqlOptions db.MySQLOptions
	_ = copier.Copy(&mysqlOptions, c.ComponentConfig.Generic.MySQL)
	storeClient, err := wireStoreClient(&mysqlOptions)
	if err != nil {
		return err
	}

	var watchNamespaces map[string]cache.Config
	if c.ComponentConfig.Generic.Namespace != "" {
		watchNamespaces = map[string]cache.Config{
			c.ComponentConfig.Generic.Namespace: {},
		}
	}

	req, _ := labels.NewRequirement(v1beta1.ChainNameLabel, selection.Exists, nil)
	chainSecretCacheSelector := labels.NewSelector().Add(*req)

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := ctrl.NewManager(c.Kubeconfig, ctrl.Options{
		Scheme: scheme,
		// Metrics:                    c.ComponentConfig.Generic.MetricsBindAddress,
		LeaderElection:             c.ComponentConfig.Generic.LeaderElection.LeaderElect,
		LeaderElectionID:           c.ComponentConfig.Generic.LeaderElection.ResourceName,
		LeaseDuration:              &c.ComponentConfig.Generic.LeaderElection.LeaseDuration.Duration,
		RenewDeadline:              &c.ComponentConfig.Generic.LeaderElection.RenewDeadline.Duration,
		RetryPeriod:                &c.ComponentConfig.Generic.LeaderElection.RetryPeriod.Duration,
		LeaderElectionResourceLock: c.ComponentConfig.Generic.LeaderElection.ResourceLock,
		LeaderElectionNamespace:    c.ComponentConfig.Generic.LeaderElection.ResourceNamespace,
		HealthProbeBindAddress:     c.ComponentConfig.Generic.HealthzBindAddress,
		PprofBindAddress:           c.ComponentConfig.Generic.PprofBindAddress,
		Cache: cache.Options{
			DefaultNamespaces: watchNamespaces,
			SyncPeriod:        &c.ComponentConfig.Generic.SyncPeriod.Duration,
			ByObject: map[client.Object]cache.ByObject{
				// Note: Only Secrets with the cluster name label are cached.
				// The default client of the manager won't use the cache for secrets at all (see Client.Cache.DisableFor).
				// The cached secrets will only be used by the secretCachingClient we create below.
				// &corev1.Secret{}: {
				// Label: clusterSecretCacheSelector,
				// },
				&corev1.ConfigMap{}: {Label: chainSecretCacheSelector},
				&corev1.Secret{}:    {Label: chainSecretCacheSelector},
			},
		},
		Client: client.Options{
			Cache: &client.CacheOptions{
				DisableFor: []client.Object{
					&corev1.ConfigMap{},
					&corev1.Secret{},
				},
			},
		},
		WebhookServer: webhook.NewServer(
			webhook.Options{
				// Port:    webhookPort,
				// CertDir: webhookCertDir,
				// TLSOpts: tlsOptionOverrides,
			},
		),
	})
	if err != nil {
		klog.ErrorS(err, "Unable to new controller manager")
		return err
	}

	machineMetricsCollector := metrics.NewMinerCollector(mgr.GetClient(), c.ComponentConfig.Generic.Namespace)
	ctrlmetrics.Registry.MustRegister(machineMetricsCollector)

	// Initialize event recorder.
	record.InitFromRecorder(mgr.GetEventRecorderFor("onex-controller-manager"))

	// Start to register controllers.
	clientBuilder, rootClientBuilder := createClientBuilders(c)

	cctx, err := CreateControllerContext(ctx, c, rootClientBuilder, clientBuilder, storeClient)
	if err != nil {
		klog.ErrorS(err, "Error building controller context")
		return err
	}

	if err := setupChecks(mgr); err != nil {
		return err
	}

	if err := addControllers(ctx, cctx, mgr, NewControllerDescriptors()); err != nil {
		return err
	}

	cctx.InformerFactory.Start(ctx.Done())
	cctx.ObjectOrMetadataInformerFactory.Start(ctx.Done())
	close(cctx.InformersStarted)

	return mgr.Start(ctx)
}

func addControllers(ctx context.Context, cctx ControllerContext, mgr ctrl.Manager, controllerDescriptors map[string]*ControllerDescriptor) error {
	// Each controller is passed a context where the logger has the name of
	// the controller set through WithName. That name then becomes the prefix of
	// of all log messages emitted by that controller.
	for _, controllerDesc := range controllerDescriptors {
		if controllerDesc.RequiresSpecialHandling() {
			continue
		}

		if err := addController(ctx, cctx, mgr, controllerDesc); err != nil {
			return err
		}
	}

	return nil
}

func addController(ctx context.Context, cctx ControllerContext, mgr ctrl.Manager, controllerDescriptor *ControllerDescriptor) error {
	controllerName := controllerDescriptor.Name()

	for _, featureGate := range controllerDescriptor.GetRequiredFeatureGates() {
		if !utilfeature.DefaultFeatureGate.Enabled(featureGate) {
			klog.InfoS("Controller is disabled by a feature gate", "controller", controllerName, "requiredFeatureGates", controllerDescriptor.GetRequiredFeatureGates())
			return nil
		}
	}

	if !cctx.IsControllerEnabled(controllerDescriptor) {
		klog.InfoS("Warning: controller is disabled", "controller", controllerName)
		return nil
	}

	klog.V(1).InfoS("Starting controller", "controller", controllerName)

	addFunc := controllerDescriptor.GetAddFunc()
	enabled, err := addFunc(klog.NewContext(ctx, klog.LoggerWithName(klog.Background(), controllerName)), mgr, cctx)
	if err != nil {
		klog.ErrorS(err, "Error starting controller", "controller", controllerName)
		return err
	}
	if !enabled {
		klog.InfoS("Warning: skipping controller", "controller", controllerName)
		return nil
	}

	klog.InfoS("Register controller", "controller", controllerName)

	return nil
}

func setupChecks(mgr ctrl.Manager) error {
	// add handlers
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		klog.ErrorS(err, "Unable to create health check")
		return err
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		klog.ErrorS(err, "Unable to create ready check")
		return err
	}

	/*
		if err := mgr.AddHealthzCheck("healthz", mgr.GetWebhookServer().StartedChecker()); err != nil {
			klog.Exitf("Unable to create health check: %v", err)
		}

		if err := mgr.AddReadyzCheck("readyz", mgr.GetWebhookServer().StartedChecker()); err != nil {
			klog.Exitf("Unable to create ready check: %v", err)
		}
	*/
	return nil
}

//nolint:unused
func setupWebhooks(mgr ctrl.Manager) {
	if err := (&webhooks.Chain{}).SetupWebhookWithManager(mgr); err != nil {
		klog.Exitf("Unable to create Chain webhook: %v", err)
	}
}

// ControllerContext defines the context object for controller
type ControllerContext struct {
	// ClientBuilder will provide a client for this controller to use
	ClientBuilder clientbuilder.ControllerClientBuilder

	// InformerFactory gives access to informers for the controller.
	InformerFactory informers.SharedInformerFactory

	// ObjectOrMetadataInformerFactory gives access to informers for typed resources
	// and dynamic resources by their metadata. All generic controllers currently use
	// object metadata - if a future controller needs access to the full object this
	// would become GenericInformerFactory and take a dynamic client.
	ObjectOrMetadataInformerFactory informerfactory.InformerFactory

	// Config provides access to init options for a given controller
	Config *config.CompletedConfig

	// DeferredDiscoveryRESTMapper is a RESTMapper that will defer
	// initialization of the RESTMapper until the first mapping is
	// requested.
	RESTMapper *restmapper.DeferredDiscoveryRESTMapper

	// InformersStarted is closed after all of the controllers have been initialized and are running.  After this point it is safe,
	// for an individual controller to start the shared informers. Before it is closed, they should not.
	InformersStarted chan struct{}

	// ResyncPeriod generates a duration each time it is invoked; this is so that
	// multiple controllers don't get into lock-step and all hammer the apiserver
	// with list requests simultaneously.
	ResyncPeriod func() time.Duration

	// ControllerManagerMetrics provides a proxy to set controller manager specific metrics.
	ControllerManagerMetrics *controllersmetrics.ControllerManagerMetrics

	// GraphBuilder gives an access to dependencyGraphBuilder which keeps tracks of resources in the cluster
	GraphBuilder *garbagecollector.GraphBuilder

	// New by OneX
	MetadataClient           metadata.Interface
	ControllerManagerOptions controller.Options
	Store                    store.IStore
}

// IsControllerEnabled checks if the context's controllers enabled or not
func (c ControllerContext) IsControllerEnabled(controllerDescriptor *ControllerDescriptor) bool {
	controllersDisabledByDefault := sets.NewString()
	if controllerDescriptor.IsDisabledByDefault() {
		controllersDisabledByDefault.Insert(controllerDescriptor.Name())
	}
	return genericcontrollermanager.IsControllerEnabled(controllerDescriptor.Name(), controllersDisabledByDefault, c.Config.ComponentConfig.Generic.Controllers)
}

// AddFunc is used to launch a particular controller. It returns a controller
// that can optionally implement other interfaces so that the controller manager
// can support the requested features.
// The returned controller may be nil, which will be considered an anonymous controller
// that requests no additional features from the controller manager.
// Any error returned will cause the controller process to `Fatal`
// The bool indicates whether the controller was enabled.
type AddFunc func(ctx context.Context, mgr ctrl.Manager, cctx ControllerContext) (enabled bool, err error)

type ControllerDescriptor struct {
	name                      string
	addFunc                   AddFunc
	requiredFeatureGates      []featuregate.Feature
	aliases                   []string
	isDisabledByDefault       bool
	isCloudProviderController bool
	requiresSpecialHandling   bool
}

func (r *ControllerDescriptor) Name() string {
	return r.name
}

func (r *ControllerDescriptor) GetAddFunc() AddFunc {
	return r.addFunc
}

func (r *ControllerDescriptor) GetRequiredFeatureGates() []featuregate.Feature {
	return append([]featuregate.Feature(nil), r.requiredFeatureGates...)
}

// GetAliases returns aliases to ensure backwards compatibility and should never be removed!
// Only addition of new aliases is allowed, and only when a canonical name is changed (please see CHANGE POLICY of controller names)
func (r *ControllerDescriptor) GetAliases() []string {
	return append([]string(nil), r.aliases...)
}

func (r *ControllerDescriptor) IsDisabledByDefault() bool {
	return r.isDisabledByDefault
}

func (r *ControllerDescriptor) IsCloudProviderController() bool {
	return r.isCloudProviderController
}

// RequiresSpecialHandling should return true only in a special non-generic controllers like ServiceAccountTokenController
func (r *ControllerDescriptor) RequiresSpecialHandling() bool {
	return r.requiresSpecialHandling
}

// KnownControllers returns all known controllers's name
func KnownControllers() []string {
	return sets.StringKeySet(NewControllerDescriptors()).List()
}

// ControllerAliases returns a mapping of aliases to canonical controller names
func ControllerAliases() map[string]string {
	aliases := map[string]string{}
	for name, c := range NewControllerDescriptors() {
		for _, alias := range c.GetAliases() {
			aliases[alias] = name
		}
	}
	return aliases
}

func DisabledControllers() []string {
	var controllersDisabledByDefault []string

	for name, c := range NewControllerDescriptors() {
		if c.IsDisabledByDefault() {
			controllersDisabledByDefault = append(controllersDisabledByDefault, name)
		}
	}

	sort.Strings(controllersDisabledByDefault)

	return controllersDisabledByDefault
}

// NewControllerDescriptors is a public map of named controller groups (you can start more than one in an init func)
// paired to their ControllerDescriptor wrapper object that includes InitFunc.
// This allows for structured downstream composition and subdivision.
func NewControllerDescriptors() map[string]*ControllerDescriptor {
	controllers := map[string]*ControllerDescriptor{}
	aliases := sets.NewString()

	// All the controllers must fulfil common constraints, or else we will explode.
	register := func(controllerDesc *ControllerDescriptor) {
		if controllerDesc == nil {
			panic("received nil controller for a registration")
		}
		name := controllerDesc.Name()
		if len(name) == 0 {
			panic("received controller without a name for a registration")
		}
		if _, found := controllers[name]; found {
			panic(fmt.Sprintf("controller name %q was registered twice", name))
		}
		if controllerDesc.GetAddFunc() == nil {
			panic(fmt.Sprintf("controller %q does not have an init function", name))
		}

		for _, alias := range controllerDesc.GetAliases() {
			if aliases.Has(alias) {
				panic(fmt.Sprintf("controller %q has a duplicate alias %q", name, alias))
			}
			aliases.Insert(alias)
		}

		controllers[name] = controllerDesc
	}

	// First add "special" controllers that aren't initialized normally. These controllers cannot be initialized
	// in the main controller loop initialization, so we add them here only for the metadata and duplication detection.
	// app.ControllerDescriptor#RequiresSpecialHandling should return true for such controllers
	// The only known special case is the ServiceAccountTokenController which *must* be started
	// first to ensure that the SA tokens for future controllers will exist. Think very carefully before adding new
	// special controllers.
	register(newGarbageCollectorControllerDescriptor())
	register(newNamespacedResourcesDeleterControllerDescriptor())
	register(newChainControllerDescriptor())
	register(newChainSyncControllerDescriptor())
	register(newMinerSetSyncControllerDescriptor())
	register(newMinerSyncControllerDescriptor())

	for _, alias := range aliases.UnsortedList() {
		if _, ok := controllers[alias]; ok {
			panic(fmt.Sprintf("alias %q conflicts with a controller name", alias))
		}
	}

	return controllers
}

// CreateControllerContext creates a context struct containing references to resources needed by the
// controllers such as clientBuilder. rootClientBuilder is only used for
// the shared-informers client.
func CreateControllerContext(
	ctx context.Context,
	s *config.CompletedConfig,
	rootClientBuilder clientbuilder.ControllerClientBuilder,
	clientBuilder clientbuilder.ControllerClientBuilder,
	storeClient store.IStore,
) (ControllerContext, error) {
	// Informer transform to trim ManagedFields for memory efficiency.
	trim := func(obj interface{}) (interface{}, error) {
		if accessor, err := meta.Accessor(obj); err == nil {
			if accessor.GetManagedFields() != nil {
				accessor.SetManagedFields(nil)
			}
		}
		return obj, nil
	}

	// In this case, we are using Kubernetes informers because the HTTP request paths and parameters
	// are ultimately the same, and the onex-apiserver can still handle the requests correctly.
	versionedClient := rootClientBuilder.ClientOrDie("shared-informers")
	sharedInformers := informers.NewSharedInformerFactoryWithOptions(versionedClient, ResyncPeriod(s)(), informers.WithTransform(trim))

	metadataClient := metadata.NewForConfigOrDie(rootClientBuilder.ConfigOrDie("metadata-informers"))
	metadataInformers := metadatainformer.NewSharedInformerFactoryWithOptions(metadataClient, ResyncPeriod(s)(), metadatainformer.WithTransform(trim))

	// If apiserver is not running we should wait for some time and fail only then. This is particularly
	// important when we start apiserver and controller manager at the same time.
	if err := genericcontrollermanager.WaitForAPIServer(versionedClient, 10*time.Second); err != nil {
		return ControllerContext{}, fmt.Errorf("failed to wait for apiserver being healthy: %w", err)
	}

	// Use a discovery client capable of being refreshed.
	discoveryClient := rootClientBuilder.DiscoveryClientOrDie("controller-discovery")
	cachedClient := cacheddiscovery.NewMemCacheClient(discoveryClient)
	restMapper := restmapper.NewDeferredDiscoveryRESTMapper(cachedClient)
	go wait.Until(func() {
		restMapper.Reset()
	}, 30*time.Second, ctx.Done())

	cctx := ControllerContext{
		ClientBuilder:                   clientBuilder,
		InformerFactory:                 sharedInformers,
		ObjectOrMetadataInformerFactory: informerfactory.NewInformerFactory(sharedInformers, metadataInformers),
		Config:                          s,
		RESTMapper:                      restMapper,
		InformersStarted:                make(chan struct{}),
		ResyncPeriod:                    ResyncPeriod(s),
		ControllerManagerMetrics:        controllersmetrics.NewControllerManagerMetrics("onex-controller-manager"),
		MetadataClient:                  metadataClient,
		Store:                           storeClient,
	}

	if cctx.Config.ComponentConfig.GarbageCollectorController.EnableGarbageCollector &&
		cctx.IsControllerEnabled(NewControllerDescriptors()[names.GarbageCollectorController]) {
		ignoredResources := make(map[schema.GroupResource]struct{})
		for _, r := range cctx.Config.ComponentConfig.GarbageCollectorController.GCIgnoredResources {
			ignoredResources[schema.GroupResource{Group: r.Group, Resource: r.Resource}] = struct{}{}
		}

		cctx.GraphBuilder = garbagecollector.NewDependencyGraphBuilder(
			ctx,
			metadataClient,
			cctx.RESTMapper,
			ignoredResources,
			cctx.ObjectOrMetadataInformerFactory,
			cctx.InformersStarted,
		)
	}

	// Added by OneX
	cctx.ControllerManagerOptions = controller.Options{
		MaxConcurrentReconciles: int(s.ComponentConfig.Generic.Parallelism),
		RecoverPanic:            ptr.To(true),
		RateLimiter:             ratelimiter.DefaultControllerRateLimiter(),
	}

	controllersmetrics.Register()
	return cctx, nil
}

// createClientBuilders creates clientBuilder and rootClientBuilder from the given configuration.
func createClientBuilders(c *config.CompletedConfig) (clientBuilder clientbuilder.ControllerClientBuilder, rootClientBuilder clientbuilder.ControllerClientBuilder) {
	rootClientBuilder = clientbuilder.SimpleControllerClientBuilder{
		ClientConfig: c.Kubeconfig,
	}

	clientBuilder = rootClientBuilder
	return
}

// ResyncPeriod returns a function which generates a duration each time it is
// invoked; this is so that multiple controllers don't get into lock-step and all
// hammer the apiserver with list requests simultaneously.
func ResyncPeriod(c *config.CompletedConfig) func() time.Duration {
	return func() time.Duration {
		// factor := rand.Float64() + 1
		// return time.Duration(float64(c.MinResyncPeriod.Nanoseconds()) * factor) // TODO?
		return 1 * time.Second
	}
}
