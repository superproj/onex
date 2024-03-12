# Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
# Use of this source code is governed by a MIT style
# license that can be found in the LICENSE file. The original repo for
# this file is https://github.com/superproj/onex.
#

# build-allinone-image.sh 在 build/docker/onex-allinone 目录中以 build.sh 软连接
# 的形式存在，是为了防止误删 build/docker 目录
# Copy of scripts/build-allinone-image.sh
ONEX_ROOT=$(dirname "${BASH_SOURCE[0]}")/../../..

source "${ONEX_ROOT}/scripts/installation/onex.sh"
# ONEX_ENV_FILE 变量来自 onex.sh
source ${ONEX_ENV_FILE}

cd ${ONEX_ROOT}

# 生成一个管理onex服务的管理脚本
function cat_onex_admin()
{
  cat << EOF
#!/bin/bash

function status()
{
  has_failed=false
  for service in ${ONEX_SERVER_SIDE_COMPONENTS[@]}
  do
      # 查看 service 的运行状态，如果输出中包含 active (running) 字样说明 service 成功启动。
      if ! systemctl status \${service} | grep -q 'active' &>/dev/null;then
        has_failed=true
        echo -e "\033[31mfailed to start \${service}, maybe not installed properly.\033[0m"
      else
        echo -e "\033[32mstarted \${service} successfully.\033[0m"
      fi
  done

  # 只要有一个启动失败则认为启动失败
  [[ \${has_failed} == "true" ]] && return 1
}

eval \$*
EOF
}

# 设置一些环境变量
export ONEX_MYSQL_HOST=onex-mariadb
export ONEX_REDIS_HOST=onex-redis
export ONEX_ETCD_HOST=onex-etcd
export ONEX_MONGO_HOST=onex-mongo
export ONEX_KAFKA_HOST=onex-kafka
export ONEX_JAEGER_HOST=onex-jaeger

# 生成构建Dockerfile需要的构建产物
onex::onex::build_artifacts

# 复制构建产物到指定目录
# OUTPUT_DIR, DST_DIR 由 Makefile 传入
# 整个构建产物保存位置传递路径为：onex::onex::build_artifacts -> Makefile -> 此脚本
# 整个传递路径唯一，所以本脚本能够正确或者 onex::onex::build_artifacts 中设置的产物
# 保存目录
mkdir -p ${DST_DIR}/bin
cp ${OUTPUT_DIR}/platforms/${IMAGE_PLAT}/* ${DST_DIR}/bin/
cp -r ${OUTPUT_DIR}/appconfig ${DST_DIR}/
cp -r ${OUTPUT_DIR}/cert ${DST_DIR}/
cp -r ${OUTPUT_DIR}/config ${DST_DIR}/
cp -r ${OUTPUT_DIR}/systemd ${DST_DIR}/
cat_onex_admin > ${DST_DIR}/onex-admin.sh
chmod +x ${DST_DIR}/onex-admin.sh
