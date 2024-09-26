#!/usr/bin/env bash

# Copyright 2019 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# This script switches to the preferred version for specified module.
# Usage: `scripts/pin-dependency.sh $MODULE $SHA-OR-TAG`.
# Example: `scripts/pin-dependency.sh github.com/docker/docker 501cb131a7b7`.

set -o errexit
set -o nounset
set -o pipefail

ONEX_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
source "${ONEX_ROOT}/scripts/lib/init.sh"

# Detect problematic GOPROXY settings that prevent lookup of dependencies
if [[ "${GOPROXY:-}" == "off" ]]; then
  onex::log::error "Cannot run with \$GOPROXY=off"
  exit 1
fi

onex::golang::setup_env
onex::util::require-jq

# Explicitly set GOFLAGS to ignore vendor, since GOFLAGS=-mod=vendor breaks dependency resolution while rebuilding vendor
export GOWORK=off
export GOFLAGS=-mod=mod

dep="${1:-}"
sha="${2:-}"

# Specifying a different repo is optional.
replacement=
case ${dep} in
    *=*)
        # shellcheck disable=SC2001
        replacement=$(echo "${dep}" | sed -e 's/.*=//')
        # shellcheck disable=SC2001
        dep=$(echo "${dep}" | sed -e 's/=.*//')
        ;;
    *)
        replacement="${dep}"
        ;;
esac

if [[ -z "${dep}" || -z "${replacement}" || -z "${sha}" ]]; then
  echo "Usage:"
  echo "  scripts/pin-dependency.sh \$MODULE[=\$REPLACEMENT] \$SHA-OR-TAG"
  echo ""
  echo "Examples:"
  echo "  scripts/pin-dependency.sh github.com/docker/docker 501cb131a7b7"
  echo "  scripts/pin-dependency.sh github.com/docker/docker=github.com/johndoe/docker my-experimental-branch"
  echo ""
  echo "Replacing with a different repository is useful for testing but"
  echo "the result should never be merged into Kubernetes!"
  echo ""
  exit 1
fi

# Find the resolved version before trying to use it.
echo "Running: go mod download ${replacement}@${sha}"
if meta=$(go mod download -json "${replacement}@${sha}"); then
    rev=$(echo "${meta}" | jq -r ".Version")
else
    error=$(echo "${meta}" | jq -r ".Error")
    echo "Download failed: ${error}" >&2
    exit 1
fi
echo "Resolved to ${replacement}@${rev}"

# Add the require directive
echo "Running: go mod edit -require ${dep}@${rev}"
go mod edit -require "${dep}@${rev}"

# Add the replace directive
if [ "${replacement}" != "${dep}" ]; then
  echo "Running: go mod edit -replace ${dep}=${replacement}@${rev}"
  go mod edit -replace "${dep}=${replacement}@${rev}"
fi

echo ""
echo "Run scripts/update-vendor.sh to rebuild the vendor directory"
