#!/usr/bin/env bash

# Copyright 2015 The Kubernetes Authors.
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

# This script genertates `*/api.pb.go` from the protobuf file `*/api.proto`.
# Usage:
#     scripts/update-generated-protobuf-dockerized.sh "${APIROOTS[@]}"
#     An example APIROOT is: "k8s.io/api/admissionregistration/v1"

set -o errexit
set -o nounset
set -o pipefail

ONEX_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
source "${ONEX_ROOT}/scripts/lib/init.sh"
source "${ONEX_ROOT}/scripts/lib/protoc.sh"

onex::protoc::check_protoc
onex::golang::setup_env

GOPROXY=off go install k8s.io/code-generator/cmd/go-to-protobuf
GOPROXY=off go install k8s.io/code-generator/cmd/go-to-protobuf/protoc-gen-gogo

# requires the 'proto' tag to build (will remove when ready)
# searches for the protoc-gen-gogo extension in the output directory
# satisfies import of github.com/gogo/protobuf/gogoproto/gogo.proto and the
# core Google protobuf types
PATH="${ONEX_ROOT}/_output/bin:${PATH}" \
  go-to-protobuf \
  -v "${KUBE_VERBOSE}" \
  --go-header-file "${ONEX_ROOT}/scripts/boilerplate/boilerplate.generatego.txt" \
  --proto-import="${ONEX_ROOT}/third_party/protobuf" \
  --proto-import="${GOPATH}/src" \
  --output-dir="${GOPATH}/src" \
  --packages="$(IFS=, ; echo "$*")"
