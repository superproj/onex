// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package initializer

import (
	"k8s.io/apiserver/pkg/admission"

	clientset "github.com/superproj/onex/pkg/generated/clientset/versioned"
	"github.com/superproj/onex/pkg/generated/informers"
)

// WantsInternalMinerInformerFactory defines a function which sets InformerFactory for admission plugins that need it.
type WantsInternalMinerInformerFactory interface {
	admission.InitializationValidator
	SetInternalMinerInformerFactory(informers.SharedInformerFactory)
}

// WantsExternalMinerClientSet defines a function which sets external ClientSet for admission plugins that need it.
type WantsExternalMinerClientSet interface {
	admission.InitializationValidator
	SetExternalMinerClientSet(clientset.Interface)
}
