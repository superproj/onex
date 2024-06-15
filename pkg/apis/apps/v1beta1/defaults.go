// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func addDefaultingFuncs(scheme *runtime.Scheme) error {
	return RegisterDefaults(scheme)
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
