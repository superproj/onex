// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package deny

import (
	"context"
	"errors"
	"io"

	"k8s.io/klog/v2"

	"k8s.io/apiserver/pkg/admission"
)

// PluginName indicates name of admission plugin.
const PluginName = "AlwaysDeny"

// Register registers a plugin.
func Register(plugins *admission.Plugins) {
	plugins.Register(PluginName, func(config io.Reader) (admission.Interface, error) {
		return NewAlwaysDeny(), nil
	})
}

// alwaysDeny is an implementation of admission.Interface which always says no to an admission request.
type alwaysDeny struct{}

var (
	_ admission.MutationInterface   = alwaysDeny{}
	_ admission.ValidationInterface = alwaysDeny{}
)

// Admit makes an admission decision based on the request attributes.
func (alwaysDeny) Admit(ctx context.Context, a admission.Attributes, o admission.ObjectInterfaces) (err error) {
	return admission.NewForbidden(a, errors.New("admission control is denying all modifications"))
}

// Validate makes an admission decision based on the request attributes.  It is NOT allowed to mutate.
func (alwaysDeny) Validate(ctx context.Context, a admission.Attributes, o admission.ObjectInterfaces) (err error) {
	return admission.NewForbidden(a, errors.New("admission control is denying all modifications"))
}

// Handles returns true if this admission controller can handle the given operation
// where operation can be one of CREATE, UPDATE, DELETE, or CONNECT.
func (alwaysDeny) Handles(operation admission.Operation) bool {
	return true
}

// NewAlwaysDeny creates an always deny admission handler.
func NewAlwaysDeny() admission.Interface {
	// DEPRECATED: AlwaysDeny denys all admission request, it is no use.
	klog.Warningf("%s admission controller is deprecated. "+
		"Please remove this controller from your configuration files and scripts", PluginName)
	return new(alwaysDeny)
}
