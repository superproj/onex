// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

// Package options contains flags and options for initializing an apiserver
package options

import (
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apiserver/pkg/admission"
	genericoptions "k8s.io/apiserver/pkg/server/options"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/logs"
	logsapi "k8s.io/component-base/logs/api/v1"
	"k8s.io/component-base/metrics"
	"k8s.io/kubernetes/pkg/api/legacyscheme"

	"github.com/superproj/onex/internal/pkg/options"
	"github.com/superproj/onex/pkg/apis/apps/v1beta1"
	"github.com/superproj/onex/pkg/generated/informers"
)

const defaultEtcdPathPrefix = "/registry/onex.io"

// ServerRunOptions contains state for master/api server.
type ServerRunOptions struct {
	// RecommendedOptions *genericoptions.RecommendedOptions
	GenericServerRunOptions *genericoptions.ServerRunOptions
	RecommendedOptions      *options.RecommendedOptions
	Features                *genericoptions.FeatureOptions
	Metrics                 *metrics.Options
	Logs                    *logs.Options
	Traces                  *genericoptions.TracingOptions
	// CloudOptions            *cloud.CloudOptions

	EnableLogsHandler bool
	EventTTL          time.Duration

	SharedInformerFactory informers.SharedInformerFactory
}

// NewServerRunOptions returns a new ServerRunOptions.
func NewServerRunOptions() *ServerRunOptions {
	o := &ServerRunOptions{
		GenericServerRunOptions: genericoptions.NewServerRunOptions(),
		RecommendedOptions: options.NewRecommendedOptions(
			defaultEtcdPathPrefix,
			legacyscheme.Codecs.LegacyCodec(v1beta1.SchemeGroupVersion),
		),
		Features: genericoptions.NewFeatureOptions(),
		Metrics:  metrics.NewOptions(),
		Logs:     logs.NewOptions(),
		Traces:   genericoptions.NewTracingOptions(),

		EnableLogsHandler: true,
		EventTTL:          1 * time.Hour,
		// CloudOptions: cloud.NewCloudOptions(),
	}

	o.RecommendedOptions.Etcd.StorageConfig.EncodeVersioner = runtime.NewMultiGroupVersioner(
		v1beta1.SchemeGroupVersion,
		schema.GroupKind{Group: v1beta1.GroupName},
	)

	// Redirect the certificates output directory to avoid creating the "apiserver.local.config" directory in the root directory
	// and keep the root directory clean.
	o.RecommendedOptions.SecureServing.ServerCert.CertDirectory = "_output/certificates"

	// the following three lines remove dependence with kube-apiserver
	o.RecommendedOptions.Authorization = nil
	o.RecommendedOptions.CoreAPI = nil
	// We only register the plugin of onex-apiserver,
	// so we need to clear the plugin registered by apiserver by default.
	o.RecommendedOptions.Admission.Plugins = admission.NewPlugins()

	// register all custom dmission plugins
	RegisterAllAdmissionPlugins(o.RecommendedOptions.Admission.Plugins)
	o.RecommendedOptions.Admission.RecommendedPluginOrder = AllOrderedPlugins
	o.RecommendedOptions.Admission.DefaultOffPlugins = DefaultOffAdmissionPlugins()

	// Overwrite the default for storage data format.
	o.RecommendedOptions.Etcd.DefaultStorageMediaType = "application/vnd.kubernetes.protobuf"
	return o
}

func (o ServerRunOptions) Flags() (fss cliflag.NamedFlagSets) {
	o.GenericServerRunOptions.AddUniversalFlags(fss.FlagSet("generic"))
	o.RecommendedOptions.AddFlags(fss.FlagSet("recommended"))
	o.Features.AddFlags(fss.FlagSet("features"))
	o.Metrics.AddFlags(fss.FlagSet("metrics"))
	logsapi.AddFlags(o.Logs, fss.FlagSet("logs"))
	o.Traces.AddFlags(fss.FlagSet("traces"))
	// o.CloudOptions.AddFlags(fss.FlagSet("cloud"))

	// Note: the weird ""+ in below lines seems to be the only way to get gofmt to
	// arrange these text blocks sensibly. Grrr.
	fs := fss.FlagSet("misc")
	fs.DurationVar(&o.EventTTL, "event-ttl", o.EventTTL,
		"Amount of time to retain events.")

	fs.BoolVar(&o.EnableLogsHandler, "enable-logs-handler", o.EnableLogsHandler,
		"If true, install a /logs handler for the apiserver logs.")
	_ = fs.MarkDeprecated("enable-logs-handler", "This flag will be removed in v1.19")

	return fss
}

// Validate validates ServerRunOptions.
func (o ServerRunOptions) Validate(args []string) error {
	errors := []error{}
	errors = append(errors, o.RecommendedOptions.Validate()...)
	// errors = append(errors, o.CloudOptions.Validate()...)
	return utilerrors.NewAggregate(errors)
}
