#!/usr/bin/env bash

# Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
# Use of this source code is governed by a MIT style
# license that can be found in the LICENSE file. The original repo for
# this file is https://github.com/superproj/onex.
#


# This file is not intended to be run automatically. It is meant to be run
# immediately before exporting docs. We do not want to check these documents in
# by default.

set -o errexit
set -o nounset
set -o pipefail

ONEX_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
source "${ONEX_ROOT}/scripts/lib/init.sh"

onex::golang::setup_env

BINS=(
  gen-docs
  gen-man
  gen-onex-docs
  gen-yaml
)
make build -C "${ONEX_ROOT}" BINS="${BINS[*]}"

onex::util::ensure-temp-dir

onex::util::gen-docs "${ONEX_TEMP}"

# remove all of the old docs
onex::util::remove-gen-docs

# Copy fresh docs into the repo.
# the shopt is so that we get docs/.generated_docs from the glob.
shopt -s dotglob
cp -af "${ONEX_TEMP}"/* "${ONEX_ROOT}"
shopt -u dotglob
