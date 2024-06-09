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

# This script checks that various type of documents(* .md, * .yaml and man
# files) are generated correctly(just verifies the generation process).
# Usage: `scripts/verify-generated-docs.sh`.

set -o errexit
set -o nounset
set -o pipefail

ONEX_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
source "${ONEX_ROOT}/scripts/lib/verify-generated.sh"

onex::verify::generated "Generated docs need to be updated" "Please run 'scripts/update-generated-docs.sh'" scripts/update-generated-docs.sh "$@"
