// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package rest

import (
	"context"
	"fmt"
	"time"

	flowcontrolv1 "k8s.io/api/flowcontrol/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	flowcontrolbootstrap "k8s.io/apiserver/pkg/apis/flowcontrol/bootstrap"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"
	serverstorage "k8s.io/apiserver/pkg/server/storage"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	"k8s.io/kubernetes/pkg/apis/flowcontrol"
	flowschemastore "k8s.io/kubernetes/pkg/registry/flowcontrol/flowschema/storage"
	prioritylevelconfigurationstore "k8s.io/kubernetes/pkg/registry/flowcontrol/prioritylevelconfiguration/storage"

	serializerutil "github.com/superproj/onex/internal/pkg/util/serializer"
	"github.com/superproj/onex/internal/registry/flowcontrol/ensurer"
	flowcontrolclient "github.com/superproj/onex/pkg/generated/clientset/versioned/typed/flowcontrol/v1"
	"github.com/superproj/onex/pkg/generated/informers"
	flowcontrollisters "github.com/superproj/onex/pkg/generated/listers/flowcontrol/v1"
)

var _ genericapiserver.PostStartHookProvider = RESTStorageProvider{}

// RESTStorageProvider is a provider of REST storage
type RESTStorageProvider struct {
	InformerFactory informers.SharedInformerFactory
}

// PostStartHookName is the name of the post-start-hook provided by flow-control storage
const PostStartHookName = "priority-and-fairness-config-producer"

// NewRESTStorage creates a new rest storage for flow-control api models.
func (p RESTStorageProvider) NewRESTStorage(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter generic.RESTOptionsGetter) (genericapiserver.APIGroupInfo, error) {
	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(flowcontrol.GroupName, legacyscheme.Scheme, legacyscheme.ParameterCodec, legacyscheme.Codecs)
	apiGroupInfo.NegotiatedSerializer = serializerutil.NewProtocolShieldSerializers(&legacyscheme.Codecs)

	flowSchemaStorage, flowSchemaStatusStorage, err := flowschemastore.NewREST(restOptionsGetter)
	if err != nil {
		return genericapiserver.APIGroupInfo{}, err
	}

	priorityLevelConfigurationStorage, priorityLevelConfigurationStatusStorage, err := prioritylevelconfigurationstore.NewREST(restOptionsGetter)
	if err != nil {
		return genericapiserver.APIGroupInfo{}, err
	}

	restStorageMap := map[string]rest.Storage{
		"flowschemas":        flowSchemaStorage,
		"flowschemas/status": flowSchemaStatusStorage,

		"prioritylevelconfigurations":        priorityLevelConfigurationStorage,
		"prioritylevelconfigurations/status": priorityLevelConfigurationStatusStorage,
	}

	apiGroupInfo.VersionedResourcesStorageMap[flowcontrolv1.SchemeGroupVersion.Version] = restStorageMap

	return apiGroupInfo, nil
}

// GroupName return the api group name.
func (p RESTStorageProvider) GroupName() string {
	return flowcontrol.GroupName
}

// PostStartHook returns the hook func that launches the config provider
func (p RESTStorageProvider) PostStartHook() (string, genericapiserver.PostStartHookFunc, error) {
	bce := &bootstrapConfigurationEnsurer{
		informersSynced: []cache.InformerSynced{
			p.InformerFactory.Flowcontrol().V1().PriorityLevelConfigurations().Informer().HasSynced,
			p.InformerFactory.Flowcontrol().V1().FlowSchemas().Informer().HasSynced,
		},
		fsLister:  p.InformerFactory.Flowcontrol().V1().FlowSchemas().Lister(),
		plcLister: p.InformerFactory.Flowcontrol().V1().PriorityLevelConfigurations().Lister(),
	}
	return PostStartHookName, bce.ensureAPFBootstrapConfiguration, nil
}

type bootstrapConfigurationEnsurer struct {
	informersSynced []cache.InformerSynced
	fsLister        flowcontrollisters.FlowSchemaLister
	plcLister       flowcontrollisters.PriorityLevelConfigurationLister
}

func (bce *bootstrapConfigurationEnsurer) ensureAPFBootstrapConfiguration(hookContext genericapiserver.PostStartHookContext) error {
	clientset, err := flowcontrolclient.NewForConfig(hookContext.LoopbackClientConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize clientset for APF - %w", err)
	}

	err = func() error {
		// get a derived context that gets cancelled after 5m or
		// when the StopCh gets closed, whichever happens first.
		ctx, cancel := contextFromChannelAndMaxWaitDuration(hookContext.StopCh, 5*time.Minute)
		defer cancel()

		if !cache.WaitForCacheSync(ctx.Done(), bce.informersSynced...) {
			return fmt.Errorf("APF bootstrap ensurer timed out waiting for cache sync")
		}

		err = wait.PollImmediateUntilWithContext(
			ctx,
			time.Second,
			func(context.Context) (bool, error) {
				if err := ensure(ctx, clientset, bce.fsLister, bce.plcLister); err != nil {
					klog.ErrorS(err, "APF bootstrap ensurer ran into error, will retry later")
					return false, nil
				}
				return true, nil
			})
		if err != nil {
			return fmt.Errorf("unable to initialize APF bootstrap configuration: %w", err)
		}
		return nil
	}()
	if err != nil {
		return err
	}

	// we have successfully initialized the bootstrap configuration, now we
	// spin up a goroutine which reconciles the bootstrap configuration periodically.
	go func() {
		ctx := wait.ContextForChannel(hookContext.StopCh)
		wait.PollImmediateUntil(
			time.Minute,
			func() (bool, error) {
				if err := ensure(ctx, clientset, bce.fsLister, bce.plcLister); err != nil {
					klog.ErrorS(err, "APF bootstrap ensurer ran into error, will retry later")
				}
				// always auto update both suggested and mandatory configuration
				return false, nil
			}, hookContext.StopCh)
		klog.Info("APF bootstrap ensurer is exiting")
	}()

	return nil
}

func ensure(ctx context.Context, clientset flowcontrolclient.FlowcontrolV1Interface, fsLister flowcontrollisters.FlowSchemaLister, plcLister flowcontrollisters.PriorityLevelConfigurationLister) error {

	if err := ensureSuggestedConfiguration(ctx, clientset, fsLister, plcLister); err != nil {
		// We should not attempt creation of mandatory objects if ensuring the suggested
		// configuration resulted in an error.
		// This only happens when the stop channel is closed.
		return fmt.Errorf("failed ensuring suggested settings - %w", err)
	}

	if err := ensureMandatoryConfiguration(ctx, clientset, fsLister, plcLister); err != nil {
		return fmt.Errorf("failed ensuring mandatory settings - %w", err)
	}

	if err := removeDanglingBootstrapConfiguration(ctx, clientset, fsLister, plcLister); err != nil {
		return fmt.Errorf("failed to delete removed settings - %w", err)
	}

	return nil
}

func ensureSuggestedConfiguration(ctx context.Context, clientset flowcontrolclient.FlowcontrolV1Interface, fsLister flowcontrollisters.FlowSchemaLister, plcLister flowcontrollisters.PriorityLevelConfigurationLister) error {
	plcOps := ensurer.NewPriorityLevelConfigurationOps(clientset.PriorityLevelConfigurations(), plcLister)
	if err := ensurer.EnsureConfigurations(ctx, plcOps, flowcontrolbootstrap.SuggestedPriorityLevelConfigurations, ensurer.NewSuggestedEnsureStrategy[*flowcontrolv1.PriorityLevelConfiguration]()); err != nil {
		return err
	}

	fsOps := ensurer.NewFlowSchemaOps(clientset.FlowSchemas(), fsLister)
	return ensurer.EnsureConfigurations(ctx, fsOps, flowcontrolbootstrap.SuggestedFlowSchemas, ensurer.NewSuggestedEnsureStrategy[*flowcontrolv1.FlowSchema]())
}

func ensureMandatoryConfiguration(ctx context.Context, clientset flowcontrolclient.FlowcontrolV1Interface, fsLister flowcontrollisters.FlowSchemaLister, plcLister flowcontrollisters.PriorityLevelConfigurationLister) error {
	plcOps := ensurer.NewPriorityLevelConfigurationOps(clientset.PriorityLevelConfigurations(), plcLister)
	if err := ensurer.EnsureConfigurations(ctx, plcOps, flowcontrolbootstrap.MandatoryPriorityLevelConfigurations, ensurer.NewMandatoryEnsureStrategy[*flowcontrolv1.PriorityLevelConfiguration]()); err != nil {
		return err
	}

	fsOps := ensurer.NewFlowSchemaOps(clientset.FlowSchemas(), fsLister)
	return ensurer.EnsureConfigurations(ctx, fsOps, flowcontrolbootstrap.MandatoryFlowSchemas, ensurer.NewMandatoryEnsureStrategy[*flowcontrolv1.FlowSchema]())
}

func removeDanglingBootstrapConfiguration(ctx context.Context, clientset flowcontrolclient.FlowcontrolV1Interface, fsLister flowcontrollisters.FlowSchemaLister, plcLister flowcontrollisters.PriorityLevelConfigurationLister) error {
	if err := removeDanglingBootstrapFlowSchema(ctx, clientset, fsLister); err != nil {
		return err
	}

	return removeDanglingBootstrapPriorityLevel(ctx, clientset, plcLister)
}

func removeDanglingBootstrapFlowSchema(ctx context.Context, clientset flowcontrolclient.FlowcontrolV1Interface, fsLister flowcontrollisters.FlowSchemaLister) error {
	bootstrap := append(flowcontrolbootstrap.MandatoryFlowSchemas, flowcontrolbootstrap.SuggestedFlowSchemas...)
	fsOps := ensurer.NewFlowSchemaOps(clientset.FlowSchemas(), fsLister)
	return ensurer.RemoveUnwantedObjects(ctx, fsOps, bootstrap)
}

func removeDanglingBootstrapPriorityLevel(ctx context.Context, clientset flowcontrolclient.FlowcontrolV1Interface, plcLister flowcontrollisters.PriorityLevelConfigurationLister) error {
	bootstrap := append(flowcontrolbootstrap.MandatoryPriorityLevelConfigurations, flowcontrolbootstrap.SuggestedPriorityLevelConfigurations...)
	plcOps := ensurer.NewPriorityLevelConfigurationOps(clientset.PriorityLevelConfigurations(), plcLister)
	return ensurer.RemoveUnwantedObjects(ctx, plcOps, bootstrap)
}

// contextFromChannelAndMaxWaitDuration returns a Context that is bound to the
// specified channel and the wait duration. The derived context will be
// cancelled when the specified channel stopCh is closed or the maximum wait
// duration specified in maxWait elapses, whichever happens first.
//
// Note the caller must *always* call the CancelFunc, otherwise resources may be leaked.
func contextFromChannelAndMaxWaitDuration(stopCh <-chan struct{}, maxWait time.Duration) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		defer cancel()

		select {
		case <-stopCh:
		case <-time.After(maxWait):

		// the caller can explicitly cancel the context which is an
		// indication to us to exit the goroutine immediately.
		// Note that we are calling cancel more than once when we are here,
		// CancelFunc is idempotent and we expect no ripple effects here.
		case <-ctx.Done():
		}
	}()
	return ctx, cancel
}
