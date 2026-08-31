// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/onexstack/onex.
//

package options

import (
	"fmt"
	"strings"

	apiextensionsapiserver "k8s.io/apiextensions-apiserver/pkg/apiserver"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apiserver/pkg/features"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	aggregatorscheme "k8s.io/kube-aggregator/pkg/apiserver/scheme"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
)

// authorizedModes is the set of authorizer plugins onex-apiserver supports.
// Keep it in sync with the switch in internal/controlplane/apiserver/authorizer.go.
var authorizedModes = sets.NewString("AlwaysAllow", "AlwaysDeny", "RBAC", "Webhook")

// Validate checks ServerRunOptions and return a slice of found errs.
func (o *Options) Validate() []error {
	errs := []error{}
	errs = append(errs, o.GenericServerRunOptions.Validate()...)
	errs = append(errs, o.RecommendedOptions.Validate()...)
	errs = append(errs, o.Features.Validate()...)
	errs = append(errs, o.Metrics.Validate()...)
	errs = append(errs, o.Traces.Validate()...)
	errs = append(errs, o.APIEnablement.Validate(legacyscheme.Scheme, apiextensionsapiserver.Scheme, aggregatorscheme.Scheme)...)
	// errs = append(errs, o.CloudOptions.Validate()...)
	errs = append(errs, validateAuthorizationModeFlags(o)...)
	errs = append(errs, validateUnknownVersionInteroperabilityProxyFlags(o)...)

	return errs
}

func validateAuthorizationModeFlags(options *Options) []error {
	errs := []error{}
	for _, mode := range strings.Split(options.AuthorizationMode, ",") {
		if mode == "" {
			continue
		}
		if !authorizedModes.Has(mode) {
			errs = append(errs, fmt.Errorf(
				"unsupported authorization mode %q; supported modes are AlwaysAllow, AlwaysDeny, RBAC and Webhook", mode))
		}
	}
	return errs
}

func validateUnknownVersionInteroperabilityProxyFlags(options *Options) []error {
	err := []error{}
	if !utilfeature.DefaultFeatureGate.Enabled(features.UnknownVersionInteroperabilityProxy) {
		if options.PeerCAFile != "" {
			err = append(err, fmt.Errorf("--peer-ca-file requires UnknownVersionInteroperabilityProxy feature to be turned on"))
		}
		if options.PeerAdvertiseAddress.PeerAdvertiseIP != "" {
			err = append(err, fmt.Errorf("--peer-advertise-ip requires UnknownVersionInteroperabilityProxy feature to be turned on"))
		}
		if options.PeerAdvertiseAddress.PeerAdvertisePort != "" {
			err = append(err, fmt.Errorf("--peer-advertise-port requires UnknownVersionInteroperabilityProxy feature to be turned on"))
		}
	}
	return err
}
