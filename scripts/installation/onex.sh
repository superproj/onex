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

source $(dirname "${BASH_SOURCE[0]}")/man.sh

# Install onex using containerization.
onex::onex::docker::install()
{
  onex::onex::prepare

  if [[ "${INSTALL_WITH_FRESH_IMAGE}" -eq 1 ]];then
    make -C ${ONEX_ROOT} image IMAGES=onex-allinone VERSION=${ONEX_IMAGE_VERSION}
  fi

  # 创建一个数据卷，将 onex 容器中的安装目录挂载到宿主机上
  onex::util::sudo "mkdir -p ${ONEX_INSTALL_DIR}"
  # docker 命名卷挂在规则：
  #   1. 如果host目录为空（目录可以存在，但里面不能有文件），容器目录不为空，
  #      则将容器目录内容先复制到host目录，再以host目录为主
  #   2. 如果host目录不为空，则以host目录位置
  # 为了将 onex 容器中 /opt/onex 目录内容挂载到 host 上的 /opt/onex 目录，方便更新容器中的安装文件
  # 需要将存储组件的数据保存在类似 /data/onex.thirdpart 目录下，而非：/opt/onex/thirdpart 目录下
  # 并且，再创建时，要确保 /opt/onex 目录下内容为空
  # 请参考：https://stackoverflow.com/questions/66474191/how-does-volume-mount-from-container-to-host-and-vice-versa-work
  docker volume create --driver local -o type=none -o device=${ONEX_INSTALL_DIR} -o o=bind onex-volume

  echo "Create onex using ccr.ccs.tencentyun.com/superproj/onex-allinone-amd64:${ONEX_IMAGE_VERSION}"
  # 启动 onex 容器
  # 注意：启动时需要通过 --privileged 容器 root 权限，否则可能会报以下错误：
  # "Failed to get D-Bus connection: Operation not permitted"
  onex::common::network
  docker run -d --name onex \
    --network onex \
    --privileged \
    -v onex-volume:${ONEX_INSTALL_DIR}:rw \
    -p ${ONEX_ACCESS_HOST}:${ONEX_FAKESERVER_HTTP_PORT}:${ONEX_FAKESERVER_HTTP_PORT} \
    -p ${ONEX_ACCESS_HOST}:${ONEX_FAKESERVER_GRPC_PORT}:${ONEX_FAKESERVER_GRPC_PORT} \
    -p ${ONEX_ACCESS_HOST}:${ONEX_USERCENTER_HTTP_PORT}:${ONEX_USERCENTER_HTTP_PORT} \
    -p ${ONEX_ACCESS_HOST}:${ONEX_USERCENTER_GRPC_PORT}:${ONEX_USERCENTER_GRPC_PORT} \
    -p ${ONEX_ACCESS_HOST}:${ONEX_APISERVER_SECURE_PORT}:${ONEX_APISERVER_SECURE_PORT} \
    -p ${ONEX_ACCESS_HOST}:${ONEX_GATEWAY_HTTP_PORT}:${ONEX_GATEWAY_HTTP_PORT} \
    -p ${ONEX_ACCESS_HOST}:${ONEX_GATEWAY_GRPC_PORT}:${ONEX_GATEWAY_GRPC_PORT} \
    -p ${ONEX_ACCESS_HOST}:${ONEX_NIGHTWATCH_HEALTH_CHECK_PORT}:${ONEX_NIGHTWATCH_HEALTH_CHECK_PORT} \
    -p ${ONEX_ACCESS_HOST}:${ONEX_PUMP_HEALTH_CHECK_PORT}:${ONEX_PUMP_HEALTH_CHECK_PORT} \
    -p ${ONEX_ACCESS_HOST}:${ONEX_CACHESERVER_GRPC_PORT}:${ONEX_CACHESERVER_GRPC_PORT} \
    -p ${ONEX_ACCESS_HOST}:${ONEX_CONTROLLER_MANAGER_METRICS_PORT}:${ONEX_CONTROLLER_MANAGER_METRICS_PORT} \
    -p ${ONEX_ACCESS_HOST}:${ONEX_CONTROLLER_MANAGER_HEALTHZ_PORT}:${ONEX_CONTROLLER_MANAGER_HEALTHZ_PORT} \
    -p ${ONEX_ACCESS_HOST}:${ONEX_MINERSET_CONTROLLER_METRICS_PORT}:${ONEX_MINERSET_CONTROLLER_METRICS_PORT} \
    -p ${ONEX_ACCESS_HOST}:${ONEX_MINERSET_CONTROLLER_HEALTHZ_PORT}:${ONEX_MINERSET_CONTROLLER_HEALTHZ_PORT} \
    -p ${ONEX_ACCESS_HOST}:${ONEX_MINER_CONTROLLER_METRICS_PORT}:${ONEX_MINER_CONTROLLER_METRICS_PORT} \
    -p ${ONEX_ACCESS_HOST}:${ONEX_MINER_CONTROLLER_HEALTHZ_PORT}:${ONEX_MINER_CONTROLLER_HEALTHZ_PORT} \
    -p ${ONEX_ACCESS_HOST}:${ONEX_TOYBLC_HTTP_PORT}:${ONEX_TOYBLC_HTTP_PORT} \
    ccr.ccs.tencentyun.com/superproj/onex-allinone-amd64:${ONEX_IMAGE_VERSION}
  onex::onex::info
}

# Uninstall the docker container.
onex::onex::docker::uninstall()
{
  docker rm -f onex &>/dev/null
  docker volume rm onex-volume &>/dev/null
  onex::log::info "uninstall onex successfully"
}

onex::onex::build_artifacts()
{
  # 设置 Makefile 构建出产物的保存位置
  export OUTPUT_DIR=${LOCAL_OUTPUT_ROOT}

  # 进入到 OneX 项目仓库根目录，开始安装操作
  pushd "${ONEX_ROOT}" >/dev/null 2>&1

  # 构建需要的产物
  # 构建服务二进制文件
  echo "Building onex artifacts, this may take a while as it requires download go packages ..."
  make build

  # 生成 Systemd Unit 文件
  make gen.systemd

  # 生成应用配置文件
  make gen.appconfig

  # 生成 CA 证书
  make gen.ca

  # 生成 kubectl admin kubeconfig 文件
  make gen.kubeconfig
}

# Install the onex step by step.
# sbs is the abbreviation for "step by step".
onex::onex::sbs::install()
{
  onex::onex::prepare

  # 编译构建产物
  onex::onex::build_artifacts

  # 强制停止所有 OneX Systemd 服务，防止复制文件失败
  for service in "${ONEX_SERVER_SIDE_COMPONENTS[@]}"
  do
    set +o errexit
    onex::util::sudo "systemctl stop ${service}" &>/dev/null
    set -o errexit
  done

  OS=$(go env GOOS)
  ARCH=$(go env GOARCH)

  # 复制构建产物到指定的目录下
  # 为了确保当 ONEX_INSTALL_DIR 设置为 /opt/onex 时，也能安装成功，这里始终用
  # root 权限去安装
  ## 复制所有的二进制文件到安装目录
  echo "===========> Copying binaries"
  [ ! -d ${ONEX_BIN_DIR} ] && onex::util::sudo "mkdir -p ${ONEX_BIN_DIR}"
  onex::util::sudo "cp ${LOCAL_OUTPUT_ROOT}/platforms/${OS}/${ARCH}/* ${ONEX_BIN_DIR}/"

  ## 复制所有的配置文件到配置目录
  [ ! -d ${ONEX_CONFIG_DIR} ] && onex::util::sudo "mkdir -p ${ONEX_CONFIG_DIR}"
  echo "===========> Copying configurations"
  onex::util::sudo "cp ${LOCAL_OUTPUT_ROOT}/appconfig/* ${ONEX_CONFIG_DIR}/"

  ## 复制 CA 文件到配置目录
  echo "===========> Copying cert directory"
  onex::util::sudo "cp -r ${LOCAL_OUTPUT_ROOT}/cert ${ONEX_CONFIG_DIR}/"

  ## 复制 admin kubeconfig 文件到配置目录
  echo "===========> Copying kubeconfig file"
  onex::util::sudo "cp ${LOCAL_OUTPUT_ROOT}/config ${ONEX_CONFIG_DIR}/"

  echo "===========> Copying systemd unit files"
  onex::util::sudo "cp ${LOCAL_OUTPUT_ROOT}/systemd/* /etc/systemd/system/"

  # 启动 Systemd 服务
  # 依次启动服务
  onex::util::sudo "systemctl daemon-reload"
  for service in "${ONEX_SERVER_SIDE_COMPONENTS[@]}"
  do
    echo "===========> Starting ${service} service"
    onex::util::sudo "systemctl enable ${service}" &>/dev/null
    onex::util::sudo "systemctl restart ${service}"
  done

  onex::man::install
  onex::onex::info
  onex::log::info "install onex successfully"
}

# Uninstall the onex step by step.
onex::onex::sbs::uninstall()
{
  # 防止删除 / 目录
  [[ -z ${ONEX_INSTALL_DIR} ]] || [[ ${ONEX_INSTALL_DIR} == "/" ]] &&
    onex::log::error "OneX installation directory must be setted"

  # 这里要注意 uninstall 顺序
  #
  # 停掉所有的 OneX Systemd 服务
  for service in "${ONEX_SERVER_SIDE_COMPONENTS[@]}"
  do
    set +o errexit
    onex::util::sudo "systemctl stop ${service}" &>/dev/null
    set -o errexit
  done

  # 注意：这里不能删除整个 ${ONEX_INSTALL_DIR}，因为还有其他组件也安装在这个目录下，例如：thirdparty 目录
  onex::util::sudo "rm -rf ${ONEX_INSTALL_DIR}"
  onex::util::sudo "rm -f /etc/systemd/system/onex-*.service"
  onex::man::uninstall
  onex::log::info "uninstall onex successfully"
}

# Print necessary information after docker or sbs installation.
onex::onex::info()
{
  echo -e ${C_GREEN}onex has been installed, here are some useful information:${C_NORMAL}
  cat << EOF | sed 's/^/  /'
onex-fakeserver:
  http port: ${ONEX_FAKESERVER_HTTP_ADDR}
  grpc port: ${ONEX_FAKESERVER_GRPC_ADDR}
onex-usercenter:
  http port: ${ONEX_USERCENTER_HTTP_ADDR}
  grpc port: ${ONEX_USERCENTER_GRPC_ADDR}
onex-apiserver:
  http secure port: ${ONEX_APISERVER_SECURE_PORT}
  access onex-apiserver via kubectl: kubectl --kubeconfig=${ONEX_ADMIN_KUBECONFIG} api-resources
onex-gateway:
  http port: ${ONEX_GATEWAY_HTTP_ADDR}
  grpc port: ${ONEX_GATEWAY_GRPC_ADDR}
onex-nightwatch:
 health check port : ${ONEX_NIGHTWATCH_HEALTH_CHECK_PORT}
 health check path: ${ONEX_NIGHTWATCH_HEALTH_CHECK_PATH}
onex-pump:
 health check port : ${ONEX_PUMP_HEALTH_CHECK_PORT}
 health check path: ${ONEX_PUMP_HEALTH_CHECK_PATH}
onex-controller-manager:
  health check port: ${ONEX_CONTROLLER_MANAGER_HEALTHZ_PORT}
  metrics port: ${ONEX_CONTROLLER_MANAGER_METRICS_PORT}
onex-minerset-controller:
  health check port: ${ONEX_MINERSET_CONTROLLER_HEALTHZ_PORT}
  metrics port: ${ONEX_MINERSET_CONTROLLER_METRICS_PORT}
onex-miner-controller:
  health check port: ${ONEX_MINER_CONTROLLER_HEALTHZ_PORT}
  metrics port: ${ONEX_MINER_CONTROLLER_METRICS_PORT}
onex-cacheserver:
  grpc port: ${ONEX_CACHESERVER_GRPC_ADDR}
onex-toyblc:
  toyblc address: 0x210d9eD12CEA87E33a98AA7Bcb4359eABA9e800e
  toyblc p2p port: 6001
  toyblc peers: ws://localhost:6001
  toyblc http port: 56080
EOF
}

# Status check after docker or sbs installation.
onex::onex::sbs::status()
{
  has_failed=false
  for service in "${ONEX_SERVER_SIDE_COMPONENTS[@]}"
  do
      # 查看 service 的运行状态，如果输出中包含 active (running) 字样说明 service 成功启动。
      if ! systemctl status ${service} | grep -q 'active';then
        has_failed=true
        onex::log::info "$(printf "${C_RED}%25s failed to start, maybe not installed properly.${C_NORMAL}" ${service})"
      else
        onex::log::info "$(printf "${C_GREEN}%25s started successfully.${C_NORMAL}" ${service})"
      fi
  done

  # 只要有一个启动失败则认为启动失败
  [[ ${has_failed} == "true" ]] && return 1
}

# 要实现幂等
onex::onex::prepare()
{
  pushd "${ONEX_ROOT}" >/dev/null 2>&1

  # 2. 配置 $HOME/.bashrc 添加一些便捷入口
  if ! grep -q 'Alias for onex quick access' $HOME/.bashrc; then
    cat << 'EOF' >> $HOME/.bashrc
# Alias for onex quick access
export GOSRC="$WORKSPACE/golang/src"
# OneX project root directory, used in many places.
export ONEX_ROOT="$GOSRC/github.com/superproj/onex"
# Allows you to run latest compiled onex components like executing
# Linux commands, for example: onexctl.
export PATH=${ONEX_ROOT}/_output/platforms/linux/amd64:${ONEX_ROOT}/scripts:$PATH
export OVERSION=v1.0.0
# a very convenient alias used to enter superproj root directory.
alias sp="cd $GOSRC/github.com/superproj"
# a very convenient alias used to enter onex root directory.
alias o="cd $GOSRC/github.com/superproj/onex"
EOF
  fi

  # 初始化 MariaDB 数据库，创建 onex 数据库
  # 登录数据库并创建 onex 用户
  # 注意：给一个 MySQL 用户授权时，最好遵循最小权原则，只赋给他需要的数据库权限，例如：onex.*。
  # 但为了方便大家安装访问，这里授权 onex 用户可以访问所有数据库所有表的权限
  mariadb -h${ONEX_MYSQL_HOST} -P${ONEX_MYSQL_PORT} -u"${ONEX_MYSQL_ADMIN_USERNAME}" -p"${ONEX_MYSQL_ADMIN_PASSWORD}" << EOF
GRANT ALL PRIVILEGES ON *.* TO ${ONEX_MYSQL_USERNAME}@'%' identified by "${ONEX_MYSQL_PASSWORD}";
FLUSH PRIVILEGES;
EOF

  # 用 onex 用户登录 MySQL，执行 onex.sql 文件，创建 onex 数据库
  mariadb -h${ONEX_MYSQL_HOST} -P${ONEX_MYSQL_PORT} -u"${ONEX_MYSQL_ADMIN_USERNAME}" -p"${ONEX_MYSQL_ADMIN_PASSWORD}" << 'EOF'
source configs/onex.sql;
use onex;
INSERT INTO `uc_user` VALUES (0,'user-admin','admin',1,'admin','$2a$10$KeHjeGtHOuUYs6l76fgLSeDdjBgfv7loo89svN6p5r40XItHc/NV2', 'colin404@foxmail.com','181X',now(),now());
EOF

  # 初始化 MongoDB，创建 onex 用户
  encoded=$(echo -n "${ONEX_MONGO_ADMIN_PASSWORD}"|jq -sRr @uri)
  mongosh --quiet mongodb://${ONEX_MONGO_ADMIN_USERNAME}:${encoded}@${ONEX_MONGO_URL}/${ONEX_MONGO_DATABASE}?authSource=admin << EOF
db.createUser({user:"${ONEX_MONGO_USERNAME}",pwd:"${ONEX_MONGO_PASSWORD}",roles:["dbOwner"]})
db.auth("${ONEX_MONGO_USERNAME}", "${ONEX_MONGO_PASSWORD}")
quit;
EOF
}

if [[ "$*" =~ onex::onex:: ]]; then
  eval $*
fi
