#!/bin/bash

# Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
# Use of this source code is governed by a MIT style
# license that can be found in the LICENSE file. The original repo for
# this file is https://github.com/superproj/onex.
#


# The script file is used to restart a specific component.
# Usage: `scripts/local-up-cluster/restart-component.sh kube-scheduler`

set -o errexit
set -o nounset
set -o pipefail

ONEX_ROOT=$(dirname "${BASH_SOURCE[0]}")/../..
source "${ONEX_ROOT}/scripts/lib/init.sh"

KUBE_ROOT=${KUBE_ROOT:-$GOPATH/src/k8s.io/kubernetes}
# 存放 Kuberentes 组件启动命令
ONEX_COMMAND_LINE_DIR=${ONEX_COMMAND_LINE_DIR:-${ONEX_OUTPUT}/k8s-cmdlines}

command="$1"

if [[ "${command}" == "" ]];then
  onex::log::error "Cannot run scripts/local-up-cluster/restart-component.sh with empty component name"
  exit 1
fi

# 获取进程ID和命令行参数
set +o errexit
process_info=$(ps -eo pid,cmd|grep kubernetes/_output/bin/${command}|egrep -v 'sudo|grep')
set -o errexit

# 从进程信息中提取进程ID
pid=$(echo "$process_info" | awk '{print $1}')

# 如果 ${pid} 不为空，说明进程存在，需要先 kill 掉
if [[ "${pid}" != "" ]];then
  onex::util::sudo "kill -9 ${pid}"
fi

# 重新运行命令
onex::util::sudo ${ONEX_COMMAND_LINE_DIR}/${command}
