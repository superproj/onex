// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

// +k8s:deepcopy-gen=package
// +k8s:conversion-gen=github.com/superproj/onex/internal/controller/minerset/apis/config
// +k8s:conversion-gen=k8s.io/component-base/config/v1alpha1
// +k8s:conversion-gen-external-types=github.com/superproj/onex/internal/controller/minerset/apis/config/v1beta1
// +k8s:defaulter-gen=TypeMeta
// +k8s:defaulter-gen-input=github.com/superproj/onex/internal/controller/minerset/apis/config/v1beta1
// +groupName=minersetcontroller.config.onex.io

package v1beta1 // import "github.com/superproj/onex/internal/controller/minerset/apis/config/v1beta1"
