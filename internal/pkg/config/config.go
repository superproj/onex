// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package config

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clientset "github.com/superproj/onex/pkg/generated/clientset/versioned"
)

type ConfigurationName string

// Configuration name.
const (
	OneXName        ConfigurationName = "onex"
	MinerTypesName  ConfigurationName = "minertypes"
	IDGeneraterName ConfigurationName = "idgenerater"
)

func (cn ConfigurationName) String() string {
	return string(cn)
}

func (cn ConfigurationName) NamespacedName() types.NamespacedName {
	return types.NamespacedName{Name: cn.String(), Namespace: metav1.NamespaceSystem}
}

func (cn ConfigurationName) GetConfig(cli any) (*corev1.ConfigMap, error) {
	var err error
	cm := new(corev1.ConfigMap)

	switch v := cli.(type) {
	case clientset.Interface:
		cm, err = v.CoreV1().ConfigMaps(metav1.NamespaceSystem).Get(
			context.Background(),
			cn.String(),
			metav1.GetOptions{},
		)
	case client.Client:
		err = v.Get(context.Background(), cn.NamespacedName(), cm)
	default:
		err = fmt.Errorf("unsupported kubernetes client")
	}

	return cm, err
}
