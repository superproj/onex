// Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/superproj/onex.
//

package kube

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	restclient "k8s.io/client-go/rest"

	"github.com/superproj/onex/pkg/version"
)

const (
	unknowString = "unknown"
)

// buildUserAgent builds a User-Agent string from given args.
func buildUserAgent(command, version, os, arch, commit string) string {
	return fmt.Sprintf(
		"%s/%s (%s/%s) onex.io/%s", command, version, os, arch, commit)
}

// DefaultOneXAPIUserAgent returns a User-Agent string built from static global vars.
func DefaultOneXUserAgent() string {
	return buildUserAgent(
		adjustCommand(os.Args[0]),
		adjustVersion(version.Get().GitVersion),
		runtime.GOOS,
		runtime.GOARCH,
		adjustCommit(version.Get().GitCommit))
}

// SetOneXDefaults sets default values on the provided client config for accessing the
// OneX API or returns an error if any of the defaults are impossible or invalid.
func SetOneXDefaults(config *restclient.Config) {
	if len(config.UserAgent) == 0 {
		config.UserAgent = DefaultOneXUserAgent()
	}
}

// adjustSourceName returns the name of the source calling the client.
func adjustSourceName(c string) string {
	if c == "" {
		return unknowString
	}
	return c
}

// adjustCommit returns sufficient significant figures of the commit's git hash.
func adjustCommit(c string) string {
	if c == "" {
		return unknowString
	}
	if len(c) > 7 {
		return c[:7]
	}
	return c
}

// adjustVersion strips "alpha", "beta", etc. from version in form
// major.minor.patch-[alpha|beta|etc].
func adjustVersion(v string) string {
	if v == "" {
		return unknowString
	}
	seg := strings.SplitN(v, "-", 2)
	return seg[0]
}

// adjustCommand returns the last component of the
// OS-specific command path for use in User-Agent.
func adjustCommand(p string) string {
	// Unlikely, but better than returning "".
	if p == "" {
		return unknowString
	}
	return filepath.Base(p)
}

func GetUserAgent(userAgent string) string {
	return DefaultOneXUserAgent() + "/" + adjustSourceName(userAgent)
}

func AddUserAgent(config *restclient.Config, userAgent string) *restclient.Config {
	config.UserAgent = GetUserAgent(userAgent)
	return config
}
