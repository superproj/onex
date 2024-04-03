// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package reflect

import (
	"fmt"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

func ToGormDBMap(obj any, fields []string) (map[string]any, error) {
	reflectType := reflect.ValueOf(obj).Type()
	reflectValue := reflect.ValueOf(obj)
	for reflectType.Kind() == reflect.Slice || reflectType.Kind() == reflect.Ptr {
		reflectType = reflectType.Elem()
		reflectValue = reflect.ValueOf(obj).Elem()
	}

	ret := make(map[string]any, 0)
	for _, f := range fields {
		fs, exist := reflectType.FieldByName(f)
		if !exist {
			return nil, fmt.Errorf("unknow field " + f)
		}

		tagMap := parseTagSetting(fs.Tag)
		gormfiled, exist := tagMap["COLUMN"]
		if !exist {
			return nil, fmt.Errorf("undef gorm field " + f)
		}

		ret[gormfiled] = reflectValue.FieldByName(f)
	}
	return ret, nil
}

func parseTagSetting(tags reflect.StructTag) map[string]string {
	setting := map[string]string{}
	for _, str := range []string{tags.Get("sql"), tags.Get("gorm")} {
		if str == "" {
			continue
		}
		tags := strings.Split(str, ";")
		for _, value := range tags {
			v := strings.Split(value, ":")
			k := strings.TrimSpace(strings.ToUpper(v[0]))
			if len(v) >= 2 {
				setting[k] = strings.Join(v[1:], ":")
			} else {
				setting[k] = k
			}
		}
	}
	return setting
}

func GetObjFieldsMap(obj any, fields []string) map[string]any {
	ret := make(map[string]any)

	modelReflect := reflect.ValueOf(obj)
	if modelReflect.Kind() == reflect.Ptr {
		modelReflect = modelReflect.Elem()
	}

	modelRefType := modelReflect.Type()
	fieldsCount := modelReflect.NumField()
	var fieldData any
	for i := 0; i < fieldsCount; i++ {
		field := modelReflect.Field(i)
		if len(fields) != 0 && !findString(fields, modelRefType.Field(i).Name) {
			continue
		}

		switch field.Kind() {
		case reflect.Struct:
			fallthrough
		case reflect.Ptr:
			fieldData = GetObjFieldsMap(field.Interface(), []string{})
		default:
			fieldData = field.Interface()
		}

		ret[modelRefType.Field(i).Name] = fieldData
	}

	return ret
}

func CopyObj(from any, to any, fields []string) (changed bool, err error) {
	fromMap := GetObjFieldsMap(from, fields)
	toMap := GetObjFieldsMap(to, fields)
	if reflect.DeepEqual(fromMap, toMap) {
		return false, nil
	}

	t := reflect.ValueOf(to).Elem()
	for k, v := range fromMap {
		val := t.FieldByName(k)
		val.Set(reflect.ValueOf(v))
	}
	return true, nil
}

// CopyObjViaYaml marshal "from" to yaml data, then unMarshal data to "to".
func CopyObjViaYaml(to any, from any) error {
	if from == nil || to == nil {
		return nil
	}

	data, err := yaml.Marshal(from)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, to)
}

// StructName used to get the struct name from the obj.
func StructName(obj any) string {
	t := reflect.TypeOf(obj)
	if t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	}

	return t.Name()
}

// findString return true if target in slice, return false if not.
func findString(slice []string, target string) bool {
	for _, str := range slice {
		if str == target {
			return true
		}
	}
	return false
}
