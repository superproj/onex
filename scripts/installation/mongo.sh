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
ONEX_MONGO_HOST=${ONEX_MONGO_HOST:-127.0.0.1}
ONEX_MONGO_PORT=${ONEX_MONGO_PORT:-27017}
ONEX_MONGO_URL=${ONEX_MONGO_HOST}:${ONEX_MONGO_PORT}
ONEX_MONGO_DATABASE=${ONEX_MONGO_DATABASE:-onex}
ONEX_MONGO_ADMIN_USERNAME=${ONEX_MONGO_ADMIN_USERNAME:-root}
ONEX_MONGO_ADMIN_PASSWORD=${ONEX_MONGO_ADMIN_PASSWORD:-'onex(#)666'}
#ONEX_MONGO_ADMIN_AUTH=${ONEX_MONGO_ADMIN_USERNAME}:"'${ONEX_MONGO_ADMIN_PASSWORD}'"
ONEX_MONGO_ADMIN_AUTH=${ONEX_MONGO_ADMIN_USERNAME}:${ONEX_MONGO_ADMIN_PASSWORD}

# Install mongo using containerization.
onex::mongo::docker::install()
{
  onex::mongo::pre_install

  onex::common::network
  docker run -d --name onex-mongo \
    --network onex \
    -v ${ONEX_THIRDPARTY_INSTALL_DIR}/mongo:/data \
    -p ${ONEX_ACCESS_HOST}:${ONEX_MONGO_PORT}:27017 \
    -e MONGO_INITDB_ROOT_USERNAME=${ONEX_MONGO_ADMIN_USERNAME} \
    -e MONGO_INITDB_ROOT_PASSWORD=${ONEX_MONGO_ADMIN_PASSWORD} \
    mongodb/mongodb-community-server:7.0.3-ubuntu2204

  sleep 10
  onex::mongo::status || return 1
  onex::mongo::info
  onex::log::info "install mongo successfully"
}


onex::mongo::pre_install()
{
  # 获取 MongoDB 公钥
  echo ${LINUX_PASSWORD} | sudo -S wget -qO - https://www.mongodb.org/static/pgp/server-7.0.asc | sudo apt-key add -

  # 添加 MongoDB APT 源
  echo ${LINUX_PASSWORD} | sudo -S echo "deb [arch=amd64,arm64] https://repo.mongodb.org/apt/debian $(lsb_release -cs)/mongodb-org/7.0 main" | sudo tee /etc/apt/sources.list.d/mongodb-org-7.0.list

  # 安装libssl1.1，否则安装 mongo 时会报以下错误：
  # mongodb-org-mongos : Depends: libssl1.1 (>= 1.1.1) but it is not installable
  wget http://archive.ubuntu.com/ubuntu/pool/main/o/openssl/libssl1.1_1.1.1f-1ubuntu2_amd64.deb -P /tmp/
  echo ${LINUX_PASSWORD} | sudo -S -i dpkg -i /tmp/libssl1.1_1.1.1f-1ubuntu2_amd64.deb

  onex::util::sudo "apt update"

  # 安装 MongoDB 客户端
  onex::util::sudo "apt install -y mongodb-mongosh"
}

# Uninstall the docker container.
onex::mongo::docker::uninstall()
{
  docker rm -f onex-mongo &>/dev/null
  onex::util::sudo "rm -rf ${ONEX_THIRDPARTY_INSTALL_DIR}/mongo"
  onex::log::info "uninstall mongo successfully"
}

# Install the mongo step by step.
# sbs is the abbreviation for "step by step".
onex::mongo::sbs::install()
{
  onex::mongo::pre_install

  echo ${LINUX_PASSWORD} | sudo -S apt install -y gnupg

  # 安装 MongoDB 服务端
  # 以为我们uninstall时会删除配置文件，所以要使用--force-confmiss 重新安装配置文件
  onex::util::sudo "apt -y -o Dpkg::Options::="--force-confmiss" --reinstall install mongodb-org mongodb-org-server"

  # 开启外网访问权限和登录验证
  echo ${LINUX_PASSWORD} | sudo -S sed -i '/bindIp/{s/127.0.0.1/0.0.0.0/}' /etc/mongod.conf
  # 关闭认证以创建 root 用户
  echo ${LINUX_PASSWORD} | sudo -S sed -i '/^#security/a\security:\n  authorization: disabled' /etc/mongod.conf

  # 启动 MongoDB，并设置开机启动
  onex::util::sudo "systemctl enable mongod"
  onex::util::sudo "systemctl restart mongod"
  echo "Sleeping 5s to wait for mongo to complete startup ..."
  sleep 5

  # 创建管理员账号，设置管理员密码
  echo ${LINUX_PASSWORD} | sudo -S mongosh --quiet "mongodb://${ONEX_MONGO_URL}" <<EOF
use admin
db.createUser({user:"${ONEX_MONGO_ADMIN_USERNAME}",pwd:"${ONEX_MONGO_ADMIN_PASSWORD}",roles:["root"]})
db.auth("${ONEX_MONGO_ADMIN_USERNAME}", "${ONEX_MONGO_ADMIN_PASSWORD}")
quit
EOF
  # 开启认证
  echo ${LINUX_PASSWORD} | sudo -S sed -i '/authorization:/s/disabled/enabled/g' /etc/mongod.conf

  onex::util::sudo "systemctl restart mongod"

  echo "Sleeping 5s to wait for mongo to complete startup ..."
  sleep 5

  onex::mongo::status || return 1
  onex::mongo::info
  onex::log::info "install mongo successfully"
}

# Uninstall the mongo step by step.
onex::mongo::sbs::uninstall()
{
  set +o errexit
  onex::util::sudo "systemctl stop mongodb"
  onex::util::sudo "systemctl disable mongodb"
  onex::util::sudo "apt remove -y mongodb-org mongodb-org-server" # 这里我们客户端不卸载
  onex::util::sudo "rm -rvf /var/lib/mongodb"
  onex::util::sudo "rm -vf /etc/apt/sources.list.d/mongodb-org-7.0.list"
  onex::util::sudo "rm -vf /etc/mongod.conf"
  onex::util::sudo "rm -vf /lib/systemd/system/mongod.service"
  onex::util::sudo "rm -vf /tmp/mongodb-*.sock"
  set -o errexit

  onex::log::info "uninstall mongo successfully"
}

# Print necessary information after docker or sbs installation.
onex::mongo::info()
{
  echo -e ${C_GREEN}mongo has been installed, here are some useful information:${C_NORMAL}
  encoded=$(echo -n "${ONEX_MONGO_ADMIN_PASSWORD}"|jq -sRr @uri)
  cat << EOF | sed 's/^/  /'
Mongo access url is: mongodb://${ONEX_MONGO_URL}
  Mongo admin username is: ${ONEX_MONGO_ADMIN_USERNAME}
  Mongo admin password is: ${ONEX_MONGO_ADMIN_PASSWORD}
    MongoDB Login Command: mongosh mongodb://${ONEX_MONGO_ADMIN_USERNAME}:'${encoded}'@${ONEX_MONGO_URL}
EOF
}

# Status check after docker or sbs installation.
onex::mongo::status()
{
  onex::util::telnet ${ONEX_MONGO_HOST} ${ONEX_MONGO_PORT} || return 1
}

if [[ "$*" =~ onex::mongo:: ]]; then
  eval $*
fi
