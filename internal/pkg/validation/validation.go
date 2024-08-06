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

	"github.com/google/wire"
	"k8s.io/klog/v2"

	"github.com/superproj/onex/internal/pkg/middleware/validate"
)

// Validator implements the validate.IValidator interface.
type Validator struct {
	registry map[string]reflect.Value
}

// ProviderSet is the validator providers.
var ProviderSet = wire.NewSet(NewValidator, wire.Bind(new(validate.IValidator), new(*Validator)))

// NewValidator creates and initializes a custom validator.
func NewValidator(customValidator any) *Validator {
	return &Validator{registry: extractValidationMethods(customValidator)}
}

// Validate validates the request using the appropriate validation method.
func (v *Validator) Validate(ctx context.Context, request any) error {
	validationFunc, ok := v.registry[reflect.TypeOf(request).Elem().Name()]
	if !ok {
		return nil // No validation function found for the request type
	}

	result := validationFunc.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(request)})
	if !result[0].IsNil() {
		return result[0].Interface().(error)
	}

	return nil
}

// extractValidationMethods extracts and returns a map of validation functions
// from the provided custom validator.
func extractValidationMethods(customValidator any) map[string]reflect.Value {
	funcs := make(map[string]reflect.Value)
	validatorType := reflect.TypeOf(customValidator)
	validatorValue := reflect.ValueOf(customValidator)

	for i := 0; i < validatorType.NumMethod(); i++ {
		method := validatorType.Method(i)
		methodValue := validatorValue.MethodByName(method.Name)

		if !methodValue.IsValid() || !strings.HasPrefix(method.Name, "Validate") {
			continue
		}

		methodType := methodValue.Type()

		// Ensure the method takes a context.Context and a pointer
		if methodType.NumIn() != 2 || methodType.NumOut() != 1 ||
			methodType.In(0) != reflect.TypeOf((*context.Context)(nil)).Elem() ||
			methodType.In(1).Kind() != reflect.Pointer {
			continue
		}

		// Ensure the method name matches the expected naming convention
		requestTypeName := methodType.In(1).Elem().Name()
		if method.Name != ("Validate" + requestTypeName) {
			continue
		}

		// Ensure the return type is error
		if methodType.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
			continue
		}

		klog.V(4).InfoS("Registering validator", "validator", requestTypeName)
		funcs[requestTypeName] = methodValue
	}

	return funcs
}
