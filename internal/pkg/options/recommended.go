// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package options

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apiserver/pkg/admission"
	admissionmetrics "k8s.io/apiserver/pkg/admission/metrics"
	"k8s.io/apiserver/pkg/authentication/authenticatorfactory"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/apiserver/pkg/server/dynamiccertificates"
	genericoptions "k8s.io/apiserver/pkg/server/options"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	openapicommon "k8s.io/kube-openapi/pkg/common"
)

var configScheme = runtime.NewScheme()

// RecommendedOptions contains the recommended options for running an API server.
// If you add something to this list, it should be in a logical grouping.
// Each of them can be nil to leave the feature unconfigured on ApplyTo.
type RecommendedOptions struct {
	*genericoptions.RecommendedOptions
}

func NewRecommendedOptions(prefix string, codec runtime.Codec) *RecommendedOptions {
	return &RecommendedOptions{genericoptions.NewRecommendedOptions(prefix, codec)}
}

// ApplyTo adds RecommendedOptions to the server configuration.
// pluginInitializers can be empty, it is only need for additional initializers.
func (o *RecommendedOptions) ApplyTo(config *genericapiserver.RecommendedConfig) error {
	if err := o.Etcd.ApplyTo(&config.Config); err != nil {
		return err
	}
	if err := o.EgressSelector.ApplyTo(&config.Config); err != nil {
		return err
	}
	if err := o.Traces.ApplyTo(config.Config.EgressSelector, &config.Config); err != nil {
		return err
	}
	if err := o.SecureServing.ApplyTo(&config.Config.SecureServing, &config.Config.LoopbackClientConfig); err != nil {
		return err
	}
	/* UPDATEME: When add authentication and authorization features.
	if err := o.Authentication.ApplyTo(&config.Config.Authentication, config.SecureServing, config.OpenAPIConfig); err != nil {
		return err
	}
	if err := o.Authorization.ApplyTo(&config.Config.Authorization); err != nil {
		return err
	}
	*/
	if err := authenticationApplyTo(o.Authentication, &config.Config.Authentication, config.SecureServing, config.OpenAPIConfig); err != nil {
		return err
	}
	if err := o.Authorization.ApplyTo(&config.Config.Authorization); err != nil {
		return err
	}
	if err := o.Audit.ApplyTo(&config.Config); err != nil {
		return err
	}
	if err := o.CoreAPI.ApplyTo(config); err != nil {
		return err
	}

	var kubeClient *kubernetes.Clientset
	var dynamicClient *dynamic.DynamicClient
	if config.ClientConfig != nil {
		var err error
		kubeClient, err = kubernetes.NewForConfig(config.ClientConfig)
		if err != nil {
			return err
		}
		dynamicClient, err = dynamic.NewForConfig(config.ClientConfig)
		if err != nil {
			return err
		}
	}
	if err := o.Features.ApplyTo(&config.Config, kubeClient, config.SharedInformerFactory); err != nil {
		return err
	}

	initializers, err := o.ExtraAdmissionInitializers(config)
	if err != nil {
		return err
	}

	if err := o.Admission.ApplyTo(&config.Config, config.SharedInformerFactory, kubeClient, dynamicClient, o.FeatureGate,
		initializers...); err != nil {
		return err
	}

	return nil
}

// admissionOptionsApplyTo adds the admission chain to the server configuration.
// In case admission plugin names were not provided by a cluster-admin they will be prepared from the
// recommended/default values.
// In addition the method lazily initializes a generic plugin that is appended to the list of pluginInitializers
// note this method uses:
//
//	genericconfig.Authorizer
func admissionOptionsApplyTo(
	a *genericoptions.AdmissionOptions,
	c *genericapiserver.Config,
	// features featuregate.FeatureGate,
	pluginInitializers ...admission.PluginInitializer,
) error {
	if a == nil {
		return nil
	}

	pluginNames := enabledPluginNames(a)

	pluginsConfigProvider, err := admission.ReadAdmissionConfiguration(pluginNames, a.ConfigFile, configScheme)
	if err != nil {
		return fmt.Errorf("failed to read plugin config: %w", err)
	}

	initializersChain := admission.PluginInitializers{}
	initializersChain = append(initializersChain, pluginInitializers...)

	admissionChain, err := a.Plugins.NewFromPlugins(pluginNames, pluginsConfigProvider, initializersChain, a.Decorators)
	if err != nil {
		return err
	}

	c.AdmissionControl = admissionmetrics.WithStepMetrics(admissionChain)
	return nil
}

// enabledPluginNames makes use of RecommendedPluginOrder, DefaultOffPlugins,
// EnablePlugins, DisablePlugins fields
// to prepare a list of ordered plugin names that are enabled.
func enabledPluginNames(a *genericoptions.AdmissionOptions) []string {
	allOffPlugins := append(a.DefaultOffPlugins.List(), a.DisablePlugins...)
	disabledPlugins := sets.NewString(allOffPlugins...)
	enabledPlugins := sets.NewString(a.EnablePlugins...)
	disabledPlugins = disabledPlugins.Difference(enabledPlugins)
	orderedPlugins := []string{}
	for _, plugin := range a.RecommendedPluginOrder {
		if !disabledPlugins.Has(plugin) {
			orderedPlugins = append(orderedPlugins, plugin)
		}
	}

	return orderedPlugins
}

func authenticationApplyTo(s *genericoptions.DelegatingAuthenticationOptions, authenticationInfo *genericapiserver.AuthenticationInfo,
	servingInfo *genericapiserver.SecureServingInfo, openAPIConfig *openapicommon.Config,
) error {
	if s == nil {
		authenticationInfo.Authenticator = nil
		return nil
	}

	cfg := authenticatorfactory.DelegatingAuthenticatorConfig{
		Anonymous:                true,
		CacheTTL:                 s.CacheTTL,
		WebhookRetryBackoff:      s.WebhookRetryBackoff,
		TokenAccessReviewTimeout: s.TokenRequestTimeout,
	}

	var err error

	// get the clientCA information
	clientCASpecified := s.ClientCert != genericoptions.ClientCertAuthenticationOptions{}
	var clientCAProvider dynamiccertificates.CAContentProvider
	if clientCASpecified {
		clientCAProvider, err = s.ClientCert.GetClientCAContentProvider()
		if err != nil {
			return fmt.Errorf("unable to load client CA provider: %w", err)
		}
		cfg.ClientCertificateCAContentProvider = clientCAProvider
		if err = authenticationInfo.ApplyClientCert(cfg.ClientCertificateCAContentProvider, servingInfo); err != nil {
			return fmt.Errorf("unable to assign client CA provider: %w", err)
		}
	}

	requestHeaderCAFileSpecified := len(s.RequestHeader.ClientCAFile) > 0
	var requestHeaderConfig *authenticatorfactory.RequestHeaderConfig
	if requestHeaderCAFileSpecified {
		requestHeaderConfig, err = s.RequestHeader.ToAuthenticationRequestHeaderConfig()
		if err != nil {
			return fmt.Errorf("unable to create request header authentication config: %w", err)
		}
	}

	if requestHeaderConfig != nil {
		cfg.RequestHeaderConfig = requestHeaderConfig
		if err = authenticationInfo.ApplyClientCert(cfg.RequestHeaderConfig.CAContentProvider, servingInfo); err != nil {
			return fmt.Errorf("unable to load request-header-client-ca-file: %w", err)
		}
	}

	// create authenticator
	authenticator, securityDefinitions, err := cfg.New()
	if err != nil {
		return err
	}
	authenticationInfo.Authenticator = authenticator
	if openAPIConfig != nil {
		openAPIConfig.SecurityDefinitions = securityDefinitions
	}

	return nil
}
