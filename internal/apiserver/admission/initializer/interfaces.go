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

// WantsExternalInformerFactory defines a function which sets InformerFactory for admission plugins that need it.
type WantsExternalInformerFactory interface {
	admission.InitializationValidator
	SetExternalInformerFactory(informers.SharedInformerFactory)
}

// WantsExternalClientSet defines a function which sets external ClientSet for admission plugins that need it.
type WantsExternalClientSet interface {
	admission.InitializationValidator
	SetExternalClientSet(clientset.Interface)
}
