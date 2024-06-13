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
KUBE_VERBOSE="${KUBE_VERBOSE:-1}"

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
BOILERPLATE_FILENAME="scripts/boilerplate/boilerplate.go.txt"
ONEX_MODULE_NAME="github.com/superproj/onex"
PLURAL_EXCEPTIONS="Endpoints:Endpoints,ResourceClaimParameters:ResourceClaimParameters,ResourceClassParameters:ResourceClassParameters"
EXTRA_GENERATE_PKG="k8s.io/api/core/v1"
OUTPUT_PKG="github.com/superproj/onex/pkg/generated"
APPLYCONFIG_PKG="${OUTPUT_PKG}/applyconfigurations"

# Any time we call sort, we want it in the same locale.
export LC_ALL="C"

# Work around for older grep tools which might have options we don't want.
unset GREP_OPTIONS

if [[ "${DBG_CODEGEN}" == 1 ]]; then
    onex::log::status "DBG: starting generated_files"
fi

# Generate a list of directories we don't want to play in.
DIRS_TO_AVOID=()
onex::util::read-array DIRS_TO_AVOID < <(
    git ls-files -cmo --exclude-standard -- ':!:vendor/*' ':(glob)*/**/go.work' \
        | while read -r F; do \
            echo ':!:'"$(dirname "${F}")"; \
        done
    )

function git_find() {
    # Similar to find but faster and easier to understand.  We want to include
    # modified and untracked files because this might be running against code
    # which is not tracked by git yet.
    git ls-files -cmo --exclude-standard ':!:vendor/*' "${DIRS_TO_AVOID[@]}" "$@"
}

function git_grep() {
    # We want to include modified and untracked files because this might be
    # running against code which is not tracked by git yet.
    # We need vendor exclusion added at the end since it has to be part of
    # the pathspecs which are specified last.
    git grep --untracked "$@" ':!:vendor/*' "${DIRS_TO_AVOID[@]}"
}

# Generate a list of all files that have a `+k8s:` comment-tag.  This will be
# used to derive lists of files/dirs for generation tools.
if [[ "${DBG_CODEGEN}" == 1 ]]; then
    onex::log::status "DBG: finding all +k8s: tags"
fi
ALL_K8S_TAG_FILES=()
onex::util::read-array ALL_K8S_TAG_FILES < <(
    git_grep -l \
        -e '^// *+k8s:'                `# match +k8s: tags` \
        -- \
        ':!:*/testdata/*'              `# not under any testdata` \
        ':(glob)**/*.go'               `# in any *.go file` \
    )
if [[ "${DBG_CODEGEN}" == 1 ]]; then
    onex::log::status "DBG: found ${#ALL_K8S_TAG_FILES[@]} +k8s: tagged files"
fi

#
# Code generation logic.
#

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
            cmd pkg staging \
            | while read -r -d $'\0' F; do dirname "${F}"; done \
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

    if onex::protoc::check_protoc >/dev/null; then
      scripts/_update-generated-protobuf-dockerized.sh "${apis[@]}"
    else
      onex::log::status "protoc ${PROTOC_VERSION} not found (can install with scripts/install-protoc.sh); generating containerized..."
      build/run.sh scripts/_update-generated-protobuf-dockerized.sh "${apis[@]}"
    fi

    # Fix `pkg/apis/apps/v1beta1/generated.pb.go:49:10: undefined: ObjectReference` compile errors
    cp ${ONEX_ROOT}/manifests/generated.pb.go.fix ${ONEX_ROOT}/pkg/apis/apps/v1beta1/generated.pb.go
}

# Deep-copy generation
#
# Any package that wants deep-copy functions generated must include a
# comment-tag in column 0 of one file of the form:
#     // +k8s:deepcopy-gen=<VALUE>
#
# The <VALUE> may be one of:
#     generate: generate deep-copy functions into the package
#     register: generate deep-copy functions and register them with a
#               scheme
function codegen::deepcopy() {
    # Build the tool.
    GOPROXY=off go install \
        k8s.io/code-generator/cmd/deepcopy-gen

    # The result file, in each pkg, of deep-copy generation.
    local output_file="${GENERATED_FILE_PREFIX}deepcopy.go"

    # Find all the directories that request deep-copy generation.
    if [[ "${DBG_CODEGEN}" == 1 ]]; then
        onex::log::status "DBG: finding all +k8s:deepcopy-gen tags"
    fi
    local tag_dirs=()
    onex::util::read-array tag_dirs < <( \
        grep -l --null '+k8s:deepcopy-gen=' "${ALL_K8S_TAG_FILES[@]}" \
            | while read -r -d $'\0' F; do dirname "${F}"; done \
            | sort -u)
    if [[ "${DBG_CODEGEN}" == 1 ]]; then
        onex::log::status "DBG: found ${#tag_dirs[@]} +k8s:deepcopy-gen tagged dirs"
    fi

    local tag_pkgs=()
    for dir in "${tag_dirs[@]}"; do
        tag_pkgs+=("./$dir")
    done

    onex::log::status "Generating deepcopy code for ${#tag_pkgs[@]} targets"
    if [[ "${DBG_CODEGEN}" == 1 ]]; then
        onex::log::status "DBG: running deepcopy-gen for:"
        for dir in "${tag_dirs[@]}"; do
            onex::log::status "DBG:     $dir"
        done
    fi

    git_find -z ':(glob)**'/"${output_file}" | xargs -0 rm -f

    deepcopy-gen \
        -v "${KUBE_VERBOSE}" \
        --go-header-file "${BOILERPLATE_FILENAME}" \
        --output-file "${output_file}" \
        --bounding-dirs "${ONEX_MODULE_NAME},k8s.io/api" \
        "${tag_pkgs[@]}" \
        "$@"

    if [[ "${DBG_CODEGEN}" == 1 ]]; then
        onex::log::status "Generated deepcopy code"
    fi
}

# Generates types_swagger_doc_generated file for the given group version.
# $1: Name of the group version
# $2: Path to the directory where types.go for that group version exists. This
# is the directory where the file will be generated.
function gen_types_swagger_doc() {
    local group_version="$1"
    local gv_dir="$2"
    local tmpfile
    tmpfile="${TMPDIR:-/tmp}/types_swagger_doc_generated.$(date +%s).go"

    if [[ "${DBG_CODEGEN}" == 1 ]]; then
        onex::log::status "DBG: running gen-swaggertype-docs for ${group_version} at ${gv_dir}"
    fi

    {
        cat "${BOILERPLATE_FILENAME}"
        echo
        echo "package ${group_version##*/}"
        # Indenting here prevents the boilerplate checker from thinking this file
        # is generated - gofmt will fix the indents anyway.
        cat <<EOF

          // This file contains a collection of methods that can be used from go-restful to
          // generate Swagger API documentation for its models. Please read this PR for more
          // information on the implementation: https://github.com/emicklei/go-restful/pull/215
          //
          // TODOs are ignored from the parser (e.g. TODO(andronat):... || TODO:...) if and only if
          // they are on one line! For multiple line or blocks that you want to ignore use ---.
          // Any context after a --- is ignored.
          //
          // Those methods can be generated by using scripts/update-codegen.sh

          // AUTO-GENERATED FUNCTIONS START HERE. DO NOT EDIT.
EOF
    } > "${tmpfile}"

    for types in $(find ${gv_dir} -name "*types.go")
    do
        gen-swaggertype-docs \
            -s \
            "${types}" \
            -f - \
            >> "${tmpfile}"
    done

    echo "// AUTO-GENERATED FUNCTIONS END HERE" >> "${tmpfile}"

    gofmt -w -s "${tmpfile}"
    mv "${tmpfile}" "${gv_dir}/types_swagger_doc_generated.go"
}

# swagger generation
#
# Some of the later codegens depend on the results of this, so it needs to come
# first in the case of regenerating everything.
function codegen::swagger() {
    # Build the tool
    if ! command -v gen-swaggertype-docs &> /dev/null ; then
        GOPROXY=off go install ${ONEX_ROOT}/cmd/gen-swaggertype-docs
    fi

    local group_versions=()
    IFS=" " read -r -a group_versions <<< "${ONEX_AVAILABLE_GROUP_VERSIONS}"

    onex::log::status "Generating swagger for ${#group_versions[@]} targets"

    git_find -z ':(glob)**/types_swagger_doc_generated.go' | xargs -0 rm -f

    # Regenerate files.
    for group_version in "${group_versions[@]}"; do
      gen_types_swagger_doc "${group_version}" "$(onex::util::group-version-to-pkg-path "${group_version}")"
    done
}

# prerelease-lifecycle generation
#
# Any package that wants prerelease-lifecycle functions generated must include a
# comment-tag in column 0 of one file of the form:
#     // +k8s:prerelease-lifecycle-gen=true
function codegen::prerelease() {
    # Build the tool.
    GOPROXY=off go install \
        k8s.io/code-generator/cmd/prerelease-lifecycle-gen

    # The result file, in each pkg, of prerelease-lifecycle generation.
    local output_file="${GENERATED_FILE_PREFIX}prerelease-lifecycle.go"

    # Find all the directories that request prerelease-lifecycle generation.
    if [[ "${DBG_CODEGEN}" == 1 ]]; then
        onex::log::status "DBG: finding all +k8s:prerelease-lifecycle-gen tags"
    fi
    local tag_dirs=()
    onex::util::read-array tag_dirs < <( \
        grep -l --null '+k8s:prerelease-lifecycle-gen=true' "${ALL_K8S_TAG_FILES[@]}" \
            | while read -r -d $'\0' F; do dirname "${F}"; done \
            | sort -u)
    if [[ "${DBG_CODEGEN}" == 1 ]]; then
        onex::log::status "DBG: found ${#tag_dirs[@]} +k8s:prerelease-lifecycle-gen tagged dirs"
    fi

    local tag_pkgs=()
    for dir in "${tag_dirs[@]}"; do
        tag_pkgs+=("./$dir")
    done

    onex::log::status "Generating prerelease-lifecycle code for ${#tag_pkgs[@]} targets"
    if [[ "${DBG_CODEGEN}" == 1 ]]; then
        onex::log::status "DBG: running prerelease-lifecycle-gen for:"
        for dir in "${tag_dirs[@]}"; do
            onex::log::status "DBG:     $dir"
        done
    fi

    git_find -z ':(glob)**'/"${output_file}" | xargs -0 rm -f

    prerelease-lifecycle-gen \
        -v "${KUBE_VERBOSE}" \
        --go-header-file "${BOILERPLATE_FILENAME}" \
        --output-file "${output_file}" \
        "${tag_pkgs[@]}" \
        "$@"

    if [[ "${DBG_CODEGEN}" == 1 ]]; then
        onex::log::status "Generated prerelease-lifecycle code"
    fi
}

# Defaulter generation
#
# Any package that wants defaulter functions generated must include a
# comment-tag in column 0 of one file of the form:
#     // +k8s:defaulter-gen=<VALUE>
#
# The <VALUE> depends on context:
#     on types:
#       true:  always generate a defaulter for this type
#       false: never generate a defaulter for this type
#     on functions:
#       covers: if the function name matches SetDefault_NAME, instructs
#               the generator not to recurse
#     on packages:
#       FIELDNAME: any object with a field of this name is a candidate
#                  for having a defaulter generated
function codegen::defaults() {
    # Build the tool.
    GOPROXY=off go install \
        k8s.io/code-generator/cmd/defaulter-gen

    # The result file, in each pkg, of defaulter generation.
    local output_file="${GENERATED_FILE_PREFIX}defaults.go"

    # All directories that request any form of defaulter generation.
    if [[ "${DBG_CODEGEN}" == 1 ]]; then
        onex::log::status "DBG: finding all +k8s:defaulter-gen tags"
    fi
    local tag_dirs=()
    onex::util::read-array tag_dirs < <( \
        grep -l --null '+k8s:defaulter-gen=' "${ALL_K8S_TAG_FILES[@]}" \
            | while read -r -d $'\0' F; do dirname "${F}"; done \
            | sort -u)
    if [[ "${DBG_CODEGEN}" == 1 ]]; then
        onex::log::status "DBG: found ${#tag_dirs[@]} +k8s:defaulter-gen tagged dirs"
    fi

    local tag_pkgs=()
    for dir in "${tag_dirs[@]}"; do
        tag_pkgs+=("./$dir")
    done

    onex::log::status "Generating defaulter code for ${#tag_pkgs[@]} targets"
    if [[ "${DBG_CODEGEN}" == 1 ]]; then
        onex::log::status "DBG: running defaulter-gen for:"
        for dir in "${tag_dirs[@]}"; do
            onex::log::status "DBG:     $dir"
        done
    fi

    git_find -z ':(glob)**'/"${output_file}" | xargs -0 rm -f

    defaulter-gen \
        -v "${KUBE_VERBOSE}" \
        --go-header-file "${BOILERPLATE_FILENAME}" \
        --output-file "${output_file}" \
        "${tag_pkgs[@]}" \
        "$@"

    if [[ "${DBG_CODEGEN}" == 1 ]]; then
        onex::log::status "Generated defaulter code"
    fi
}

# Conversion generation

# Any package that wants conversion functions generated into it must
# include one or more comment-tags in its `doc.go` file, of the form:
#     // +k8s:conversion-gen=<INTERNAL_TYPES_DIR>
#
# The INTERNAL_TYPES_DIR is a project-local path to another directory
# which should be considered when evaluating peer types for
# conversions.  An optional additional comment of the form
#     // +k8s:conversion-gen-external-types=<EXTERNAL_TYPES_DIR>
#
# identifies where to find the external types; if there is no such
# comment then the external types are sought in the package where the
# `k8s:conversion` tag is found.
#
# Conversions, in both directions, are generated for every type name
# that is defined in both an internal types package and the external
# types package.
#
# TODO: it might be better in the long term to make peer-types explicit in the
# IDL.
function codegen::conversions() {
    # Build the tool.
    GOPROXY=off go install \
        k8s.io/code-generator/cmd/conversion-gen

    # The result file, in each pkg, of conversion generation.
    local output_file="${GENERATED_FILE_PREFIX}conversion.go"

    # All directories that request any form of conversion generation.
    if [[ "${DBG_CODEGEN}" == 1 ]]; then
        onex::log::status "DBG: finding all +k8s:conversion-gen tags"
    fi
    local tag_dirs=()
    onex::util::read-array tag_dirs < <(\
        grep -l --null '^// *+k8s:conversion-gen=' "${ALL_K8S_TAG_FILES[@]}" \
            | while read -r -d $'\0' F; do dirname "${F}"; done \
            | sort -u)
    if [[ "${DBG_CODEGEN}" == 1 ]]; then
        onex::log::status "DBG: found ${#tag_dirs[@]} +k8s:conversion-gen tagged dirs"
    fi

    local tag_pkgs=()
    for dir in "${tag_dirs[@]}"; do
        tag_pkgs+=("./$dir")
    done

    local extra_peer_pkgs=(
        k8s.io/kubernetes/pkg/apis/core
        k8s.io/kubernetes/pkg/apis/core/v1
        k8s.io/api/core/v1
    )

    onex::log::status "Generating conversion code for ${#tag_pkgs[@]} targets"
    if [[ "${DBG_CODEGEN}" == 1 ]]; then
        onex::log::status "DBG: running conversion-gen for:"
        for dir in "${tag_dirs[@]}"; do
            onex::log::status "DBG:     $dir"
        done
    fi

    git_find -z ':(glob)**'/"${output_file}" | xargs -0 rm -f

    conversion-gen \
        -v "${KUBE_VERBOSE}" \
        --go-header-file "${BOILERPLATE_FILENAME}" \
        --output-file "${output_file}" \
        $(printf -- " --extra-peer-dirs %s" "${extra_peer_pkgs[@]}") \
        "${tag_pkgs[@]}" \
        "$@"

    if [[ "${DBG_CODEGEN}" == 1 ]]; then
        onex::log::status "Generated conversion code"
    fi
}

# $@: directories to exclude
# example:
#    k8s_tag_files_except foo bat/qux
function k8s_tag_files_except() {
    for f in "${ALL_K8S_TAG_FILES[@]}"; do
        local excl=""
        for x in "$@"; do
            if [[ "$f" =~ "$x"/.* ]]; then
                excl="true"
                break
            fi
        done
        if [[ "${excl}" != true ]]; then
            echo "$f"
        fi
    done
}

# OpenAPI generation
#
# Any package that wants open-api functions generated must include a
# comment-tag in column 0 of one file of the form:
#     // +k8s:openapi-gen=true
function todo::codegen::openapi() {
    # Build the tool.
    GOPROXY=off go install \
        k8s.io/kube-openapi/cmd/openapi-gen

    # The result file, in each pkg, of open-api generation.
    local output_file="${GENERATED_FILE_PREFIX}openapi.go"

    local output_dir="pkg/generated/openapi"
    local output_pkg="k8s.io/kubernetes/${output_dir}"
    local known_violations_file="${API_KNOWN_VIOLATIONS_DIR}/violation_exceptions.list"

    local report_file="${OUT_DIR}/api_violations.report"
    # When UPDATE_API_KNOWN_VIOLATIONS is set to be true, let the generator to write
    # updated API violations to the known API violation exceptions list.
    if [[ "${UPDATE_API_KNOWN_VIOLATIONS}" == true ]]; then
        report_file="${known_violations_file}"
    fi

    if [[ "${DBG_CODEGEN}" == 1 ]]; then
        onex::log::status "DBG: finding all +k8s:openapi-gen tags"
    fi

    local tag_files=()
    onex::util::read-array tag_files < <(
        k8s_tag_files_except \
            staging/src/k8s.io/code-generator \
            staging/src/k8s.io/sample-apiserver
        )

    local tag_dirs=()
    onex::util::read-array tag_dirs < <(
        grep -l --null '+k8s:openapi-gen=' "${tag_files[@]}" \
            | while read -r -d $'\0' F; do dirname "${F}"; done \
            | sort -u)

    if [[ "${DBG_CODEGEN}" == 1 ]]; then
        onex::log::status "DBG: found ${#tag_dirs[@]} +k8s:openapi-gen tagged dirs"
    fi

    local tag_pkgs=()
    for dir in "${tag_dirs[@]}"; do
        tag_pkgs+=("./$dir")
    done

    onex::log::status "Generating openapi code"
    if [[ "${DBG_CODEGEN}" == 1 ]]; then
        onex::log::status "DBG: running openapi-gen for:"
        for dir in "${tag_dirs[@]}"; do
            onex::log::status "DBG:     $dir"
        done
    fi

    git_find -z ':(glob)pkg/generated/**'/"${output_file}" | xargs -0 rm -f

    openapi-gen \
        -v "${KUBE_VERBOSE}" \
        --go-header-file "${BOILERPLATE_FILENAME}" \
        --output-file "${output_file}" \
        --output-dir "${output_dir}" \
        --output-pkg "${output_pkg}" \
        --report-filename "${report_file}" \
        "${tag_pkgs[@]}" \
        "$@"

    touch "${report_file}"
    local known_filename="${known_violations_file}"
    if ! diff -u "${known_filename}" "${report_file}"; then
        echo -e "ERROR:"
        echo -e "\tAPI rule check failed - reported violations differ from known violations"
        echo -e "\tPlease read api/api-rules/README.md to resolve the failure in ${known_filename}"
    fi

    if [[ "${DBG_CODEGEN}" == 1 ]]; then
        onex::log::status "Generated openapi code"
    fi
}

# OpenAPI generation
#
# Any package that wants open-api functions generated must include a
# comment-tag in column 0 of one file of the form:
#     // +k8s:openapi-gen=true
function codegen::openapi() {
    # Build the tool.
    # Please make sure to use openapi-gen version v0.29.3 here.
    if ! command -v gen-swaggertype-docs &> /dev/null ; then
        GOPROXY=off go install k8s.io/code-generator/cmd/openapi-gen@v0.29.3
    fi

    # The result file, in each pkg, of open-api generation.
    local output_file="${GENERATED_FILE_PREFIX}openapi"

    local output_dir="pkg/generated/openapi"
    local output_pkg="github.com/superproj/onex/${output_dir}"
    local known_violations_file="${API_KNOWN_VIOLATIONS_DIR}/violation_exceptions.list"

    local report_file="${OUT_DIR}/api_violations.report"
    # When UPDATE_API_KNOWN_VIOLATIONS is set to be true, let the generator to write
    # updated API violations to the known API violation exceptions list.
    if [[ "${UPDATE_API_KNOWN_VIOLATIONS}" == true ]]; then
        report_file="${known_violations_file}"
    fi

    if [[ "${DBG_CODEGEN}" == 1 ]]; then
        onex::log::status "DBG: finding all +k8s:openapi-gen tags"
    fi

    local tag_files=()
    onex::util::read-array tag_files < <(
        k8s_tag_files_except \
            staging/src/k8s.io/code-generator \
            staging/src/k8s.io/sample-apiserver
        )

    local tag_dirs=()
    onex::util::read-array tag_dirs < <(
        grep -l --null '+k8s:openapi-gen=' "${tag_files[@]}" \
            | while read -r -d $'\0' F; do dirname "${F}"; done \
            | sort -u)

    if [[ "${DBG_CODEGEN}" == 1 ]]; then
        onex::log::status "DBG: found ${#tag_dirs[@]} +k8s:openapi-gen tagged dirs"
    fi

    local tag_pkgs=()
    for dir in "${tag_dirs[@]}"; do
        tag_pkgs+=("./$dir")
    done

    onex::log::status "Generating openapi code"
    if [[ "${DBG_CODEGEN}" == 1 ]]; then
        onex::log::status "DBG: running openapi-gen for:"
        for dir in "${tag_dirs[@]}"; do
            onex::log::status "DBG:     $dir"
        done
    fi

    git_find -z ':(glob)pkg/generated/**'/"${output_file}" | xargs -0 rm -f

    openapi-gen \
        -v "${KUBE_VERBOSE}" \
        --go-header-file "${BOILERPLATE_FILENAME}" \
        -O "${output_file}" \
        -i 'k8s.io/apimachinery/pkg/apis/meta/v1,k8s.io/apimachinery/pkg/runtime,k8s.io/apimachinery/pkg/version,k8s.io/kubernetes/pkg/apis/core,k8s.io/api/core/v1,k8s.io/api/autoscaling/v1,k8s.io/api/coordination/v1,github.com/superproj/onex/pkg/apis/apps/v1beta1' \
        --output-base "${GOPATH}/src" \
        -p "${output_pkg}" \
        --report-filename "${report_file}" \
        "${tag_pkgs[@]}" \
        "$@"

    if [[ "${DBG_CODEGEN}" == 1 ]]; then
        onex::log::status "Generated openapi code"
    fi
}

function codegen::applyconfigs() {
    GOPROXY=off go install \
        k8s.io/kubernetes/pkg/generated/openapi/cmd/models-schema \
        k8s.io/code-generator/cmd/applyconfiguration-gen

    local ext_apis=()
    onex::util::read-array ext_apis < <(
        cd "${ONEX_ROOT}"
        git_find -z ':(glob)pkg/apis/**/*types.go' \
            | while read -r -d $'\0' F; do dirname "${ONEX_MODULE_NAME}/${F}"; done \
            | sort -u)
    ext_apis+=("k8s.io/apimachinery/pkg/apis/meta/v1" "k8s.io/api/core/v1" "k8s.io/api/autoscaling/v1")

    onex::log::status "Generating apply-config code for ${#ext_apis[@]} targets"
    if [[ "${DBG_CODEGEN}" == 1 ]]; then
        onex::log::status "DBG: running applyconfiguration-gen for:"
        for api in "${ext_apis[@]}"; do
            onex::log::status "DBG:     $api"
        done
    fi

    (git_grep -l --null \
        -e '^// Code generated by applyconfiguration-gen. DO NOT EDIT.$' \
        -- \
        ':(glob)pkg/generated/**/*.go' \
        || true) \
        | xargs -0 rm -f

    applyconfiguration-gen \
        -v "${KUBE_VERBOSE}" \
        --openapi-schema <(models-schema) \
        --go-header-file "${BOILERPLATE_FILENAME}" \
        --output-dir "${ONEX_ROOT}/pkg/generated/applyconfigurations" \
        --output-pkg "${APPLYCONFIG_PKG}" \
        "${ext_apis[@]}" \
        "$@"

    if [[ "${DBG_CODEGEN}" == 1 ]]; then
        onex::log::status "Generated apply-config code"
    fi
}

function codegen::clients() {
    GOPROXY=off go install \
        k8s.io/code-generator/cmd/client-gen

    IFS=" " read -r -a group_versions <<< "${ONEX_AVAILABLE_GROUP_VERSIONS}"
    local gv_dirs=()
    for gv in "${group_versions[@]}"; do
        # add items, but strip off any leading apis/ you find to match command expectations
        local api_dir
        api_dir=$(onex::util::group-version-to-pkg-path "${gv}")
        local nopkg_dir=${api_dir#pkg/}
        nopkg_dir=${nopkg_dir#staging/src/k8s.io/api/}
        local pkg_dir=${nopkg_dir#apis/}

        # skip groups that aren't being served, clients for these don't matter
        if [[ " ${KUBE_NONSERVER_GROUP_VERSIONS} " == *" ${gv} "* ]]; then
          continue
        fi

        gv_dirs+=("${pkg_dir}")
    done
    gv_dirs+=("${EXTRA_GENERATE_PKG}")

    onex::log::status "Generating client code for ${#gv_dirs[@]} targets"
    if [[ "${DBG_CODEGEN}" == 1 ]]; then
        onex::log::status "DBG: running client-gen for:"
        for dir in "${gv_dirs[@]}"; do
            onex::log::status "DBG:     $dir"
        done
    fi

    (git_grep -l --null \
        -e '^// Code generated by client-gen. DO NOT EDIT.$' \
        -- \
        ':(glob)pkg/generated/**/*.go' \
        || true) \
        | xargs -0 rm -f

    # UPDATEME: When add new k8s resource.
    client-gen \
        -v "${KUBE_VERBOSE}" \
        --go-header-file "${BOILERPLATE_FILENAME}" \
        --included-types-overrides core/v1/Namespace,core/v1/ConfigMap,core/v1/Event,core/v1/Secret \
        --output-dir "${ONEX_ROOT}/pkg/generated/clientset" \
        --output-pkg="${OUTPUT_PKG}/clientset" \
        --clientset-name="versioned" \
        --input-base="" \
        --plural-exceptions "${PLURAL_EXCEPTIONS}" \
        --apply-configuration-package "${APPLYCONFIG_PKG}" \
        $(printf -- " --input %s" "${gv_dirs[@]}") \
        "$@"

    # Fix generated namespace clients
    ${ONEX_ROOT}/scripts/fix-generated-client.sh

    if [[ "${DBG_CODEGEN}" == 1 ]]; then
        onex::log::status "Generated client code"
    fi
}

function codegen::listers() {
    if ! command -v gen-swaggertype-docs &> /dev/null ; then
        GOPROXY=off go install k8s.io/code-generator/cmd/lister-gen
    fi

    local ext_apis=()
    onex::util::read-array ext_apis < <(
        cd "${ONEX_ROOT}"
        git_find -z ':(glob)pkg/apis/**/*types.go' \
            | while read -r -d $'\0' F; do dirname "github.com/superproj/onex/${F}"; done \
            | sort -u)
    ext_apis+=("${EXTRA_GENERATE_PKG}")

    onex::log::status "Generating lister code for ${#ext_apis[@]} targets"
    if [[ "${DBG_CODEGEN}" == 1 ]]; then
        onex::log::status "DBG: running lister-gen for:"
        for api in "${ext_apis[@]}"; do
            onex::log::status "DBG:     $api"
        done
    fi

    (git_grep -l --null \
        -e '^// Code generated by lister-gen. DO NOT EDIT.$' \
        -- \
        ':(glob)pkg/generated/**/*.go' \
        || true) \
        | xargs -0 rm -f

    lister-gen \
        -v "${KUBE_VERBOSE}" \
        --go-header-file "${BOILERPLATE_FILENAME}" \
        --included-types-overrides core/v1/Namespace,core/v1/ConfigMap,core/v1/Event,core/v1/Secret \
        --output-dir "${ONEX_ROOT}/pkg/generated/listers" \
        --output-pkg "${OUTPUT_PKG}/listers" \
        --plural-exceptions "${PLURAL_EXCEPTIONS}" \
        "${ext_apis[@]}" \
        "$@"

    if [[ "${DBG_CODEGEN}" == 1 ]]; then
        onex::log::status "Generated lister code"
    fi
}

function codegen::informers() {
    if ! command -v gen-swaggertype-docs &> /dev/null ; then
        GOPROXY=off go install k8s.io/code-generator/cmd/informer-gen
    fi

    local ext_apis=()
    onex::util::read-array ext_apis < <(
        cd "${ONEX_ROOT}"
        git_find -z ':(glob)pkg/apis/**/*types.go' \
            | while read -r -d $'\0' F; do dirname "github.com/superproj/onex/${F}"; done \
            | sort -u)
    ext_apis+=("${EXTRA_GENERATE_PKG}")

    onex::log::status "Generating informer code for ${#ext_apis[@]} targets"
    if [[ "${DBG_CODEGEN}" == 1 ]]; then
        onex::log::status "DBG: running informer-gen for:"
        for api in "${ext_apis[@]}"; do
            onex::log::status "DBG:     $api"
        done
    fi

    (git_grep -l --null \
        -e '^// Code generated by informer-gen. DO NOT EDIT.$' \
        -- \
        ':(glob)pkg/generated/**/*.go' \
        || true) \
        | xargs -0 rm -f

    informer-gen \
        -v "${KUBE_VERBOSE}" \
        --go-header-file "${BOILERPLATE_FILENAME}" \
        --included-types-overrides core/v1/Namespace,core/v1/ConfigMap,core/v1/Event,core/v1/Secret \
        --output-dir "${ONEX_ROOT}/pkg/generated/informers" \
        --output-pkg "${OUTPUT_PKG}/informers" \
        --single-directory \
        --versioned-clientset-package "${OUTPUT_PKG}/clientset/versioned" \
        --listers-package "${OUTPUT_PKG}/listers" \
        --plural-exceptions "${PLURAL_EXCEPTIONS}" \
        "${ext_apis[@]}" \
        "$@"

    if [[ "${DBG_CODEGEN}" == 1 ]]; then
        onex::log::status "Generated informer code"
    fi
}

function indent() {
    while read -r X; do
        echo "    ${X}"
    done
}

function unused::codegen::subprojects() {
    # Call generation on sub-projects.
    local subs=(
        staging/src/k8s.io/code-generator/examples
        staging/src/k8s.io/kube-aggregator
        staging/src/k8s.io/sample-apiserver
        staging/src/k8s.io/sample-controller
        staging/src/k8s.io/metrics
        staging/src/k8s.io/apiextensions-apiserver
        staging/src/k8s.io/apiextensions-apiserver/examples/client-go
    )

    local codegen
    codegen="${ONEX_ROOT}/staging/src/k8s.io/code-generator"
    for sub in "${subs[@]}"; do
        onex::log::status "Generating code for subproject ${sub}"
        pushd "${sub}" >/dev/null
        CODEGEN_PKG="${codegen}" \
        UPDATE_API_KNOWN_VIOLATIONS="${UPDATE_API_KNOWN_VIOLATIONS}" \
        API_KNOWN_VIOLATIONS_DIR="${API_KNOWN_VIOLATIONS_DIR}" \
            ./scripts/update-codegen.sh > >(indent) 2> >(indent >&2)
        popd >/dev/null
    done
}

function unused::codegen::protobindings() {
    # Each element of this array is a directory containing subdirectories which
    # eventually contain a file named "api.proto".
    local apis=(
        "staging/src/k8s.io/cri-api/pkg/apis/runtime"

        "staging/src/k8s.io/kubelet/pkg/apis/podresources"

        "staging/src/k8s.io/kubelet/pkg/apis/deviceplugin"

        "staging/src/k8s.io/kms/apis"
        "staging/src/k8s.io/apiserver/pkg/storage/value/encrypt/envelope/kmsv2"

        "staging/src/k8s.io/kubelet/pkg/apis/dra"

        "staging/src/k8s.io/kubelet/pkg/apis/pluginregistration"
        "pkg/kubelet/pluginmanager/pluginwatcher/example_plugin_apis"
    )

    onex::log::status "Generating protobuf bindings for ${#apis[@]} targets"
    if [[ "${DBG_CODEGEN}" == 1 ]]; then
        onex::log::status "DBG: generating protobuf bindings for:"
        for dir in "${apis[@]}"; do
            onex::log::status "DBG:     $dir"
        done
    fi

    for api in "${apis[@]}"; do
        git_find -z ":(glob)${api}"/'**/api.pb.go' \
            | xargs -0 rm -f
    done

    if onex::protoc::check_protoc >/dev/null; then
      scripts/_update-generated-proto-bindings-dockerized.sh "${apis[@]}"
    else
      onex::log::status "protoc ${PROTOC_VERSION} not found (can install with scripts/install-protoc.sh); generating containerized..."
      # NOTE: All output from this script needs to be copied back to the calling
      # source tree.  This is managed in onex::build::copy_output in build/common.sh.
      # If the output set is changed update that function.
      build/run.sh scripts/_update-generated-proto-bindings-dockerized.sh "${apis[@]}"
    fi
}

#
# main
#

function list_codegens() {
    (
        shopt -s extdebug
        declare -F \
            | cut -f3 -d' ' \
            | grep "^codegen::" \
            | while read -r fn; do declare -F "$fn"; done \
            | sort -n -k2 \
            | cut -f1 -d' ' \
            | sed 's/^codegen:://'
    )
}

# shellcheck disable=SC2207 # safe, no functions have spaces
all_codegens=($(list_codegens))

function print_codegens() {
    echo "available codegens:"
    for g in "${all_codegens[@]}"; do
        echo "    $g"
    done
}

# Validate and accumulate flags to pass thru and codegens to run if args are
# specified.
flags_to_pass=()
codegens_to_run=()
for arg; do
    # Use -? to list known codegens.
    if [[ "${arg}" == "-?" ]]; then
        print_codegens
        exit 0
    fi
    if [[ "${arg}" =~ ^- ]]; then
        flags_to_pass+=("${arg}")
        continue
    fi
    # Make sure each non-flag arg matches at least one codegen.
    nmatches=0
    for t in "${all_codegens[@]}"; do
        if [[ "$t" =~ ${arg} ]]; then
            nmatches=$((nmatches+1))
            # Don't run codegens twice, just keep the first match.
            # shellcheck disable=SC2076 # we want literal matching
            if [[ " ${codegens_to_run[*]} " =~ " $t " ]]; then
                continue
            fi
            codegens_to_run+=("$t")
            continue
        fi
    done
    if [[ ${nmatches} == 0 ]]; then
        echo "ERROR: no codegens match pattern '${arg}'"
        echo
        print_codegens
        exit 1
    fi
    # The array-syntax abomination is to accommodate older bash.
    codegens_to_run+=("${matches[@]:+"${matches[@]}"}")
done

# If no codegens were specified, run them all.
if [[ "${#codegens_to_run[@]}" == 0 ]]; then
    codegens_to_run=("${all_codegens[@]}")
fi

for g in "${codegens_to_run[@]}"; do
    # The array-syntax abomination is to accommodate older bash.
    "codegen::${g}" "${flags_to_pass[@]:+"${flags_to_pass[@]}"}"
done
