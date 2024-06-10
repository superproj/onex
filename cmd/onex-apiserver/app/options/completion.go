// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package options

import (
	"fmt"
	"net"
	"os"

	"k8s.io/apiserver/pkg/admission"
	genericapiserver "k8s.io/apiserver/pkg/server"
	serveroptions "k8s.io/apiserver/pkg/server/options"
	"k8s.io/klog/v2"
	netutils "k8s.io/utils/net"

	"github.com/superproj/onex/internal/admission/initializer"
	"github.com/superproj/onex/internal/apiserver/storage"
	"github.com/superproj/onex/pkg/generated/clientset/versioned"
	"github.com/superproj/onex/pkg/generated/informers"
)

// completedOptions is a private wrapper that enforces a call of Complete() before Run can be invoked.
type completedOptions struct {
	*ServerRunOptions
}

type CompletedOptions struct {
	// Embed a private pointer that cannot be instantiated outside of this package.
	*completedOptions
}

// Complete set default ServerRunOptions.
// Should be called after onex-apiserver flags parsed.
func (o *ServerRunOptions) Complete() (CompletedOptions, error) {
	if o == nil {
		return CompletedOptions{completedOptions: &completedOptions{}}, nil
	}

	// set defaults
	if err := o.GenericServerRunOptions.DefaultAdvertiseAddress(o.RecommendedOptions.SecureServing.SecureServingOptions); err != nil {
		return CompletedOptions{}, err
	}

	// TODO have a "real" external address
	if err := o.RecommendedOptions.SecureServing.MaybeDefaultWithSelfSignedCerts(
		o.GenericServerRunOptions.AdvertiseAddress.String(),
		[]string{"onex.io"},
		[]net.IP{netutils.ParseIPSloppy("127.0.0.1")}); err != nil {
		return CompletedOptions{}, fmt.Errorf("error creating self-signed certificates: %w", err)
	}

	//nolint: nestif
	if len(o.GenericServerRunOptions.ExternalHost) == 0 {
		if len(o.GenericServerRunOptions.AdvertiseAddress) > 0 {
			o.GenericServerRunOptions.ExternalHost = o.GenericServerRunOptions.AdvertiseAddress.String()
		} else {
			if hostname, err := os.Hostname(); err == nil {
				o.GenericServerRunOptions.ExternalHost = hostname
			} else {
				return CompletedOptions{}, fmt.Errorf("error finding host name: %w", err)
			}
		}
		klog.Infof("external host was not specified, using %v", o.GenericServerRunOptions.ExternalHost)
	}

	if o.RecommendedOptions.Etcd != nil && o.RecommendedOptions.Etcd.EnableWatchCache {
		sizes := storage.DefaultWatchCacheSizes()
		// Ensure that overrides parse correctly.
		userSpecified, err := serveroptions.ParseWatchCacheSizes(o.RecommendedOptions.Etcd.WatchCacheSizes)
		if err != nil {
			return CompletedOptions{}, err
		}
		for resource, size := range userSpecified {
			sizes[resource] = size
		}
		o.RecommendedOptions.Etcd.WatchCacheSizes, err = serveroptions.WriteWatchCacheSizes(sizes)
		if err != nil {
			return CompletedOptions{}, err
		}
	}

	o.RecommendedOptions.ExtraAdmissionInitializers = func(c *genericapiserver.RecommendedConfig) ([]admission.PluginInitializer, error) {
		client, err := versioned.NewForConfig(c.LoopbackClientConfig)
		if err != nil {
			return nil, err
		}
		informerFactory := informers.NewSharedInformerFactory(client, c.LoopbackClientConfig.Timeout)
		o.SharedInformerFactory = informerFactory
		return []admission.PluginInitializer{initializer.New(informerFactory, client)}, nil
	}

	completed := completedOptions{
		ServerRunOptions: o,
	}
	return CompletedOptions{&completed}, nil
}
