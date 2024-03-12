// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package contract

import "sync"

// BootstrapContract encodes information about the Cluster API contract for bootstrap objects.
type BootstrapContract struct{}

var (
	bootstrap     *BootstrapContract
	onceBootstrap sync.Once
)

// Bootstrap provide access to the information about the Cluster API contract for bootstrap objects.
func Bootstrap() *BootstrapContract {
	onceBootstrap.Do(func() {
		bootstrap = &BootstrapContract{}
	})
	return bootstrap
}

// Ready provide access to status.ready field in a bootstrap object.
func (b *BootstrapContract) Ready() *Bool {
	return &Bool{
		path: []string{"status", "ready"},
	}
}

// DataSecretName provide access to status.dataSecretName field in a bootstrap object.
func (b *BootstrapContract) DataSecretName() *String {
	return &String{
		path: []string{"status", "dataSecretName"},
	}
}

// FailureReason provides access to the status.failureReason field in an bootstrap object. Note that this field is optional.
func (b *BootstrapContract) FailureReason() *String {
	return &String{
		path: []string{"status", "failureReason"},
	}
}

// FailureMessage provides access to the status.failureMessage field in an bootstrap object. Note that this field is optional.
func (b *BootstrapContract) FailureMessage() *String {
	return &String{
		path: []string{"status", "failureMessage"},
	}
}
