// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package admit

import (
	"context"
	"io"

	"k8s.io/apiserver/pkg/admission"
	"k8s.io/klog/v2"
)

// PluginName indicates name of admission plugin.
const PluginName = "AlwaysAdmit"

// Register registers a plugin.
func Register(plugins *admission.Plugins) {
	plugins.Register(PluginName, func(config io.Reader) (admission.Interface, error) {
		return NewAlwaysAdmit(), nil
	})
}

// alwaysAdmit is an implementation of admission.Interface which always says yes to an admit request.
type alwaysAdmit struct{}

var (
	_ admission.MutationInterface   = alwaysAdmit{}
	_ admission.ValidationInterface = alwaysAdmit{}
)

// Admit makes an admission decision based on the request attributes.
func (alwaysAdmit) Admit(ctx context.Context, a admission.Attributes, o admission.ObjectInterfaces) (err error) {
	return nil
}

// Validate makes an admission decision based on the request attributes.  It is NOT allowed to mutate.
func (alwaysAdmit) Validate(ctx context.Context, a admission.Attributes, o admission.ObjectInterfaces) (err error) {
	return nil
}

// Handles returns true if this admission controller can handle the given operation
// where operation can be one of CREATE, UPDATE, DELETE, or CONNECT.
func (alwaysAdmit) Handles(operation admission.Operation) bool {
	return true
}

// NewAlwaysAdmit creates a new always admit admission handler.
func NewAlwaysAdmit() admission.Interface {
	// DEPRECATED: AlwaysAdmit admit all admission request, it is no use.
	klog.Warningf("%s admission controller is deprecated. "+
		"Please remove this controller from your configuration files and scripts", PluginName)
	return new(alwaysAdmit)
}
