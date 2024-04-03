// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package validation

import (
	"context"
	"reflect"
	"strings"

	"k8s.io/klog/v2"
)

func GetValidateFuncs(v any) map[string]reflect.Value {
	funcs := make(map[string]reflect.Value)
	typeOf := reflect.TypeOf(v)
	valueOf := reflect.ValueOf(v)
	for i := 0; i < typeOf.NumMethod(); i++ {
		m := typeOf.Method(i)
		val := valueOf.MethodByName(m.Name)
		if !val.IsValid() {
			continue
		}

		if !strings.HasPrefix(m.Name, "Validate") {
			continue
		}

		typ := val.Type()
		if typ.NumIn() != 2 || typ.NumOut() != 1 {
			continue
		}

		if typ.In(0) != reflect.TypeOf((*context.Context)(nil)).Elem() {
			continue
		}

		if typ.In(1).Kind() != reflect.Pointer {
			continue
		}

		vName := typ.In(1).Elem().Name()
		if m.Name != ("Validate" + vName) {
			continue
		}

		if typ.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
			continue
		}

		klog.V(4).InfoS("Register validator", "validator", vName)
		funcs[vName] = val
	}

	return funcs
}
