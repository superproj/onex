#!/bin/bash

# Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
# Use of this source code is governed by a MIT style
# license that can be found in the LICENSE file. The original repo for
# this file is https://github.com/superproj/onex.
#


# This script builds and runs a local kubernetes cluster. You may need to run
# this as root to allow kubelet to open docker's socket, and to write the test
# CA in /var/run/kubernetes.
# Usage: `scripts/local-up-cluster/start-local-cluster.sh`

set -o errexit
set -o nounset
set -o pipefail

ONEX_ROOT=$(dirname "${BASH_SOURCE[0]}")/../..
source "${ONEX_ROOT}/scripts/lib/init.sh"

KUBE_ROOT=${KUBE_ROOT:-$GOPATH/src/k8s.io/kubernetes}
# 存放 Kuberentes 组件启动命令
ONEX_COMMAND_LINE_DIR=${ONEX_COMMAND_LINE_DIR:-${ONEX_OUTPUT}/k8s-cmdlines}

cd ${KUBE_ROOT}

export CONTAINER_RUNTIME_ENDPOINT="unix:///run/containerd/containerd.sock"
export START_MODE=
# To avoid conflicts with the etcd port of the kind cluster, we are
# reassigning the conflicting port here.
export ETCD_PORT=2479
export ETCD_LISTEN_PEER_URLS=http://localhost:2481

function save_command_line() {
  local command="$1"
  local startup_wait_time=180
  local interval_time=2
  local find_process="ps -eo pid,cmd|grep kubernetes/_output/bin/${command}|egrep -qv 'sudo|grep'"

  set +o errexit
  onex::util::wait_for_success "${startup_wait_time}" "${interval_time}" "${find_process}"
  set -o errexit

  process_info=$(ps -eo pid,cmd|grep kubernetes/_output/bin/${command}|egrep -v 'sudo|grep')

  # 从进程信息中提取进程ID
  pid=$(echo "$process_info" | awk '{print $1}')

  #command_line=$(echo "$process_info" | awk '{$1=""; print $0}')

  if [[ ! -d ${ONEX_COMMAND_LINE_DIR} ]];then
    mkdir -p ${ONEX_COMMAND_LINE_DIR}
  fi

  < /proc/${pid}/cmdline xargs -0 printf ' %s' | sed -E "s/(--[^= ]+=)([^ ]+)/\1'\2'/g" | sed 's/ //' > ${ONEX_COMMAND_LINE_DIR}/${command}
  chmod +x ${ONEX_COMMAND_LINE_DIR}/${command}
}

function try_to_save_command_lines() {
  # save component start command lines
  for command in kube-apiserver kube-controller-manager kube-scheduler kubelet kube-proxy
  do
    save_command_line ${command} >/dev/null 2>&1 &
  done
}

# Save kubernetes startup command lines for debug purpose.
try_to_save_command_lines

# `local-up-cluster.sh` will copy and read files to or from
# the `/` directory, and root privileges are used here
echo "${LINUX_PASSWORD}" | sudo -SE ./hack/local-up-cluster.sh -O
