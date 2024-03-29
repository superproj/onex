//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.

// Code generated by deepcopy-gen. DO NOT EDIT.

package v1beta1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MinerControllerConfiguration) DeepCopyInto(out *MinerControllerConfiguration) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	if in.FeatureGates != nil {
		in, out := &in.FeatureGates, &out.FeatureGates
		*out = make(map[string]bool, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	out.SyncPeriod = in.SyncPeriod
	in.LeaderElection.DeepCopyInto(&out.LeaderElection)
	if in.Types != nil {
		in, out := &in.Types, &out.Types
		*out = make(map[string]MinerProfile, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
	out.Redis = in.Redis
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MinerControllerConfiguration.
func (in *MinerControllerConfiguration) DeepCopy() *MinerControllerConfiguration {
	if in == nil {
		return nil
	}
	out := new(MinerControllerConfiguration)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MinerControllerConfiguration) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MinerProfile) DeepCopyInto(out *MinerProfile) {
	*out = *in
	out.CPU = in.CPU.DeepCopy()
	out.Memory = in.Memory.DeepCopy()
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MinerProfile.
func (in *MinerProfile) DeepCopy() *MinerProfile {
	if in == nil {
		return nil
	}
	out := new(MinerProfile)
	in.DeepCopyInto(out)
	return out
}
