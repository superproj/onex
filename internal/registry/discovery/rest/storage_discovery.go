// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/onex.
//

package rest

import (
	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"
	serverstorage "k8s.io/apiserver/pkg/server/storage"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	"k8s.io/kubernetes/pkg/apis/discovery"
	endpointslicestore "k8s.io/kubernetes/pkg/registry/discovery/endpointslice/storage"

	serializerutil "github.com/onexstack/onex/internal/pkg/util/serializer"
	"github.com/onexstack/onex/pkg/apiserver/storage"
)

type RESTStorageProvider struct{}

// Implement RESTStorageProvider.
var _ storage.RESTStorageProvider = &RESTStorageProvider{}

// NewRESTStorage is a factory constructor to creates and returns the APIGroupInfo.
func (p RESTStorageProvider) NewRESTStorage(
	apiResourceConfigSource serverstorage.APIResourceConfigSource,
	restOptionsGetter generic.RESTOptionsGetter,
) (genericapiserver.APIGroupInfo, error) {
	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(
		discovery.GroupName,
		legacyscheme.Scheme,
		legacyscheme.ParameterCodec,
		legacyscheme.Codecs,
	)
	apiGroupInfo.NegotiatedSerializer = serializerutil.NewProtocolShieldSerializers(&legacyscheme.Codecs)

	storageMap, err := p.v1Storage(apiResourceConfigSource, restOptionsGetter)
	if err != nil {
		return genericapiserver.APIGroupInfo{}, err
	}
	apiGroupInfo.VersionedResourcesStorageMap[discoveryv1.SchemeGroupVersion.Version] = storageMap

	return apiGroupInfo, nil
}

func (p RESTStorageProvider) v1Storage(
	apiResourceConfigSource serverstorage.APIResourceConfigSource,
	restOptionsGetter generic.RESTOptionsGetter,
) (map[string]rest.Storage, error) {
	storage := map[string]rest.Storage{}

	// endpointslices
	if resource := "endpointslices"; apiResourceConfigSource.ResourceEnabled(discoveryv1.SchemeGroupVersion.WithResource(resource)) {
		endpointSliceStorage, err := endpointslicestore.NewREST(restOptionsGetter)
		if err != nil {
			return storage, err
		}
		storage[resource] = endpointSliceStorage
	}

	return storage, nil
}

func (p RESTStorageProvider) GroupName() string {
	return discovery.GroupName
}