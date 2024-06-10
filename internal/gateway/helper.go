// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package gateway

import (
	"fmt"
	"time"

	"k8s.io/client-go/tools/cache"

	clientset "github.com/superproj/onex/pkg/generated/clientset/versioned"
	"github.com/superproj/onex/pkg/generated/informers"
	"github.com/superproj/onex/pkg/log"
)

func createInformers(stopCh <-chan struct{}, client clientset.Interface) (informers.SharedInformerFactory, error) {
	f := informers.NewSharedInformerFactory(client, time.Minute)
	msinfor := f.Apps().V1beta1().MinerSets().Informer()
	minfor := f.Apps().V1beta1().Miners().Informer()

	f.Start(stopCh)
	if !cache.WaitForCacheSync(stopCh, msinfor.HasSynced, minfor.HasSynced) {
		log.Errorf("Failed to wait for caches to populate")
		return nil, fmt.Errorf("failed to wait caches to populate")
	}

	return f, nil
}
