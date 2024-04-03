#!/usr/bin/env bash

# Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
# Use of this source code is governed by a MIT style
# license that can be found in the LICENSE file. The original repo for
# this file is https://github.com/superproj/onex.
#


# This script sets up a temporary Kubernetes GOPATH and runs an arbitrary
# command under it. Go tooling requires that the current directory be under
# GOPATH or else it fails to find some things, such as the vendor directory for
# the project.
# Usage: `hack/run-in-gopath.sh <command>`.

set -o errexit
set -o nounset
set -o pipefail

ONEX_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
source "${ONEX_ROOT}/scripts/lib/init.sh"

# This sets up a clean GOPATH and makes sure we are currently in it.
onex::golang::setup_env

# Run the user-provided command.
"${@}"
