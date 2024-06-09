// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package rest

import (
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	api "k8s.io/kubernetes/pkg/apis/core"
	configmapstore "k8s.io/kubernetes/pkg/registry/core/configmap/storage"
	endpointsstore "k8s.io/kubernetes/pkg/registry/core/endpoint/storage"
	eventstore "k8s.io/kubernetes/pkg/registry/core/event/storage"
	namespacestore "k8s.io/kubernetes/pkg/registry/core/namespace/storage"
	secretstore "k8s.io/kubernetes/pkg/registry/core/secret/storage"
	servicestore "k8s.io/kubernetes/pkg/registry/core/service/storage"

	// configmapstore "github.com/superproj/onex/internal/registry/core/configmap/storage"
	// eventstore "github.com/superproj/onex/internal/registry/core/event/storage"
	// namespacestore "github.com/superproj/onex/internal/registry/core/namespace/storage".
	serializerutil "github.com/superproj/onex/internal/pkg/util/serializer"
)

// LegacyRESTStorageProvider provides information needed to build RESTStorage for kubernetes core, but
// does NOT implement the "normal" RESTStorageProvider (yet!)
type LegacyRESTStorageProvider struct {
	EventTTL time.Duration
}

// NewLegacyRESTStorage is a factory constructor to creates and returns the APIGroupInfo.
func (p LegacyRESTStorageProvider) NewLegacyRESTStorage(restOptionsGetter generic.RESTOptionsGetter) (genericapiserver.APIGroupInfo, error) {
	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(api.GroupName, legacyscheme.Scheme, legacyscheme.ParameterCodec, legacyscheme.Codecs)
	apiGroupInfo.NegotiatedSerializer = serializerutil.NewProtocolShieldSerializers(&legacyscheme.Codecs)

	namespaceStorage, namespaceStatusStorage, namespaceFinalizeStorage, err := namespacestore.NewREST(restOptionsGetter)
	if err != nil {
		return genericapiserver.APIGroupInfo{}, err
	}

	eventStorage, err := eventstore.NewREST(restOptionsGetter, uint64(p.EventTTL.Seconds()))
	if err != nil {
		return genericapiserver.APIGroupInfo{}, err
	}

	configMapStorage, err := configmapstore.NewREST(restOptionsGetter)
	if err != nil {
		return genericapiserver.APIGroupInfo{}, err
	}

	secretStorage, err := secretstore.NewREST(restOptionsGetter)
	if err != nil {
		return genericapiserver.APIGroupInfo{}, err
	}

	endpointsStorage, err := endpointsstore.NewREST(restOptionsGetter)
	if err != nil {
		return genericapiserver.APIGroupInfo{}, err
	}

	serviceRESTStorage, serviceStatusStorage, serviceRESTProxy, err := servicestore.NewREST(restOptionsGetter, "", nil, nil, endpointsStorage, nil, nil)
	if err != nil {
		return genericapiserver.APIGroupInfo{}, err
	}

	restStorageMap := map[string]rest.Storage{
		"namespaces":          namespaceStorage,
		"namespaces/status":   namespaceStatusStorage,
		"namespaces/finalize": namespaceFinalizeStorage,

		"events": eventStorage,

		"configmaps": configMapStorage,
		"secrets":    secretStorage,

		"endpoints":       endpointsStorage,
		"services":        serviceRESTStorage,
		"services/proxy":  serviceRESTProxy,
		"services/status": serviceStatusStorage,
	}

	apiGroupInfo.VersionedResourcesStorageMap[apiv1.SchemeGroupVersion.Version] = restStorageMap

	return apiGroupInfo, nil
}

// GroupName return the api group name.
func (p LegacyRESTStorageProvider) GroupName() string {
	return api.GroupName
}
