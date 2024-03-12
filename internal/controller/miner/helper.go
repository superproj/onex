// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package miner

import (
	"github.com/caarlos0/env/v8"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type MinerEnv struct {
	Image *string `env:"IMAGE"`
}

// HasMatchingLabels verifies that the Label Selector matches the given Labels.
func HasMatchingLabels(matchSelector metav1.LabelSelector, matchLabels map[string]string) bool {
	// This should never fail, validating webhook should catch this first
	selector, err := metav1.LabelSelectorAsSelector(&matchSelector)
	if err != nil {
		return false
	}
	// If a nil or empty selector creeps in, it should match nothing, not everything.
	if selector.Empty() {
		return false
	}
	if !selector.Matches(labels.Set(matchLabels)) {
		return false
	}
	return true
}

func GetMinerEnv() MinerEnv {
	var minerEnv MinerEnv
	_ = env.ParseWithOptions(&minerEnv, env.Options{Prefix: "ONEX_"})
	return minerEnv
}
