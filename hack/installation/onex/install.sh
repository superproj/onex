#!/bin/bash

# Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
# Use of this source code is governed by a MIT style
# license that can be found in the LICENSE file. The original repo for
# this file is https://github.com/superproj/onex.
#

# The root of the build/dist directory
ONEX_ROOT=$(dirname "${BASH_SOURCE[0]}")/../..
# If common.sh has already been sourced, it will not be sourced again here.
[[ -z ${COMMON_SOURCED} ]] && source ${ONEX_ROOT}/hack/installation/common.sh
# Set some environment variables.
INSTALL_DIR=${ONEX_ROOT}/hack/installation

source ${INSTALL_DIR}/jaeger.sh
source ${INSTALL_DIR}/kafka.sh
source ${INSTALL_DIR}/mariadb.sh
source ${INSTALL_DIR}/mongo.sh
source ${INSTALL_DIR}/redis.sh
source ${INSTALL_DIR}/etcd.sh
source ${INSTALL_DIR}/onex.sh
source ${INSTALL_DIR}/test.sh

# OneX 容器化快速安装
onex::install::docker::install()
{
  # 安装前置依赖
  onex::install::pre_install

  # 安装存储组件
  onex::install::storage::docker::install

  # 安装中间件
  onex::install::middleware::docker::install

  # 安装 OneX 各组件
  onex::onex::docker::install

  # 等待所有服务启动
  echo "Sleeping to wait for all onex container to complete startup ..."
  sleep 10
  # 安装后测试
  onex::test::test
}

# OneX 安装前的一些前置处理
onex::install::pre_install()
{
  # 为了确保 Debian 系统具有最新的安全补丁和软件，这里先更新系统软件包。
  # 如果你要做存量软件的版本升级，可以执行`sudo apt upgrade`，但不建议
  # `apt update` 命令会更新系统中的软件包版本
  onex::util::sudo "apt update"

  onex::util::sudo "apt install -y software-properties-common dirmngr apt-transport-https"

  # 配置 hosts
  if ! egrep -q 'onex.*.superproj.com' /etc/hosts; then
    echo ${LINUX_PASSWORD} | sudo -S cat << EOF | sudo tee -a /etc/hosts

# host configs for onex project
${ONEX_ACCESS_HOST} onex.usercenter.superproj.com
${ONEX_ACCESS_HOST} onex.gateway.superproj.com
${ONEX_ACCESS_HOST} onex.apiserver.superproj.com
${ONEX_ACCESS_HOST} onex.nightwatch.superproj.com
${ONEX_ACCESS_HOST} onex.pump.superproj.com
EOF
  fi
}

# OneX 幂等卸载，注意卸载顺序
onex::install::docker::uninstall()
{
  onex::onex::docker::uninstall || true
  onex::install::middleware::docker::uninstall || true
  onex::install::storage::docker::uninstall || true

  # 删除安装目录
  onex::util::sudo "rm -rf ${ONEX_INSTALL_DIR}"
  onex::util::sudo "rm -rf ${ONEX_THIRDPARTY_INSTALL_DIR}"

  # 卸载 onex 网络，始终返回成功
  docker network rm onex &>/dev/null || true
}

# 安装所有的存储组件
onex::install::storage::docker::install()
{
  onex::mariadb::docker::install
  onex::redis::docker::install
  onex::mongo::docker::install
  onex::etcd::docker::install
}

# 卸载所有的存储组件
onex::install::storage::docker::uninstall()
{
  onex::mariadb::docker::uninstall
  onex::redis::docker::uninstall
  onex::mongo::docker::uninstall
  onex::etcd::docker::uninstall
}

# 安装其他中间件
onex::install::middleware::docker::install()
{
  onex::jaeger::docker::install
  onex::kafka::docker::install
}

# 卸载其他中间件
onex::install::middleware::docker::uninstall()
{
  onex::jaeger::docker::uninstall
  onex::kafka::docker::uninstall
}

# OneX 脚本自动化安装
# 你可以直接阅读安装脚本，学习安装的所有细节
# 有些组件本地化手动安装会比较复杂，仍然会采用容器化安装，比如：kafka。
onex::install::sbs::install()
{
  onex::install::pre_install
  onex::install::storage::sbs::install
  onex::install::middleware::sbs::install
  onex::onex::sbs::install

  # 等待所有服务启动
  echo "Sleeping to wait for all onex container to complete startup ..."
  sleep 10
  onex::test::test
}

# OneX 幂等卸载，注意卸载顺序
onex::install::sbs::uninstall()
{
  onex::onex::sbs::uninstall || true
  onex::install::middleware::sbs::uninstall || true
  onex::install::storage::sbs::uninstall || true
  onex::util::sudo "rm -rf ${ONEX_INSTALL_DIR}"
}

# 安装所有的存储组件
onex::install::storage::sbs::install()
{
  onex::mariadb::sbs::install
  onex::redis::sbs::install
  onex::mongo::sbs::install
  onex::etcd::sbs::install
}

# 卸载所有的存储组件
onex::install::storage::sbs::uninstall()
{
  onex::mariadb::sbs::uninstall
  onex::redis::sbs::uninstall
  onex::mongo::sbs::uninstall
  onex::etcd::sbs::uninstall
}

# 安装其他中间件
# 因为 jaeger 和 kafka 手动安装比较复杂，这里仍然采用 docker 安装
onex::install::middleware::sbs::install()
{
  onex::jaeger::docker::install
  onex::kafka::docker::install
}

# 卸载其他中间件
onex::install::middleware::sbs::uninstall()
{
  onex::jaeger::sbs::uninstall
  onex::kafka::sbs::uninstall
}

if [[ "$*" =~ onex::install:: ]]; then
  eval $*
fi
