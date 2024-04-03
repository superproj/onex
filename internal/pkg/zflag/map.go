// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package zflag

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"
)

// -- map[string]string Value.
type mapValue map[string]string

func newMapValue(val map[string]string, p *map[string]string) *mapValue {
	*p = val
	return (*mapValue)(p)
}

func (mv *mapValue) Set(value string) error {
	if value == "" {
		return nil
	}
	for key := range *mv {
		delete(*mv, key)
	}

	for _, part := range strings.Split(value, ",") {
		parts := strings.SplitN(part, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid map value: %s", value)
		}
		key := parts[0]
		val := parts[1]
		(*mv)[key] = val
	}

	return nil
}

func (mv *mapValue) Type() string {
	return "map"
}

func (mv *mapValue) String() string {
	pairs := make([]string, 0, len(*mv))
	for k, v := range *mv {
		pairs = append(pairs, fmt.Sprintf("%s:%s", k, v))
	}
	return strings.Join(pairs, ",")
}

func MapVar(p *map[string]string, name string, value map[string]string, usage string, fss ...*pflag.FlagSet) {
	fs := pflag.CommandLine
	if len(fss) >= 1 {
		fs = fss[0]
	}

	fs.VarP(newMapValue(value, p), name, "", usage)
}

func MapVarP(p *map[string]string, name, shorthand string, value map[string]string, usage string, fss ...*pflag.FlagSet) {
	fs := pflag.CommandLine
	if len(fss) >= 1 {
		fs = fss[0]
	}

	fs.VarP(newMapValue(value, p), name, shorthand, usage)
}

func Map(name string, value map[string]string, usage string, fss ...*pflag.FlagSet) map[string]string {
	fs := pflag.CommandLine
	if len(fss) >= 1 {
		fs = fss[0]
	}

	p := make(map[string]string)
	MapVarP(&p, name, "", value, usage, fs)
	return p
}

func MapP(name, shorthand string, value map[string]string, usage string, fss ...*pflag.FlagSet) map[string]string {
	fs := pflag.CommandLine
	if len(fss) >= 1 {
		fs = fss[0]
	}

	p := make(map[string]string)
	MapVarP(&p, name, shorthand, value, usage, fs)
	return p
}
