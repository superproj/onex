package app

import (
	"context"
	"fmt"
	"time"

	"k8s.io/client-go/metadata"
	"k8s.io/kubernetes/pkg/controller/garbagecollector"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/superproj/onex/cmd/onex-controller-manager/names"
	chaincontroller "github.com/superproj/onex/internal/controller/chain"
	namespacecontroller "github.com/superproj/onex/internal/controller/namespace"
	resourcecleancontroller "github.com/superproj/onex/internal/controller/resourceclean"
	synccontroller "github.com/superproj/onex/internal/controller/sync"
)

func newGarbageCollectorControllerDescriptor() *ControllerDescriptor {
	return &ControllerDescriptor{
		name:    names.GarbageCollectorController,
		aliases: []string{"garbagecollector"},
		addFunc: addGarbageCollectorController,
	}
}

func newNamespacedResourcesDeleterControllerDescriptor() *ControllerDescriptor {
	return &ControllerDescriptor{
		name:    names.NamespacedResourcesDeleterController,
		aliases: []string{"namespaced-resource-deleter"},
		addFunc: addNamespacedResourcesDeleterController,
	}
}

func newChainControllerDescriptor() *ControllerDescriptor {
	return &ControllerDescriptor{
		name:    names.ChainController,
		aliases: []string{"chain"},
		addFunc: addChainController,
	}
}

func newChainSyncControllerDescriptor() *ControllerDescriptor {
	return &ControllerDescriptor{
		name:    names.ChainController,
		aliases: []string{"chain-sync"},
		addFunc: addChainSyncController,
	}
}

func newMinerSetSyncControllerDescriptor() *ControllerDescriptor {
	return &ControllerDescriptor{
		name:    names.ChainController,
		aliases: []string{"minerset-sync"},
		addFunc: addMinerSetSyncController,
	}
}

func newMinerSyncControllerDescriptor() *ControllerDescriptor {
	return &ControllerDescriptor{
		name:    names.ChainController,
		aliases: []string{"minerset-sync"},
		addFunc: addMinerSyncController,
	}
}

func newResourceCleanControllerDescriptor() *ControllerDescriptor {
	return &ControllerDescriptor{
		name:    names.ResourceCleanController,
		aliases: []string{"resource-clean"},
		addFunc: addResourceCleanController,
	}
}

// add functions
func addNamespacedResourcesDeleterController(ctx context.Context, mgr ctrl.Manager, cctx ControllerContext) (bool, error) {
	return true, namespacecontroller.NewNamespacedResourcesDeleter(mgr, cctx.Config.Client, cctx.MetadataClient).
		SetupWithManager(mgr, cctx.ControllerManagerOptions)
}

func addChainController(ctx context.Context, mgr ctrl.Manager, cctx ControllerContext) (bool, error) {
	return true, (&chaincontroller.Reconciler{
		ComponentConfig:  &cctx.Config.ComponentConfig.ChainController,
		WatchFilterValue: cctx.Config.ComponentConfig.Generic.WatchFilterValue,
	}).SetupWithManager(ctx, mgr, cctx.ControllerManagerOptions)
}

func addChainSyncController(ctx context.Context, mgr ctrl.Manager, cctx ControllerContext) (bool, error) {
	return true, (&synccontroller.ChainSyncReconciler{
		Store: cctx.Store,
	}).SetupWithManager(ctx, mgr, cctx.ControllerManagerOptions)
}

func addMinerSetSyncController(ctx context.Context, mgr ctrl.Manager, cctx ControllerContext) (bool, error) {
	return true, (&synccontroller.MinerSetSyncReconciler{
		Store: cctx.Store,
	}).SetupWithManager(ctx, mgr, cctx.ControllerManagerOptions)
}

func addMinerSyncController(ctx context.Context, mgr ctrl.Manager, cctx ControllerContext) (bool, error) {
	return true, (&synccontroller.MinerSyncReconciler{
		Store: cctx.Store,
	}).SetupWithManager(ctx, mgr, cctx.ControllerManagerOptions)
}

func addResourceCleanController(ctx context.Context, mgr ctrl.Manager, cctx ControllerContext) (bool, error) {
	mgr.Add(resourcecleancontroller.NewCleanReconciler(
		mgr.GetClient(),
		cctx.Store,
		&resourcecleancontroller.Miner{},
		&resourcecleancontroller.MinerSet{},
		&resourcecleancontroller.Chain{},
	))

	return true, nil
}

// garbageCollector used to defines a garbage collector controller.
type garbageCollector struct {
	cctx ControllerContext
}

// Start implement manager.Runnable interface.
func (gc *garbageCollector) Start(ctx context.Context) error {
	if _, err := startGarbageCollectorController(ctx, gc.cctx); err != nil {
		return err
	}

	return nil
}

func addGarbageCollectorController(ctx context.Context, mgr ctrl.Manager, cctx ControllerContext) (bool, error) {
	return true, mgr.Add(&garbageCollector{cctx})
}

func startGarbageCollectorController(ctx context.Context, cctx ControllerContext) (bool, error) {
	if !cctx.Config.ComponentConfig.GarbageCollectorController.EnableGarbageCollector {
		return false, nil
	}

	gcClientset := cctx.ClientBuilder.ClientOrDie("generic-garbage-collector")
	discoveryClient := cctx.ClientBuilder.DiscoveryClientOrDie("generic-garbage-collector")

	config := cctx.ClientBuilder.ConfigOrDie("generic-garbage-collector")
	// Increase garbage collector controller's throughput: each object deletion takes two API calls,
	// so to get |config.QPS| deletion rate we need to allow 2x more requests for this controller.
	config.QPS *= 2
	metadataClient, err := metadata.NewForConfig(config)
	if err != nil {
		return true, err
	}

	garbageCollector, err := garbagecollector.NewComposedGarbageCollector(
		ctx,
		gcClientset,
		metadataClient,
		cctx.RESTMapper,
		cctx.GraphBuilder,
	)
	if err != nil {
		return true, fmt.Errorf("failed to start the generic garbage collector: %w", err)
	}

	// Start the garbage collector.
	workers := int(cctx.Config.ComponentConfig.GarbageCollectorController.ConcurrentGCSyncs)
	go garbageCollector.Run(ctx, workers)

	// Periodically refresh the RESTMapper with new discovery information and sync
	// the garbage collector.
	go garbageCollector.Sync(ctx, discoveryClient, 30*time.Second)

	return true, nil
}
