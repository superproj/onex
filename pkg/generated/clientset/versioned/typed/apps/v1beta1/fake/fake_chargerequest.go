// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1beta1 "github.com/superproj/onex/pkg/apis/apps/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeChargeRequests implements ChargeRequestInterface
type FakeChargeRequests struct {
	Fake *FakeAppsV1beta1
	ns   string
}

var chargerequestsResource = schema.GroupVersionResource{Group: "apps.onex.io", Version: "v1beta1", Resource: "chargerequests"}

var chargerequestsKind = schema.GroupVersionKind{Group: "apps.onex.io", Version: "v1beta1", Kind: "ChargeRequest"}

// Get takes name of the chargeRequest, and returns the corresponding chargeRequest object, and an error if there is any.
func (c *FakeChargeRequests) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1beta1.ChargeRequest, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(chargerequestsResource, c.ns, name), &v1beta1.ChargeRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.ChargeRequest), err
}

// List takes label and field selectors, and returns the list of ChargeRequests that match those selectors.
func (c *FakeChargeRequests) List(ctx context.Context, opts v1.ListOptions) (result *v1beta1.ChargeRequestList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(chargerequestsResource, chargerequestsKind, c.ns, opts), &v1beta1.ChargeRequestList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1beta1.ChargeRequestList{ListMeta: obj.(*v1beta1.ChargeRequestList).ListMeta}
	for _, item := range obj.(*v1beta1.ChargeRequestList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested chargeRequests.
func (c *FakeChargeRequests) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(chargerequestsResource, c.ns, opts))

}

// Create takes the representation of a chargeRequest and creates it.  Returns the server's representation of the chargeRequest, and an error, if there is any.
func (c *FakeChargeRequests) Create(ctx context.Context, chargeRequest *v1beta1.ChargeRequest, opts v1.CreateOptions) (result *v1beta1.ChargeRequest, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(chargerequestsResource, c.ns, chargeRequest), &v1beta1.ChargeRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.ChargeRequest), err
}

// Update takes the representation of a chargeRequest and updates it. Returns the server's representation of the chargeRequest, and an error, if there is any.
func (c *FakeChargeRequests) Update(ctx context.Context, chargeRequest *v1beta1.ChargeRequest, opts v1.UpdateOptions) (result *v1beta1.ChargeRequest, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(chargerequestsResource, c.ns, chargeRequest), &v1beta1.ChargeRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.ChargeRequest), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeChargeRequests) UpdateStatus(ctx context.Context, chargeRequest *v1beta1.ChargeRequest, opts v1.UpdateOptions) (*v1beta1.ChargeRequest, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(chargerequestsResource, "status", c.ns, chargeRequest), &v1beta1.ChargeRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.ChargeRequest), err
}

// Delete takes name of the chargeRequest and deletes it. Returns an error if one occurs.
func (c *FakeChargeRequests) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(chargerequestsResource, c.ns, name, opts), &v1beta1.ChargeRequest{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeChargeRequests) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(chargerequestsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1beta1.ChargeRequestList{})
	return err
}

// Patch applies the patch and returns the patched chargeRequest.
func (c *FakeChargeRequests) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1beta1.ChargeRequest, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(chargerequestsResource, c.ns, name, pt, data, subresources...), &v1beta1.ChargeRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.ChargeRequest), err
}
