// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package invincible

import (
	"fmt"
	"sync"

	"github.com/superproj/onex/internal/pkg/client/store"
	"github.com/superproj/onex/internal/pkg/client/usercenter"
	genericoptions "github.com/superproj/onex/pkg/options"
)

// Interface is an interface that presents a aggregate clientset.
type Interface interface {
	Store() store.Interface
	UserCenter() usercenter.Interface
	// TBB() tbb.Interface
}

// impl is an implementation of Interface.
type impl struct {
	opts *InvincibleOptions
}

var (
	G    Interface
	once sync.Once
)

// Init initializes the client.
func Init(opts *InvincibleOptions) error {
	once.Do(func() {
		G = &impl{
			opts: opts,
		}
	})

	if G == nil {
		return fmt.Errorf("init cloud client failed")
	}

	return nil
}

// UserCenter return the usercenter client.
func (impl *impl) UserCenter() usercenter.Interface {
	return usercenter.NewUserCenter(impl.opts.UserCenterOptions, &genericoptions.EtcdOptions{})
}

// Store return the store client.
func (impl *impl) Store() store.Interface {
	return store.NewStore(impl.opts.GatewayStore, impl.opts.UserCenterStore)
}
