#!/bin/bash
#
# Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
# Use of this source code is governed by a MIT style
# license that can be found in the LICENSE file. The original repo for
# this file is https://github.com/superproj/onex.
#

# Common utilities, variables and checks for all build scripts.
set -eEuo pipefail

# The root of the build/dist directory
ONEX_ROOT=$(dirname "${BASH_SOURCE[0]}")/../..

# 所有的 OneX 都会统一加载 scripts/common.sh 脚本
source "${ONEX_ROOT}/scripts/common.sh"

# 设置 ONEX_ENV_FILE（重要）
ONEX_ENV_FILE=${ONEX_ENV_FILE:-${ONEX_ROOT}/manifests/env.local}
# 加载本地安装环境变量（非常重要的一步，后面很多步骤都依赖于env.local中的变量设置）
source ${ONEX_ENV_FILE}

COMMON_SOURCED=true # Sourced flag

# 设置本地/容器化安装的环境变量，主要是为了避免端口冲突
export ONEX_CONTROLLER_MANAGER_METRICS_PORT=59081
export ONEX_CONTROLLER_MANAGER_HEALTHZ_PORT=59082
export ONEX_MINERSET_CONTROLLER_METRICS_PORT=60081
export ONEX_MINERSET_CONTROLLER_HEALTHZ_PORT=60082
export ONEX_MINER_CONTROLLER_METRICS_PORT=61081
export ONEX_MINER_CONTROLLER_HEALTHZ_PORT=61082

# 确保 onex 容器网络存在。
# 在 uninstall 时，可不删除 onex 容器网络，可以作为一个无害的无用数据
onex::common::network()
{
  docker network ls |grep -q onex || docker network create onex
}
