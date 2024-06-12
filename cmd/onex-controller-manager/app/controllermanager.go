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

	"github.com/jinzhu/copier"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/client-go/metadata"
	restclient "k8s.io/client-go/rest"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/cli/globalflag"
	"k8s.io/component-base/logs"
	logsapi "k8s.io/component-base/logs/api/v1"
	"k8s.io/component-base/metrics/features"
	"k8s.io/component-base/term"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	ctrlmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/superproj/onex/cmd/onex-controller-manager/app/cleaner"
	"github.com/superproj/onex/cmd/onex-controller-manager/app/config"
	"github.com/superproj/onex/cmd/onex-controller-manager/app/options"
	onexcontroller "github.com/superproj/onex/internal/controller"
	configv1beta1 "github.com/superproj/onex/internal/controller/apis/config/v1beta1"
	"github.com/superproj/onex/internal/gateway/store"
	"github.com/superproj/onex/internal/pkg/metrics"
	"github.com/superproj/onex/internal/pkg/util/ratelimiter"
	"github.com/superproj/onex/internal/webhooks"
	v1beta1 "github.com/superproj/onex/pkg/apis/apps/v1beta1"
	apiv1 "github.com/superproj/onex/pkg/apis/core/v1"
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

			if err := o.Validate(); err != nil {
				return err
			}

			c, err := o.Config()
			if err != nil {
				return err
			}

			cc := c.Complete()
			if err := options.LogOrWriteConfig(o.WriteConfigTo, cc.ComponentConfig); err != nil {
				return err
			}

			// add feature enablement metrics
			utilfeature.DefaultMutableFeatureGate.AddMetrics()
			return Run(context.Background(), cc)
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
	namedFlagSets := o.Flags()
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

	// setup resource cleaner controller
	clean := newCleaner(mgr.GetClient(), storeClient, &cleaner.Miner{}, &cleaner.MinerSet{}, &cleaner.Chain{})
	if err := mgr.Add(clean); err != nil {
		klog.ErrorS(err, "Unable to create resource cleaner", "controller", "ResourceCleaner")
		return err
	}

	setupChecks(mgr)

	setupReconcilers(ctx, c, storeClient, mgr)

	return mgr.Start(ctx)
}

func setupChecks(mgr ctrl.Manager) {
	// add handlers
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		klog.Exitf("Unable to create health check: %v", err)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		klog.Exitf("Unable to create ready check: %v", err)
	}

	/*
		if err := mgr.AddHealthzCheck("healthz", mgr.GetWebhookServer().StartedChecker()); err != nil {
			klog.Exitf("Unable to create health check: %v", err)
		}

		if err := mgr.AddReadyzCheck("readyz", mgr.GetWebhookServer().StartedChecker()); err != nil {
			klog.Exitf("Unable to create ready check: %v", err)
		}
	*/
}

func setupReconcilers(ctx context.Context, c *config.CompletedConfig, storeClient store.IStore, mgr ctrl.Manager) {
	// setup garbage collector controller
	gc := &garbageCollector{completedConfig: c}
	if err := mgr.Add(gc); err != nil {
		klog.Exitf("Unable to create GarbageCollector controller: %v", err)
	}

	defaultOptions := controller.Options{
		MaxConcurrentReconciles: int(c.ComponentConfig.Generic.Parallelism),
		RecoverPanic:            ptr.To(true),
		RateLimiter:             ratelimiter.DefaultControllerRateLimiter(),
	}

	// setup chain controller
	if err := (&onexcontroller.ChainReconciler{
		ComponentConfig:  &c.ComponentConfig.ChainController,
		WatchFilterValue: c.ComponentConfig.Generic.WatchFilterValue,
	}).SetupWithManager(ctx, mgr, defaultOptions); err != nil {
		klog.Exitf("Unable to create Chain controller: %v", err)
	}

	// setup sync controller
	if err := (&onexcontroller.SyncReconciler{
		Store: storeClient,
	}).SetupWithManager(ctx, mgr, defaultOptions); err != nil {
		klog.Exitf("Unable to create Sync controller: %v", err)
	}

	metadataClient, err := metadata.NewForConfig(c.Kubeconfig)
	if err != nil {
		klog.Exitf("Failed to create metadata client: %v", err)
	}

	if err := (&onexcontroller.NamespacedResourcesDeleterReconciler{
		Client:         c.Client,
		MetadataClient: metadataClient,
	}).SetupWithManager(ctx, mgr, defaultOptions); err != nil {
		klog.Exitf("Unable to create Namespace controller: %v", err)
	}
}

//nolint:unused
func setupWebhooks(mgr ctrl.Manager) {
	if err := (&webhooks.Chain{}).SetupWebhookWithManager(mgr); err != nil {
		klog.Exitf("Unable to create Chain webhook: %v", err)
	}
}
