// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package initializer

import (
	"k8s.io/apiserver/pkg/admission"

	clientset "github.com/superproj/onex/pkg/generated/clientset/versioned"
	"github.com/superproj/onex/pkg/generated/informers"
)

type pluginInitializer struct {
	informers      informers.SharedInformerFactory
	externalClient clientset.Interface
}

var _ admission.PluginInitializer = pluginInitializer{}

// New creates an instance of node admission plugins initializer.
func New(
	informers informers.SharedInformerFactory,
	extClientset clientset.Interface,
) pluginInitializer {
	return pluginInitializer{
		informers:      informers,
		externalClient: extClientset,
	}
}

// Initialize checks the initialization interfaces implemented by a plugin
// and provide the appropriate initialization data.
func (i pluginInitializer) Initialize(plugin admission.Interface) {
	if wants, ok := plugin.(WantsInternalMinerInformerFactory); ok {
		wants.SetInternalMinerInformerFactory(i.informers)
	}

	if wants, ok := plugin.(WantsExternalMinerClientSet); ok {
		wants.SetExternalMinerClientSet(i.externalClient)
	}
}
