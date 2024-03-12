// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package main

import (
	"context"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/spf13/pflag"

	"github.com/superproj/onex/pkg/generated/clientset/versioned"
)

func main() {
	defaultKubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config.local.onex")
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

	if err := clientset.AppsV1beta1().Miners("default").DeleteCollection(context.Background(), metav1.DeleteOptions{}, metav1.ListOptions{
		ResourceVersion: "0",
	}); err != nil {
		panic(err)
	}
}
