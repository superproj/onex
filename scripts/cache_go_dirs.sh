#!/usr/bin/env bash

# Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
# Use of this source code is governed by a MIT style
# license that can be found in the LICENSE file. The original repo for
# this file is https://github.com/superproj/onex.
#


# This script finds, caches, and prints a list of all directories that hold
# *.go files.  If any directory is newer than the cache, re-find everything and
# update the cache.  Otherwise use the cached file.

set -o errexit
set -o nounset
set -o pipefail

if [[ -z "${1:-}" ]]; then
    echo "usage: $0 <cache-file>"
    exit 1
fi
CACHE="$1"; shift

trap 'rm -f "${CACHE}"' HUP INT TERM ERR

# This is a partial 'find' command.  The caller is expected to pass the
# remaining arguments.
#
# Example:
#   kfind -type f -name foobar.go
function kfind() {
    # We want to include the "special" vendor directories which are actually
    # part of the Kubernetes source tree (./staging/*) but we need them to be
    # named as their ./vendor/* equivalents.  Also, we do not want all of
    # ./vendor nor ./hack/tools/vendor nor even all of ./vendor/k8s.io.
    find -H .                      \
        \(                         \
        -not \(                    \
            \(                     \
                -name '_*' -o      \
                -name '.[^.]*' -o  \
                \(                 \
                  -name 'vendor'   \
                  -type d          \
                \) -o              \
                \(                 \
                  -name 'testdata' \
                  -type d          \
                \)                 \
            \) -prune              \
        \)                         \
        \)                         \
        "$@"                       \
        | sed 's|^./staging/src|vendor|'
}

# It's *significantly* faster to check whether any directories are newer than
# the cache than to blindly rebuild it.
if [[ -f "${CACHE}" && -n "${CACHE}" ]]; then
    N=$(kfind -type d -newer "${CACHE}" -print -quit | wc -l)
    if [[ "${N}" == 0 ]]; then
        cat "${CACHE}"
        exit
    fi
fi

mkdir -p "$(dirname "${CACHE}")"
kfind -type f -name \*.go  \
    | sed 's|/[^/]*$||'    \
    | sed 's|^./||'        \
    | LC_ALL=C sort -u     \
    | tee "${CACHE}"
