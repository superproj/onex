// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

// Package app implements a server that runs a set of active components.
package app

import (
	"context"
	"fmt"
	"os"

	"github.com/jinzhu/copier"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
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
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/superproj/onex/cmd/onex-miner-controller/app/config"
	"github.com/superproj/onex/cmd/onex-miner-controller/app/options"
	onexcontroller "github.com/superproj/onex/internal/controller"
	"github.com/superproj/onex/internal/pkg/util/ratelimiter"
	"github.com/superproj/onex/pkg/apis/apps/v1beta1"
	"github.com/superproj/onex/pkg/apis/apps/v1beta1/index"
	"github.com/superproj/onex/pkg/db"
	"github.com/superproj/onex/pkg/record"
	"github.com/superproj/onex/pkg/version"
)

const appName = "onex-miner-controller"

func init() {
	utilruntime.Must(logsapi.AddFeatureGates(utilfeature.DefaultMutableFeatureGate))
	utilruntime.Must(features.AddFeatureGates(utilfeature.DefaultMutableFeatureGate))
}

// NewControllerCommand creates a *cobra.Command object with default parameters.
func NewControllerCommand() *cobra.Command {
	o, err := options.NewOptions()
	if err != nil {
		klog.Background().Error(err, "Unable to initialize command options")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	cmd := &cobra.Command{
		Use: appName,
		Long: `The cloud miner controller is a daemon that embeds
the core control loops. In applications of robotics and
automation, a control loop is a non-terminating loop that regulates the state of
the system. In OneX, a controller is a control loop that watches the shared
state of the miner through the onex-apiserver and makes changes attempting to move the
current state towards the desired state.`,
		PersistentPreRunE: func(*cobra.Command, []string) error {
			// silence client-go warnings.
			// onex-miner-controller generically watches APIs (including deprecated ones),
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
			ctrl.SetLogger(klog.Background())

			cliflag.PrintFlags(cmd.Flags())

			// klog.Background will automatically use the right logger. Here use the
			// global klog.logging initialized by `logsapi.ValidateAndApply`.
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

// Run runs the controller options. This should never exit.
func Run(ctx context.Context, c *config.CompletedConfig) error {
	// To help debugging, immediately log version
	klog.InfoS("Starting miner controller", "version", version.Get().String())

	klog.InfoS("Golang settings", "GOGC", os.Getenv("GOGC"), "GOMAXPROCS", os.Getenv("GOMAXPROCS"), "GOTRACEBACK", os.Getenv("GOTRACEBACK"))

	// Do some initialization here
	var redisOptions db.RedisOptions
	_ = copier.Copy(&redisOptions, &c.ComponentConfig.Redis)
	rdb, err := db.NewRedis(&redisOptions)
	if err != nil {
		return err
	}

	var watchNamespaces map[string]cache.Config
	if c.ComponentConfig.Namespace != "" {
		watchNamespaces = map[string]cache.Config{
			c.ComponentConfig.Namespace: {},
		}
	}

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := ctrl.NewManager(c.Kubeconfig, ctrl.Options{
		LeaderElection:             c.ComponentConfig.LeaderElection.LeaderElect,
		LeaderElectionID:           c.ComponentConfig.LeaderElection.ResourceName,
		LeaseDuration:              &c.ComponentConfig.LeaderElection.LeaseDuration.Duration,
		RenewDeadline:              &c.ComponentConfig.LeaderElection.RenewDeadline.Duration,
		RetryPeriod:                &c.ComponentConfig.LeaderElection.RetryPeriod.Duration,
		LeaderElectionResourceLock: c.ComponentConfig.LeaderElection.ResourceLock,
		LeaderElectionNamespace:    c.ComponentConfig.LeaderElection.ResourceNamespace,
		HealthProbeBindAddress:     c.ComponentConfig.HealthzBindAddress,
		Metrics: metricsserver.Options{
			SecureServing: false,
			BindAddress:   c.ComponentConfig.MetricsBindAddress,
		},
		Cache: cache.Options{
			DefaultNamespaces: watchNamespaces,
			SyncPeriod:        &c.ComponentConfig.SyncPeriod.Duration,
		},
	})
	if err != nil {
		klog.ErrorS(err, "Unable to new miner controller")
		return err
	}

	// applies all the stored functions to the scheme created by controller-runtime
	_ = v1beta1.AddToScheme(mgr.GetScheme())
	_ = corev1.AddToScheme(mgr.GetScheme())

	// Initialize event recorder.
	record.InitFromRecorder(mgr.GetEventRecorderFor("onex-miner-controller"))

	if err := index.AddDefaultIndexes(ctx, mgr); err != nil {
		klog.ErrorS(err, "Unable to setup indexes")
		return err
	}

	if !c.ComponentConfig.DryRun {
		// controller-runtime for multi-cluster support, reference:
		// https://github.com/kubernetes-sigs/controller-runtime/blob/main/designs/move-cluster-specific-code-out-of-manager.md
		if err := mgr.Add(c.ProviderCluster); err != nil {
			return err
		}
	}

	if err = (&onexcontroller.MinerReconciler{
		DryRun:           c.ComponentConfig.DryRun,
		ProviderClient:   c.ProviderClient,
		ProviderCluster:  c.ProviderCluster,
		RedisClient:      rdb,
		ComponentConfig:  c.ComponentConfig,
		WatchFilterValue: c.ComponentConfig.WatchFilterValue,
	}).SetupWithManager(ctx, mgr, controller.Options{
		MaxConcurrentReconciles: int(c.ComponentConfig.Parallelism),
		RecoverPanic:            ptr.To(true),
		RateLimiter:             ratelimiter.DefaultControllerRateLimiter(),
	}); err != nil {
		klog.ErrorS(err, "Unable to create controller", "controller", "miner")
		return err
	}

	// add handlers
	if err := mgr.AddReadyzCheck("healthz", healthz.Ping); err != nil {
		klog.ErrorS(err, "Unable to set up health check")
		return err
	}

	if err := mgr.AddHealthzCheck("readyz", healthz.Ping); err != nil {
		klog.ErrorS(err, "Unable to set up ready check")
		return err
	}

	return mgr.Start(ctx)
}
