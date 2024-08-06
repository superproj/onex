// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package options

import (
	apiserveroptions "k8s.io/apiserver/pkg/server/options"

	controlplaneoptions "github.com/superproj/onex/internal/controlplane/apiserver/options"
	"github.com/superproj/onex/pkg/apiserver/storage"
)

// completedOptions is a private wrapper that enforces a call of Complete() before Run can be invoked.
type completedOptions struct {
	controlplaneoptions.CompletedOptions

	Extra
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

	controlplane, err := o.Options.Complete()
	if err != nil {
		return CompletedOptions{}, err
	}

	completed := completedOptions{
		CompletedOptions: controlplane,
		Extra:            o.Extra,
	}

	if completed.RecommendedOptions.Etcd != nil && completed.RecommendedOptions.Etcd.EnableWatchCache {
		sizes := storage.DefaultWatchCacheSizes()
		// Ensure that overrides parse correctly.
		userSpecified, err := apiserveroptions.ParseWatchCacheSizes(completed.RecommendedOptions.Etcd.WatchCacheSizes)
		if err != nil {
			return CompletedOptions{}, err
		}
		for resource, size := range userSpecified {
			sizes[resource] = size
		}
		completed.RecommendedOptions.Etcd.WatchCacheSizes, err = apiserveroptions.WriteWatchCacheSizes(sizes)
		if err != nil {
			return CompletedOptions{}, err
		}
	}

	return CompletedOptions{&completed}, nil
}
