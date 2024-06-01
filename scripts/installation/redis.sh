#!/bin/bash

# Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
# Use of this source code is governed by a MIT style
# license that can be found in the LICENSE file. The original repo for
# this file is https://github.com/superproj/onex.
#

# The root of the build/dist directory.
ONEX_ROOT=$(dirname "${BASH_SOURCE[0]}")/../..
# If common.sh has already been sourced, it will not be sourced again here.
[[ -z ${COMMON_SOURCED} ]] && source ${ONEX_ROOT}/scripts/installation/common.sh
# Set some environment variables.
ONEX_REDIS_HOST=${ONEX_REDIS_HOST:-127.0.0.1}
ONEX_REDIS_PORT=${ONEX_REDIS_PORT:-6379}
ONEX_PASSWORD=${ONEX_PASSWORD:-onex(#)666}

# Install redis using containerization.
onex::redis::docker::install()
{
  onex::redis::pre_install

  onex::common::network
  docker run -d --name onex-redis \
    --restart always \
    --network onex \
    -v ${ONEX_THIRDPARTY_INSTALL_DIR}/redis:/data \
    -p ${ONEX_ACCESS_HOST}:${ONEX_REDIS_PORT}:6379 \
    redis:7.2.3 \
    redis-server \
    --appendonly yes \
    --save 60 1 \
    --protected-mode no \
    --requirepass ${ONEX_PASSWORD} \
    --loglevel debug

  sleep 2
  onex::redis::status || return 1
  onex::redis::info
  onex::log::info "install redis successfully"
}

# Uninstall the docker container.
onex::redis::docker::uninstall()
{
  docker rm -f onex-redis &>/dev/null
  onex::util::sudo "rm -rf ${ONEX_THIRDPARTY_INSTALL_DIR}/redis"
  onex::log::info "uninstall redis successfully"
}

# Install the redis step by step.
# sbs is the abbreviation for "step by step".
onex::redis::sbs::install()
{
  onex::redis::pre_install

  # 创建 `/var/lib/redis` 目录，否则 `redis-server` 命令启动时
  # 会报：`Can't chdir to '/var/lib/redis': No such file or directory` 错误
  onex::util::sudo "mkdir -p /var/lib/redis"

  # 安装 Redis
  onex::util::sudo "apt install -y -o Dpkg::Options::="--force-confmiss" --reinstall redis-server"

  # 配置 Redis
  # 修改 `/etc/redis/redis.conf` 文件，将 daemonize 由 no 改成 yes，表示允许 Redis 在后台启动
  redis_conf=/etc/redis/redis.conf
  # 注意：有的系统 redis 配置文件路径为 `/etc/redis.conf`
  [[ -f /etc/redis.conf ]] && redis_conf=/etc/redis.conf

  echo ${LINUX_PASSWORD} | sudo -S sed -i '/^daemonize/{s/no/yes/}' ${redis_conf}

  # 修改 Redis 端口为 ${ONEX_REDIS_PORT}
  echo ${LINUX_PASSWORD} | sudo -S sed -i "s/^port.*/port ${ONEX_REDIS_PORT}/g" ${redis_conf}

  # 在 `bind 127.0.0.1` 前面添加 `#` 将其注释掉，默认情况下只允许本地连接，注释掉后外网可以连接 Redis
  echo ${LINUX_PASSWORD} | sudo -S sed -i '/^bind .*127.0.0.1/s/^/# /' ${redis_conf}

  # 修改 requirepass 配置，设置 Redis 密码
  echo ${LINUX_PASSWORD} | sudo -S sed -i 's/^# requirepass.*$/requirepass '"${ONEX_REDIS_PASSWORD}"'/' ${redis_conf}

  # 因为我们上面配置了密码登录，需要将 protected-mode 设置为 no，关闭保护模式
  echo ${LINUX_PASSWORD} | sudo -S sed -i '/^protected-mode/{s/yes/no/}' ${redis_conf}

  # 为了能够远程连上 Redis，需要执行以下命令关闭防火墙，并禁止防火墙开机启动（如果不需要远程连接，可忽略此步骤）
  set +o errexit
  #onex::util::sudo "systemctl stop firewalld.service"
  #onex::util::sudo "systemctl disable firewalld.service"
  set -o errexit

  # 重启 Redis
  #onex::util::sudo "redis-server ${redis_conf}"
  onex::util::sudo "systemctl restart redis-server"

  onex::redis::status || return 1
  onex::redis::info
  onex::log::info "install redis successfully"
}

onex::redis::pre_install()
{
  onex::util::sudo "apt install -y redis-tools"
}

# Uninstall the redis step by step.
onex::redis::sbs::uninstall()
{
  # 先删除 redis-server 进程，否则 `systemctl stop redis-server` 可能会卡主
  redis_pid=$(pgrep -f redis-server)
  [[ ${redis_pid} != "" ]] && onex::util::sudo "kill -9 ${redis_pid}"

  set +o errexit
  onex::util::sudo "systemctl stop redis-server"
  onex::util::sudo "systemctl disable redis-server"
  onex::util::sudo "apt remove -y redis-server"
  onex::util::sudo "rm -rf /var/lib/redis"
  set -o errexit
  onex::log::info "uninstall redis successfully"
}

# Print necessary information after docker or sbs installation.
onex::redis::info()
{
  echo -e ${C_GREEN}redis has been installed, here are some useful information:${C_NORMAL}
  cat << EOF | sed 's/^/  /'
Redis access endpoint is: ${ONEX_REDIS_HOST}:${ONEX_REDIS_PORT}
       Redis password is: ${ONEX_PASSWORD}
     Redis Login Command: redis-cli --no-auth-warning -h ${ONEX_REDIS_HOST} -p ${ONEX_REDIS_PORT} -a '${ONEX_REDIS_PASSWORD}'
EOF
}

# Status check after docker or sbs installation.
onex::redis::status()
{
  onex::util::telnet ${ONEX_REDIS_HOST} ${ONEX_REDIS_PORT} || return 1
  redis-cli --no-auth-warning -h ${ONEX_REDIS_HOST} -p ${ONEX_REDIS_PORT} -a "${ONEX_REDIS_PASSWORD}" --hotkeys || {
    onex::log::error "can not login with ${ONEX_REDIS_USERNAME}, redis maybe not initialized properly."
    return 1
  }
}

if [[ "$*" =~ onex::redis:: ]]; then
  eval $*
fi
