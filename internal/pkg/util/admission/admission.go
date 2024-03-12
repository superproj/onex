// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package admission

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/admission"
	"k8s.io/apiserver/pkg/authentication/user"

	"github.com/superproj/onex/internal/pkg/known"
)

func IsDeleteOperation(a admission.Attributes) bool {
	if a.GetOperation() == admission.Delete {
		return true
	}

	if a.GetOperation() == admission.Update {
		obj, ok := a.GetObject().(metav1.Object)
		if ok && !obj.GetDeletionTimestamp().IsZero() {
			return true
		}
	}

	return false
}

func IsAdmin(a admission.Attributes) bool {
	if a.GetUserInfo() == nil {
		return false
	}

	u := a.GetUserInfo().GetName()

	return u == known.AdminUsername || u == user.APIServerUser
}
