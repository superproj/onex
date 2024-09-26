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
onex::util::ensure-temp-dir

BINS=(
  gen-docs
  gen-onex-docs
  gen-man
  gen-yaml
)
make build -C "${ONEX_ROOT}" BINS="${BINS[*]}"

# Run all known doc generators (today gendocs and genman for nodectl)
# $1 is the directory to put those generated documents
function generate_docs() {
  local dest="$1"

  # Find binary
  gendocs=$(onex::util::find-binary "gen-docs")
  genonexdocs=$(onex::util::find-binary "gen-onex-docs")
  genman=$(onex::util::find-binary "gen-man")
  genyaml=$(onex::util::find-binary "gen-yaml")

  mkdir -p "${dest}/docs/guide/en-US/cmd/onexctl"
  "${gendocs}" "${dest}/docs/guide/en-US/cmd/onexctl/"

  mkdir -p "${dest}/docs/guide/en-US/cmd"
  "${genonexdocs}" "${dest}/docs/guide/en-US/cmd/" "onex-fakeserver"
  "${genonexdocs}" "${dest}/docs/guide/en-US/cmd/" "onex-usercenter"
  "${genonexdocs}" "${dest}/docs/guide/en-US/cmd/" "onex-apiserver"
  "${genonexdocs}" "${dest}/docs/guide/en-US/cmd/" "onex-gateway"
  "${genonexdocs}" "${dest}/docs/guide/en-US/cmd/" "onex-nightwatch"
  "${genonexdocs}" "${dest}/docs/guide/en-US/cmd/" "onex-pump"
  "${genonexdocs}" "${dest}/docs/guide/en-US/cmd/" "onex-toyblc"
  "${genonexdocs}" "${dest}/docs/guide/en-US/cmd/" "onex-controller-manager"
  "${genonexdocs}" "${dest}/docs/guide/en-US/cmd/" "onex-minerset-controller"
  "${genonexdocs}" "${dest}/docs/guide/en-US/cmd/" "onex-miner-controller"
  "${genonexdocs}" "${dest}/docs/guide/en-US/cmd/onexctl" "onexctl"

  mkdir -p "${dest}/docs/man/man1/"
  "${genman}" "${dest}/docs/man/man1/" "onex-fakeserver"
  "${genman}" "${dest}/docs/man/man1/" "onex-usercenter"
  "${genman}" "${dest}/docs/man/man1/" "onex-apiserver"
  "${genman}" "${dest}/docs/man/man1/" "onex-gateway"
  "${genman}" "${dest}/docs/man/man1/" "onex-nightwatch"
  "${genman}" "${dest}/docs/man/man1/" "onex-pump"
  "${genman}" "${dest}/docs/man/man1/" "onex-toyblc"
  "${genman}" "${dest}/docs/man/man1/" "onex-controller-manager"
  "${genman}" "${dest}/docs/man/man1/" "onex-minerset-controller"
  "${genman}" "${dest}/docs/man/man1/" "onex-miner-controller"
  "${genman}" "${dest}/docs/man/man1/" "onexctl"

  mkdir -p "${dest}/docs/guide/en-US/yaml/onexctl/"
  "${genyaml}" "${dest}/docs/guide/en-US/yaml/onexctl/"

  # create the list of generated files
  pushd "${dest}" > /dev/null || return 1
  touch docs/.generated_docs
  find . -type f | cut -sd / -f 2- | LC_ALL=C sort > docs/.generated_docs
  popd > /dev/null || return 1
}

# Removes previously generated docs-- we don't want to check them in. $ONEX_ROOT
# must be set.
function remove_generated_docs() {
  if [ -e "${ONEX_ROOT}/docs/.generated_docs" ]; then
    # remove all of the old docs; we don't want to check them in.
    while read -r file; do
      rm "${ONEX_ROOT}/${file}" 2>/dev/null || true
    done <"${ONEX_ROOT}/docs/.generated_docs"
    # The docs/.generated_docs file lists itself, so we don't need to explicitly
    # delete it.
  fi
}

# generate into ONEX_TMP
generate_docs "${ONEX_TEMP}"

# remove all of the existing docs in ONEX_ROOT
remove_generated_docs

# Copy fresh docs into the repo.
# the shopt is so that we get docs/.generated_docs from the glob.
shopt -s dotglob
cp -af "${ONEX_TEMP}"/* "${ONEX_ROOT}"
shopt -u dotglob
