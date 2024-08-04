#!/usr/bin/env bash

# Copyright 2020 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
# Use of this source code is governed by a MIT style
# license that can be found in the LICENSE file.

# The root of the build/dist directory.
ONEX_ROOT=$(dirname "${BASH_SOURCE[0]}")/../..
# If common.sh has already been sourced, it will not be sourced again here.
[[ -z ${COMMON_SOURCED} ]] && source ${ONEX_ROOT}/hack/installation/common.sh

# 安装后打印必要的信息
onex::man::info() {
cat << EOF
use: man onex-xxx to see onex-xxx help.
EOF
}

# 安装
onex::man::install()
{
  pushd ${ONEX_ROOT}

  # 生成各个组件的 man1 文件
  ${ONEX_ROOT}/hack/update-generated-docs.sh
  onex::util::sudo "cp docs/man/man1/* /usr/share/man/man1/"
  onex::man::status || return 1
  onex::man::info

  onex::log::info "install man pages successfully"
  popd
}

# 卸载
onex::man::uninstall()
{
  onex::util::sudo "rm -f /usr/share/man/man1/onex-*"
  onex::log::info "uninstall onex man pages successfully"
}

# 状态检查
onex::man::status()
{
  ls /usr/share/man/man1/onex-* &>/dev/null || {
    onex::log::error "onex man files not exist, maybe not installed properly"
    return 1
  }
}

if [[ $* =~ onex::man:: ]]; then
  eval $*
fi
