/*
Copyright 2015 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package minerset contains an admission controller that modifies and validation every new MinerSet.
package minerset

import (
	"context"
	"fmt"
	"io"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apiserver/pkg/admission"

	"github.com/superproj/onex/pkg/apis/apps"
	clientset "github.com/superproj/onex/pkg/generated/clientset/versioned"
	"github.com/superproj/onex/pkg/generated/informers"
	appslisters "github.com/superproj/onex/pkg/generated/listers/apps/v1beta1"
)

// PluginName indicates name of admission plugin.
const PluginName = "MinerSet"

// Register registers a plugin
func Register(plugins *admission.Plugins) {
	plugins.Register(PluginName, func(config io.Reader) (admission.Interface, error) {
		return NewPlugin(), nil
	})
}

// Plugin is an implementation of admission.Interface.
// It is validation and mutation plugin for MinerSet resource.
type Plugin struct {
	*admission.Handler
	lister appslisters.MinerSetLister
	client clientset.Interface
}

var _ admission.MutationInterface = &Plugin{}
var _ admission.ValidationInterface = &Plugin{}

// Admit makes an admission decision based on the request attributes
func (p *Plugin) Admit(ctx context.Context, attributes admission.Attributes, o admission.ObjectInterfaces) (err error) {
	// Ignore all calls to subresources or resources other than minersets.
	if shouldIgnore(attributes) {
		return nil
	}

	if !p.WaitForReady() {
		return admission.NewForbidden(attributes, fmt.Errorf("not yet ready to handle request"))
	}
	// Deletion operation does not require Admission.
	if attributes.GetOperation() == admission.Delete {
		return nil
	}

	ms, ok := attributes.GetObject().(*apps.MinerSet)
	if err != nil {
		return err
	}
	if !ok {
		return apierrors.NewBadRequest("Resource was marked with kind MinerSet but was unable to be converted")
	}
	ms.Spec.Template.Spec.DisplayName = fmt.Sprintf("miner-for-%s-minerset", ms.Name)
	// Ensure the label selector and template labels for the given MinerSet object.
	addMinerSetSelector(ms)

	return nil
}

// Validate do some validation on MinerSet.
func (p *Plugin) Validate(ctx context.Context, attributes admission.Attributes, o admission.ObjectInterfaces) (err error) {
	if shouldIgnore(attributes) {
		return nil
	}

	// Since we cannot obtain the specific resource when deleting,
	// we need to first Get and then check here.
	if attributes.GetOperation() == admission.Delete {
		ms, err := p.lister.MinerSets(attributes.GetNamespace()).Get(attributes.GetName())
		if err != nil {
			if apierrors.IsNotFound(err) {
				return nil
			}
			return apierrors.NewInternalError(fmt.Errorf("can not get minerset: %s", attributes.GetName()))
		}

		if v, ok := ms.GetAnnotations()[apps.AnnotationDeletionProtection]; ok && v == "true" {
			return admission.NewForbidden(attributes, fmt.Errorf("minerset has deletion protection turned on"))
		}

		return nil
	}

	_, ok := attributes.GetObject().(*apps.MinerSet)
	if !ok {
		return apierrors.NewBadRequest("Resource was marked with kind MinerSet but was unable to be converted")
	}

	// Here, we can add some validation logic.
	var allErrs []error
	if len(allErrs) > 0 {
		return utilerrors.NewAggregate(allErrs)
	}

	return nil
}

// addMinerSetSelector sets the label selector and template labels for the given MinerSet object.
// This function ensures that the MinerSet's Spec.Selector matches the template's labels,
// allowing the MinerSet to select the correct Miner.
func addMinerSetSelector(ms *apps.MinerSet) {
	// Set the Spec.Selector field to a LabelSelector that matches the MinerSet's name
	ms.Spec.Selector = metav1.LabelSelector{
		MatchLabels: map[string]string{
			apps.LabelMinerSet: ms.Name,
		},
	}

	// Ensure the template's labels include the MinerSet label
	// If the labels map is nil, initialize it
	if ms.Spec.Template.ObjectMeta.Labels == nil {
		ms.Spec.Template.ObjectMeta.Labels = map[string]string{}
	}
	ms.Spec.Template.ObjectMeta.Labels[apps.LabelMinerSet] = ms.Name
}

// SetInternalInformerFactory gets Lister from SharedInformerFactory.
// The lister knows how to lists MinerSets.
func (p *Plugin) SetInternalInformerFactory(f informers.SharedInformerFactory) {
	p.lister = f.Apps().V1beta1().MinerSets().Lister()
	p.SetReadyFunc(f.Apps().V1beta1().MinerSets().Informer().HasSynced)
}

// SetExternalClientSet implements the WantsExternalClientSet interface.
func (p *Plugin) SetExternalClientSet(client clientset.Interface) {
	p.client = client
}

// ValidateInitialization checks whether the plugin was correctly initialized.
func (p *Plugin) ValidateInitialization() error {
	if p.lister == nil {
		return fmt.Errorf("%s requires a machine lister", PluginName)
	}
	return nil
}

func shouldIgnore(attributes admission.Attributes) bool {
	// Ignore all calls to subresources or resources other than minersets.
	if len(attributes.GetSubresource()) != 0 || attributes.GetResource().GroupResource() != apps.Resource("minersets") {
		return true
	}

	return false
}

// NewPlugin creates a new always minerset admission control handler.
func NewPlugin() *Plugin {
	return &Plugin{
		Handler: admission.NewHandler(admission.Create, admission.Update, admission.Delete),
	}
}
