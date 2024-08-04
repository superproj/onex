#!/usr/bin/env bash

# Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
# Use of this source code is governed by a MIT style
# license that can be found in the LICENSE file. The original repo for
# this file is https://github.com/superproj/onex.
#

set -o errexit
set +o nounset
set -o pipefail

# Short-circuit if init.sh has already been sourced
[[ $(type -t onex::init::loaded) == function ]] && return 0

# Unset CDPATH so that path interpolation can work correctly
# https://github.com/minerrnetes/minerrnetes/issues/52255
unset CDPATH

# Default use go modules
export GO111MODULE=on

# The root of the build/dist directory
ONEX_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"

ONEX_OUTPUT_SUBPATH="${ONEX_OUTPUT_SUBPATH:-_output}"
ONEX_OUTPUT="${ONEX_ROOT}/${ONEX_OUTPUT_SUBPATH}"

source "${ONEX_ROOT}/hack/lib/util.sh"
source "${ONEX_ROOT}/hack/lib/logging.sh"
source "${ONEX_ROOT}/hack/lib/color.sh"

onex::log::install_errexit
onex::util::ensure-bash-version

source "${ONEX_ROOT}/hack/lib/version.sh"
source "${ONEX_ROOT}/hack/lib/golang.sh"

# list of all available group versions. This should be used when generated code
# or when starting an API server that you want to have everything.
# most preferred version for a group should appear first
# UPDATEME: New group need to update here.
ONEX_AVAILABLE_GROUP_VERSIONS="${ONEX_AVAILABLE_GROUP_VERSIONS:-\
apps/v1beta1 \
}"

# This emulates "readlink -f" which is not available on MacOS X.
# Test:
# T=/tmp/$$.$RANDOM
# mkdir $T
# touch $T/file
# mkdir $T/dir
# ln -s $T/file $T/linkfile
# ln -s $T/dir $T/linkdir
# function testone() {
#   X=$(readlink -f $1 2>&1)
#   Y=$(onex::readlinkdashf $1 2>&1)
#   if [ "$X" != "$Y" ]; then
#     echo readlinkdashf $1: expected "$X", got "$Y"
#   fi
# }
# testone /
# testone /tmp
# testone $T
# testone $T/file
# testone $T/dir
# testone $T/linkfile
# testone $T/linkdir
# testone $T/nonexistant
# testone $T/linkdir/file
# testone $T/linkdir/dir
# testone $T/linkdir/linkfile
# testone $T/linkdir/linkdir
function onex::readlinkdashf {
  # run in a subshell for simpler 'cd'
  (
    if [[ -d "${1}" ]]; then # This also catch symlinks to dirs.
      cd "${1}"
      pwd -P
    else
      cd "$(dirname "${1}")"
      local f
      f=$(basename "${1}")
      if [[ -L "${f}" ]]; then
        readlink "${f}"
      else
        echo "$(pwd -P)/${f}"
      fi
    fi
  )
}

# This emulates "realpath" which is not available on MacOS X
# Test:
# T=/tmp/$$.$RANDOM
# mkdir $T
# touch $T/file
# mkdir $T/dir
# ln -s $T/file $T/linkfile
# ln -s $T/dir $T/linkdir
# function testone() {
#   X=$(realpath $1 2>&1)
#   Y=$(onex::realpath $1 2>&1)
#   if [ "$X" != "$Y" ]; then
#     echo realpath $1: expected "$X", got "$Y"
#   fi
# }
# testone /
# testone /tmp
# testone $T
# testone $T/file
# testone $T/dir
# testone $T/linkfile
# testone $T/linkdir
# testone $T/nonexistant
# testone $T/linkdir/file
# testone $T/linkdir/dir
# testone $T/linkdir/linkfile
# testone $T/linkdir/linkdir
onex::realpath() {
  if [[ ! -e "${1}" ]]; then
    echo "${1}: No such file or directory" >&2
    return 1
  fi
  onex::readlinkdashf "${1}"
}

# Marker function to indicate init.sh has been fully sourced
onex::init::loaded() {
  return 0
}
