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
ONEX_MYSQL_HOST=${ONEX_MYSQL_HOST:-127.0.0.1}
ONEX_MYSQL_PORT=${ONEX_MYSQL_PORT:-3306}
ONEX_PASSWORD=${ONEX_PASSWORD:-onex(#)666}

# Install mariadb using containerization.
onex::mariadb::docker::install()
{
  # 安装客户端工具，访问 MariaDB
  onex::util::sudo "apt install -y mariadb-client"

  onex::common::network
  docker run -d --name onex-mariadb \
    --network onex \
    -v ${ONEX_THIRDPARTY_INSTALL_DIR}/mariadb:/var/lib/mysql \
    -p ${ONEX_ACCESS_HOST}:${ONEX_MYSQL_PORT}:3306 \
    -e MYSQL_ROOT_PASSWORD=${ONEX_PASSWORD} \
    mariadb:11.2.2

  echo "Sleeping to wait for all onex-mariadb container to complete startup ..."
  sleep 10
  onex::mariadb::status || return 1

  onex::mariadb::info
  onex::log::info "install mariadb successfully"
}

# Uninstall the docker container.
onex::mariadb::docker::uninstall()
{
  docker rm -f onex-mariadb &>/dev/null
  onex::util::sudo "rm -rf ${ONEX_THIRDPARTY_INSTALL_DIR}/mariadb"
  onex::log::info "uninstall mariadb successfully"
}

# Install the mariadb step by step.
# sbs is the abbreviation for "step by step".
onex::mariadb::sbs::install()
{
  # 本机 apt 安装后 MySQL 端口固定位 3306
  export ONEX_MYSQL_PORT=3306

  # 从指定的 URL中获取 MariaDB 的发布密钥。这个密钥用于验证 MariaDB
  # 软件包的签名，确保软件包在下载和安装过程中的完整性和安全性
  echo ${LINUX_PASSWORD} | sudo -S apt-key adv --fetch-keys 'https://mariadb.org/mariadb_release_signing_key.asc'
  # 配置 MariaDB 11.2.2 apt 源（docker install 和 sbs install 版本都要保持一致）
  echo ${LINUX_PASSWORD} | sudo -S echo "deb [arch=amd64,arm64] https://mirrors.aliyun.com/mariadb/repo/11.2.2/debian/ $(lsb_release -cs) main" | sudo tee /etc/apt/sources.list.d/mariadb-11.2.2.list

  # 注意：一定要执行 `apt update`，否则可能安装的还是旧的软件包
  onex::util::sudo "apt update"

  # 需要先创建 /var/lib/mysql/ 目录，否则 `systemctl start mariadb` 时可能会报错
  onex::util::sudo "mkdir -p /var/lib/mysql"

  # 执行以下命令，防止uninstall后，出现：`update-alternatives: error: alternative path /etc/mysql/mariadb.cnf doesn't exist` 错误
  # 安装 MariaDB 客户端和 MariaDB 服务端
  onex::util::sudo "apt install -y -o Dpkg::Options::="--force-confmiss" --reinstall mariadb-client mariadb-server"

  # 启动 MariaDB，并设置开机启动
  onex::util::sudo "systemctl enable mariadb"

  # 为了方便你访问 MySQL，这里我们设置 MySQL 允许从所有机器网卡访问
  echo ${LINUX_PASSWORD} | sudo -S sed -i 's/^bind-address.*/bind-address = 0.0.0.0/g' /etc/mysql/mariadb.conf.d/50-server.cnf

  onex::util::sudo "systemctl restart mariadb"

  #  设置 root 初始密码
  onex::util::sudo "mysqladmin -u${ONEX_MYSQL_ADMIN_USERNAME} password ${ONEX_MYSQL_ADMIN_PASSWORD}"

  onex::mariadb::status || return 1
  onex::mariadb::info
  onex::log::info "install mariadb successfully"
}

# Uninstall the mariadb step by step.
onex::mariadb::sbs::uninstall()
{
  # `|| true` 实现幂等
  onex::util::sudo "systemctl stop mariadb" || true
  onex::util::sudo "systemctl disable mariadb" || true
  onex::util::sudo "apt remove -y mariadb-client mariadb-server" || true

  # 删除配置文件和数据目录，以及其他关联安装文件
  onex::util::sudo "rm -rvf /var/lib/mysql"
  onex::util::sudo "rm -rvf /etc/mysql"
  onex::util::sudo "rm -rvf /usr/share/keyrings/mariadb.gpg"
  onex::util::sudo "rm -vf /etc/apt/sources.list.d/mariadb-11.2.2.list"
  onex::log::info "uninstall mariadb successfully"
}

# Print necessary information after docker or sbs installation.
onex::mariadb::info()
{
  onex::color::green "mariadb has been installed, here are some useful information:"
  cat << EOF | sed 's/^/  /'
MySQL access endpoint is: ${ONEX_MYSQL_HOST}:${ONEX_MYSQL_PORT}
        root password is: ${ONEX_PASSWORD}
# `mysql` will be deprecated in the future, so here use `mariadb` instead.
Access command: mariadb -h ${ONEX_MYSQL_HOST} -P ${ONEX_MYSQL_PORT} -u root -p'${ONEX_PASSWORD}'
EOF
}

# Status check after docker or sbs installation.
onex::mariadb::status()
{
  sleep 20
  # 基础检查：检查端口，基础检查
  onex::util::telnet ${ONEX_MYSQL_HOST} ${ONEX_MYSQL_PORT} || return 1

  # 终态检查：检查 MySQL 是否成功运行
  echo mariadb -h${ONEX_MYSQL_HOST} -P${ONEX_MYSQL_PORT} -u${ONEX_MYSQL_ADMIN_USERNAME} -p${ONEX_MYSQL_ADMIN_PASSWORD} -e quit
  mariadb -h${ONEX_MYSQL_HOST} -P${ONEX_MYSQL_PORT} -u${ONEX_MYSQL_ADMIN_USERNAME} -p${ONEX_MYSQL_ADMIN_PASSWORD} -e quit &>/dev/null || {
    onex::log::error "can not login with root, mariadb maybe not initialized properly."
    return 1
  }
}

if [[ "$*" =~ onex::mariadb:: ]]; then
  eval $*
fi
