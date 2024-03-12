#!/usr/bin/env bash

# Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
# Use of this source code is governed by a MIT style
# license that can be found in the LICENSE file. The original repo for
# this file is https://github.com/superproj/onex.
#


ONEX_ROOT=$(dirname "${BASH_SOURCE[0]}")/..

source "${ONEX_ROOT}/scripts/common.sh"

readonly LOCAL_OUTPUT_CONFIGPATH="${LOCAL_OUTPUT_ROOT}/configs"
mkdir -p ${LOCAL_OUTPUT_CONFIGPATH}

cd ${ONEX_ROOT}/scripts

export ONEX_APISERVER_INSECURE_BIND_ADDRESS=0.0.0.0
export ONEX_AUTHZ_SERVER_INSECURE_BIND_ADDRESS=0.0.0.0

# 集群内通过kubernetes服务名访问
export ONEX_APISERVER_HOST=miner-apiserver
export ONEX_AUTHZ_SERVER_HOST=miner-authz-server
export ONEX_PUMP_HOST=miner-pump
export ONEX_WATCHER_HOST=miner-watcher

# 配置CA证书路径
export CONFIG_USER_CLIENT_CERTIFICATE=/etc/miner/cert/admin.pem
export CONFIG_USER_CLIENT_KEY=/etc/miner/cert/admin-key.pem
export CONFIG_SERVER_CERTIFICATE_AUTHORITY=/etc/miner/cert/ca.pem

for comp in $(ls ${ONEX_ROOT/cmd})
do
  onex::log::info "generate ${LOCAL_OUTPUT_CONFIGPATH}/${comp}.yaml"
  ./gen-config.sh install/environment.sh ../configs/${comp}.yaml > ${LOCAL_OUTPUT_CONFIGPATH}/${comp}.yaml
done
