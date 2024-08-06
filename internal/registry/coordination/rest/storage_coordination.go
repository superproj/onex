// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package rest

import (
	coordinationv1 "k8s.io/api/coordination/v1"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"
	serverstorage "k8s.io/apiserver/pkg/server/storage"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	"k8s.io/kubernetes/pkg/apis/coordination"
	leasestore "k8s.io/kubernetes/pkg/registry/coordination/lease/storage"

	serializerutil "github.com/superproj/onex/internal/pkg/util/serializer"
	"github.com/superproj/onex/pkg/apiserver/storage"
	// leasestore "github.com/superproj/onex/internal/registry/coordination/lease/storage".
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
		coordination.GroupName,
		legacyscheme.Scheme,
		legacyscheme.ParameterCodec,
		legacyscheme.Codecs,
	)
	apiGroupInfo.NegotiatedSerializer = serializerutil.NewProtocolShieldSerializers(&legacyscheme.Codecs)

	storageMap, err := p.v1Storage(apiResourceConfigSource, restOptionsGetter)
	if err != nil {
		return genericapiserver.APIGroupInfo{}, err
	}
	apiGroupInfo.VersionedResourcesStorageMap[coordinationv1.SchemeGroupVersion.Version] = storageMap

	return apiGroupInfo, nil
}

func (p RESTStorageProvider) v1Storage(
	apiResourceConfigSource serverstorage.APIResourceConfigSource,
	restOptionsGetter generic.RESTOptionsGetter,
) (map[string]rest.Storage, error) {
	storage := map[string]rest.Storage{}

	// leases
	if resource := "leases"; apiResourceConfigSource.ResourceEnabled(coordinationv1.SchemeGroupVersion.WithResource(resource)) {
		leaseStorage, err := leasestore.NewREST(restOptionsGetter)
		if err != nil {
			return storage, err
		}
		storage[resource] = leaseStorage
	}

	return storage, nil
}

func (p RESTStorageProvider) GroupName() string {
	return coordination.GroupName
}
