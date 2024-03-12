#!/bin/bash

# Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
# Use of this source code is governed by a MIT style
# license that can be found in the LICENSE file. The original repo for
# this file is https://github.com/superproj/onex.
#


set -o errexit
set -o nounset
set -o pipefail

HADOLINT_VER=${1:-latest}
HADOLINT_FAILURE_THRESHOLD=${2:-warning}

FILES=$(find -- * -name Dockerfile)
while read -r file; do
  echo "Linting: ${file}"
  # Configure the linter to fail for warnings and errors. Can be set to: error | warning | info | style | ignore | none
  docker run --rm -i ghcr.io/hadolint/hadolint:"${HADOLINT_VER}" hadolint --failure-threshold "${HADOLINT_FAILURE_THRESHOLD}" - < "${file}"
done <<< "${FILES}"
