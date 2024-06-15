// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"

	known "github.com/superproj/onex/internal/pkg/known/apiserver"
)

func addDefaultingFuncs(scheme *runtime.Scheme) error {
	return RegisterDefaults(scheme)
}

/*
// SetDefaults_MinerSet sets defaults for MinerSet
func SetDefaults_MinerSet(obj *MinerSet) {
	// Set MinerSetSpec.Replicas to 1 if it is not set.
	if obj.Spec.Replicas == nil {
		obj.Spec.Replicas = new(int32)
		*obj.Spec.Replicas = 1
	}

	// Set default template
	SetDefaults_MinerSpec(&obj.Spec.Template.Spec)

	// Set default DeletePolicy as Random.
	if obj.Spec.DeletePolicy == "" {
		obj.Spec.DeletePolicy = "Random"
	}
}
*/

// SetDefaults_Miner sets defaults for Miner.
func SetDefaults_Miner(obj *Miner) {
	// Miner name prefix is fixed to `mi-`
	if obj.ObjectMeta.GenerateName == "" {
		obj.ObjectMeta.GenerateName = "mi-"
	}

	SetDefaults_MinerSpec(&obj.Spec)
}

// SetDefaults_MinerSpec sets defaults for Miner spec.
func SetDefaults_MinerSpec(obj *MinerSpec) {
	if obj.MinerType == "" {
		obj.MinerType = known.DefaultNodeMinerType
	}
}

// SetDefaults_Chain sets defaults for Chain.
func SetDefaults_Chain(obj *Chain) {
	SetDefaults_ChainSpec(&obj.Spec)
}

// SetDefaults_ChainSpec sets defaults for Chain spec.
func SetDefaults_ChainSpec(obj *ChainSpec) {
	obj.BootstrapAccount = ptr.To("0x210d9eD12CEA87E33a98AA7Bcb4359eABA9e800e")
	if obj.MinerType == "" {
		obj.MinerType = known.DefaultGenesisMinerType
	}

	if obj.MinMineIntervalSeconds <= 0 {
		obj.MinMineIntervalSeconds = 12 * 60 * 60 // 12 hours
	}
}

// SetDefaults_Evaluate sets defaults for Evaluate.
func SetDefaults_Evaluate(obj *Evaluate) {
	SetDefaults_EvaluateSpec(&obj.Spec)
}

// SetDefaults_EvaluateSpec sets defaults for Evaluate spec.
func SetDefaults_EvaluateSpec(obj *EvaluateSpec) {
}

// SetDefaults_ModelCompare sets defaults for ModelCompare.
func SetDefaults_ModelCompare(obj *ModelCompare) {
	addModelCompareSelector(obj)
	SetDefaults_ModelCompareSpec(&obj.Spec)
}

// SetDefaults_ModelCompareSpec sets defaults for ModelCompare spec.
func SetDefaults_ModelCompareSpec(obj *ModelCompareSpec) {
}

func addModelCompareSelector(obj *ModelCompare) {
	obj.Spec.Selector = metav1.LabelSelector{
		MatchLabels: map[string]string{
			ModelCompareNameLabel: obj.Name,
		},
	}

	if obj.Spec.Template.ObjectMeta.Labels == nil {
		obj.Spec.Template.ObjectMeta.Labels = map[string]string{}
	}
	obj.Spec.Template.ObjectMeta.Labels[ModelCompareNameLabel] = obj.Name
}
