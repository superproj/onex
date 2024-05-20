// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

// Package verflag defines utility functions to handle command line flags
// related to version of Kubernetes.
package version

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	flag "github.com/spf13/pflag"
)

type versionValue int

const (
	VersionFalse versionValue = 0
	VersionTrue  versionValue = 1
	VersionRaw   versionValue = 2
)

const strRawVersion string = "raw"

func (v *versionValue) IsBoolFlag() bool {
	return true
}

func (v *versionValue) Get() any {
	return *v
}

func (v *versionValue) Set(s string) error {
	if s == strRawVersion {
		*v = VersionRaw
		return nil
	}

	if strings.HasPrefix(s, "v") {
		err := SetDynamicVersion(s)
		if err == nil {
			value, _ := strconv.Atoi(s)
			*v = versionValue(value)
		}
		return err
	}

	boolVal, err := strconv.ParseBool(s)
	if boolVal {
		*v = VersionTrue
	} else {
		*v = VersionFalse
	}
	return err
}

func (v *versionValue) String() string {
	if *v == VersionRaw {
		return strRawVersion
	}
	return fmt.Sprintf("%v", bool(*v == VersionTrue))
}

// The type of the flag as required by the pflag.Value interface.
func (v *versionValue) Type() string {
	return "version"
}

func VersionVar(p *versionValue, name string, value versionValue, usage string) {
	*p = value
	flag.Var(p, name, usage)
	// "--version" will be treated as "--version=true"
	flag.Lookup(name).NoOptDefVal = "true"
}

func Version(name string, value versionValue, usage string) *versionValue {
	p := new(versionValue)
	VersionVar(p, name, value, usage)
	return p
}

const versionFlagName = "version"

var versionFlag = Version(versionFlagName, VersionFalse, "Print version information and quit")

// AddFlags registers this package's flags on arbitrary FlagSets, such that they point to the
// same value as the global flags.
func AddFlags(fs *flag.FlagSet) {
	fs.AddFlag(flag.Lookup(versionFlagName))
}

// variables for unit testing PrintAndExitIfRequested
var (
	output = io.Writer(os.Stdout)
	exit   = os.Exit
)

// PrintAndExitIfRequested will check if the -version flag was passed
// and, if so, print the version and exit.
func PrintAndExitIfRequested(appName string) {
	if *versionFlag == VersionRaw {
		fmt.Fprintf(output, "%s\n", Get().Text())
		exit(0)
	} else if *versionFlag == VersionTrue {
		fmt.Fprintf(output, "%s %s\n", appName, Get().GitVersion)
		exit(0)
	}
}
