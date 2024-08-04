#!/usr/bin/env bash

# Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
# Use of this source code is governed by a MIT style
# license that can be found in the LICENSE file. The original repo for
# this file is https://github.com/superproj/onex.
#


# The root of the build/dist directory
ONEX_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
source "${ONEX_ROOT}/hack/lib/init.sh"

# OUT_DIR can come in from the Makefile, so honor it.
readonly LOCAL_OUTPUT_ROOT="${ONEX_ROOT}/${OUT_DIR:-_output}"
readonly LOCAL_OUTPUT_CAPATH="${LOCAL_OUTPUT_ROOT}/cert"
readonly PROJ_DOMAIN="superproj.com"
# To ensure that access is available in the default k8s cluster (with cluster domain as cluster.local).
readonly DEFAULT_KUBERNETES_CLUSTER_DOMAIN="cluster.local"

# Hostname for the cert
#readonly CERT_HOSTNAME="${CERT_HOSTNAME:-onex-apiserver},127.0.0.1,localhost,"
readonly CERT_HOSTNAME="${CERT_HOSTNAME:-onex-apiserver},127.0.0.1,localhost,10.37.83.200"

# Run the cfssl command to generates certificate files for node service, the
# certificate files will save in $1 directory.
#
# Args:
#   $1 (the directory that certificate files to save)
#   $2 (the prefix of the certificate filename)
function generate-node-cert() {
  local cert_dir=${1}
  local prefix=${2:-}
  local expiry=${3:-876000h}


  mkdir -p "${cert_dir}"
  pushd "${cert_dir}" >/dev/null 2>&1

  onex::util::ensure-cfssl

  if [ ! -r "ca-config.json" ]; then
    cat >ca-config.json <<EOF
{
  "signing": {
    "default": {
      "expiry": "${expiry}"
    },
    "profiles": {
      "node": {
        "usages": [
          "signing",
          "key encipherment",
          "server auth",
          "client auth"
        ],
        "expiry": "${expiry}"
      }
  }
}
}
EOF
  fi

  if [ ! -r "ca-csr.json" ]; then
    cat >ca-csr.json <<EOF
{
  "CN": "onex",
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [
    {
      "C": "CN",
      "ST": "Shenzhen",
      "L": "Shenzhen",
      "O": "onex",
      "OU": "System"
    }
  ],
  "ca": {
    "expiry": "${expiry}"
  }
}
EOF
  fi

  if [[ ! -r "ca.pem" || ! -r "ca-key.pem" ]]; then
    ${CFSSL_BIN} gencert -initca ca-csr.json | ${CFSSLJSON_BIN} -bare ca -
  fi

  if [[ -z "${prefix}" ]];then
    return 0
  fi

  #echo "Generate "${prefix}" certificates..."
  echo '{"CN":"'"${prefix}"'","hosts":[],"key":{"algo":"rsa","size":2048},"names":[{"C":"CN","ST":"Shenzhen","L":"Shenzhen","O":"tencent","OU":"'"${prefix}"'"}]}' \
    | ${CFSSL_BIN} gencert -hostname="${CERT_HOSTNAME},${prefix/-/.}.${DEFAULT_KUBERNETES_CLUSTER_DOMAIN},${prefix/-/.}.${PROJ_DOMAIN}" -ca=ca.pem -ca-key=ca-key.pem \
    -config=ca-config.json -profile=node - | ${CFSSLJSON_BIN} -bare "${prefix}"

  # the popd will access `directory stack`, no `real` parameters is actually needed
  # shellcheck disable=SC2119
   popd >/dev/null 2>&1
}

$*
