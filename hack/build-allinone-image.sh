# Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
# Use of this source code is governed by a MIT style
# license that can be found in the LICENSE file. The original repo for
# this file is https://github.com/superproj/onex.
#

# build-allinone-image.sh 在 build/docker/onex-allinone 目录中以 build.sh 软连接
# 的形式存在，是为了防止误删 build/docker 目录
# Copy of hack/build-allinone-image.sh
ONEX_ROOT=$(dirname "${BASH_SOURCE[0]}")/../../..
ONEX_ENV_FILE=${ONEX_ENV_FILE:-${ONEX_ROOT}/manifests/env.local}

source "${ONEX_ROOT}/hack/lib/init.sh"
source ${ONEX_ENV_FILE}

cd ${ONEX_ROOT}

# 生成构建Dockerfile需要的构建产物
make build
make gen.systemd
make gen.appconfig
make gen.ca
make gen.kubeconfig

# 复制构建产物到指定目录
mkdir -p ${DST_DIR}/bin
cp ${OUTPUT_DIR}/platforms/${IMAGE_PLAT}/* ${DST_DIR}/bin/
cp -r ${OUTPUT_DIR}/appconfig ${DST_DIR}/
cp -r ${OUTPUT_DIR}/cert ${DST_DIR}/
cp -r ${OUTPUT_DIR}/config ${DST_DIR}/
cp -r ${OUTPUT_DIR}/systemd ${DST_DIR}/

