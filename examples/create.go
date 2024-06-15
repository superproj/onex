// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package main

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/superproj/onex/pkg/apis/apps/v1beta1"
	"github.com/superproj/onex/pkg/generated/clientset/versioned"
)

func main() {
	defaultKubeconfig := filepath.Join(homedir.HomeDir(), ".onex", "config")
	kubeconfig := pflag.StringP("kubeconfig", "c", defaultKubeconfig, "(optional) absolute path to the kubeconfig file")
	help := pflag.BoolP("help", "h", false, "Show this help message.")

	pflag.Parse()

	if *help {
		pflag.Usage()
		return
	}

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := versioned.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	compare := v1beta1.ModelCompare{
		ObjectMeta: metav1.ObjectMeta{
			Name: "modelcompare-from-clientgo",
		},
		Spec: v1beta1.ModelCompareSpec{
			Template: v1beta1.EvaluateTemplateSpec{
				Spec: v1beta1.EvaluateSpec{
					Provider: "text",
					SampleID: 2001,
				},
			},
			DisplayName: "test-for-modelcompare",
			ModelIDs:    []int64{1001, 1002, 1003},
		},
	}

	if _, err := clientset.AppsV1beta1().ModelCompares(metav1.NamespaceDefault).Create(context.Background(), &compare, metav1.CreateOptions{}); err != nil {
		panic(err)
	}

	fmt.Println("ModelCompare Created")
}
