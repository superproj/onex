#!/usr/bin/env bash
# Copyright 2014 The Kubernetes Authors.
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

# shellcheck disable=2046 # printf word-splitting is intentional

set -o errexit
set -o nounset
set -o pipefail

# This tool wants a different default than usual.
ONEX_VERBOSE="${ONEX_VERBOSE:-1}"

ONEX_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
source "${ONEX_ROOT}/scripts/lib/init.sh"
source "${ONEX_ROOT}/scripts/lib/protoc.sh"
cd "${ONEX_ROOT}"

onex::golang::setup_env

DBG_CODEGEN="${DBG_CODEGEN:-0}"
GENERATED_FILE_PREFIX="${GENERATED_FILE_PREFIX:-zz_generated.}"
UPDATE_API_KNOWN_VIOLATIONS="${UPDATE_API_KNOWN_VIOLATIONS:-}"
API_KNOWN_VIOLATIONS_DIR="${API_KNOWN_VIOLATIONS_DIR:-"${ONEX_ROOT}/api/api-rules"}"

OUT_DIR="_output"
PRJ_SRC_PATH="k8s.io/kubernetes"
BOILERPLATE_FILENAME="hack/boilerplate/boilerplate.generatego.txt"
APPLYCONFIG_PKG="k8s.io/client-go/applyconfigurations"

# Any time we call sort, we want it in the same locale.
export LC_ALL="C"

# Work around for older grep tools which might have options we don't want.
unset GREP_OPTIONS

if [[ "${DBG_CODEGEN}" == 1 ]]; then
    onex::log::status "DBG: starting generated_files"
fi

function git_find() {
    # Similar to find but faster and easier to understand.  We want to include
    # modified and untracked files because this might be running against code
    # which is not tracked by git yet.
    git ls-files -cmo --exclude-standard ':!:manifests/*' ':!:third_party/*' ':!:vendor/*' "$@"
}

function git_grep() {
    # We want to include modified and untracked files because this might be
    # running against code which is not tracked by git yet.
    # We need vendor exclusion added at the end since it has to be part of
    # the pathspecs which are specified last.
    git grep --untracked "$@" ':!:vendor/*'
}

# protobuf generation
#
# Some of the later codegens depend on the results of this, so it needs to come
# first in the case of regenerating everything.
function codegen::protobuf() {
    # NOTE: All output from this script needs to be copied back to the calling
    # source tree.  This is managed in onex::build::copy_output in build/common.sh.
    # If the output set is changed update that function.

    local apis=()
    onex::util::read-array apis < <(
        git grep --untracked --null -l \
            -e '// +k8s:protobuf-gen=package' \
            -- \
            cmd pkg \
            | xargs -0 -n1 dirname \
            | sed 's|^|github.com/superproj/onex/|;s|k8s.io/kubernetes/staging/src/||' \
            | sort -u)

    onex::log::status "Generating protobufs for ${#apis[@]} targets"
    if [[ "${DBG_CODEGEN}" == 1 ]]; then
        onex::log::status "DBG: generating protobufs for:"
        for dir in "${apis[@]}"; do
            onex::log::status "DBG:     $dir"
        done
    fi

    git_find -z \
        ':(glob)**/generated.proto' \
        ':(glob)**/generated.pb.go' \
        | xargs -0 rm -f

    ${ONEX_ROOT}/scripts/update-generated-protobuf-dockerized.sh "${apis[@]}"
}

codegen::protobuf
