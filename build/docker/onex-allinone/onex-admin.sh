#!/bin/bash

# Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
# Use of this source code is governed by a MIT style
# license that can be found in the LICENSE file. The original repo for
# this file is https://github.com/superproj/onex.
#


function status()
{
  has_failed=false
  for service in onex-usercenter onex-apiserver onex-gateway onex-nightwatch onex-pump onex-toyblc onex-controller-manager onex-minerset-controller onex-miner-controller onex-fakeserver onex-cacheserver
  do
      # 查看 service 的运行状态，如果输出中包含 active (running) 字样说明 service 成功启动。
      if ! systemctl status ${service} | grep -q 'active' &>/dev/null;then
        has_failed=true
        echo -e "\033[31mfailed to start ${service}, maybe not installed properly.\033[0m"
      else
        echo -e started "\033[32m${service} successfully.\033[0m"
      fi
  done

  # 只要有一个启动失败则认为启动失败
  [[ ${has_failed} == "true" ]] && return 1
}

eval $*
