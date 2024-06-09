// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package rest

import (
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"
	serverstorage "k8s.io/apiserver/pkg/server/storage"
	"k8s.io/kubernetes/pkg/api/legacyscheme"

	"github.com/superproj/onex/internal/controlplane/storage"
	serializerutil "github.com/superproj/onex/internal/pkg/util/serializer"
	chainstore "github.com/superproj/onex/internal/registry/apps/chain/storage"
	minerstore "github.com/superproj/onex/internal/registry/apps/miner/storage"
	minersetstore "github.com/superproj/onex/internal/registry/apps/minerset/storage"
	"github.com/superproj/onex/pkg/apis/apps"
	"github.com/superproj/onex/pkg/apis/apps/v1beta1"
)

// RESTStorageProvider is a struct for apps REST storage.
type RESTStorageProvider struct{}

// Implement RESTStorageProvider.
var _ storage.RESTStorageProvider = &RESTStorageProvider{}

// NewRESTStorage returns APIGroupInfo object.
func (p RESTStorageProvider) NewRESTStorage(
	apiResourceConfigSource serverstorage.APIResourceConfigSource,
	restOptionsGetter generic.RESTOptionsGetter,
) (genericapiserver.APIGroupInfo, error) {
	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(apps.GroupName, legacyscheme.Scheme, legacyscheme.ParameterCodec, legacyscheme.Codecs)
	// If you add a version here, be sure to add an entry in `k8s.io/kubernetes/cmd/kube-apiserver/app/aggregator.go with specific priorities.
	// TODO refactor the plumbing to provide the information in the APIGroupInfo

	apiGroupInfo.NegotiatedSerializer = serializerutil.NewProtocolShieldSerializers(&legacyscheme.Codecs)

	storageMap, err := p.v1beta1Storage(apiResourceConfigSource, restOptionsGetter)
	if err != nil {
		return genericapiserver.APIGroupInfo{}, err
	}
	apiGroupInfo.VersionedResourcesStorageMap[v1beta1.SchemeGroupVersion.Version] = storageMap

	return apiGroupInfo, nil
}

func (p RESTStorageProvider) v1beta1Storage(
	apiResourceConfigSource serverstorage.APIResourceConfigSource,
	restOptionsGetter generic.RESTOptionsGetter,
) (map[string]rest.Storage, error) {
	storage := map[string]rest.Storage{}

	//nolint:goconst
	// chains
	if resource := "chains"; apiResourceConfigSource.ResourceEnabled(v1beta1.SchemeGroupVersion.WithResource(resource)) {
		chainStorage, err := chainstore.NewStorage(restOptionsGetter)
		if err != nil {
			return storage, err
		}

		storage[resource] = chainStorage.Chain
		storage[resource+"/status"] = chainStorage.Status
	}

	// miners
	if resource := "miners"; apiResourceConfigSource.ResourceEnabled(v1beta1.SchemeGroupVersion.WithResource(resource)) {
		minerStorage, err := minerstore.NewStorage(restOptionsGetter)
		if err != nil {
			return storage, err
		}

		storage[resource] = minerStorage.Miner
		storage[resource+"/status"] = minerStorage.Status
	}

	// minersets
	if resource := "minersets"; apiResourceConfigSource.ResourceEnabled(v1beta1.SchemeGroupVersion.WithResource(resource)) {
		minerSetStorage, err := minersetstore.NewStorage(restOptionsGetter)
		if err != nil {
			return storage, err
		}

		storage[resource] = minerSetStorage.MinerSet
		storage[resource+"/status"] = minerSetStorage.Status
		storage[resource+"/scale"] = minerSetStorage.Scale
	}

	return storage, nil
}

// GroupName return the api group name.
func (p RESTStorageProvider) GroupName() string {
	return apps.GroupName
}
