#!/usr/bin/env bash

# Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
# Use of this source code is governed by a MIT style
# license that can be found in the LICENSE file. The original repo for
# this file is https://github.com/superproj/onex.
#

# The root of the build/dist directory
ONEX_ROOT=$(dirname "${BASH_SOURCE[0]}")/../..
[[ -z ${COMMON_SOURCED} ]] && source ${ONEX_ROOT}/scripts/installation/common.sh

ONEX_USERCENTER_ADDR=${ONEX_ACCESS_HOST}:${ONEX_USERCENTER_HTTP_PORT}
ONEX_GATEWAY_ADDR=${ONEX_ACCESS_HOST}:${ONEX_GATEWAY_HTTP_PORT}
ONEX_NIGHTWATCH_ADDR=${ONEX_ACCESS_HOST}:${ONEX_NIGHTWATCH_HEALTH_CHECK_PORT}
ONEX_PUMP_ADDR=${ONEX_ACCESS_HOST}:${ONEX_PUMP_HEALTH_CHECK_PORT}
ONEX_TOYBLC_ADDR=${ONEX_ACCESS_HOST}:${ONEX_TOYBLC_HTTP_PORT}

Header="-HContent-Type:application/json"
CURLARGS="-S"
CCURL="curl ${CURLARGS} -XPOST ${Header}" # Create
UCURL="curl ${CURLARGS} -XPUT ${Header}" # Update
RCURL="curl ${CURLARGS} -XGET" # Retrieve
DCURL="curl ${CURLARGS} -XDELETE" # Delete

# 重要：kubectl 访问 onex-apiserver，需要正确设置 KUBECONFIG 环境变量
export KUBECONFIG=${ONEX_ADMIN_KUBECONFIG}

# 确保 kubectl 命令被安装
if [[ "$(command -v kubectl)" == "" ]];then
  make -C ${ONEX_ROOT} tools.install.kubectl
fi

# 测试整个 OneX 项目
onex::test::test()
{
  onex::test::usercenter
  onex::test::apiserver
  onex::test::gateway
  onex::test::pump
  onex::test::nightwatch
  onex::test::controller # onex-controller-manager, onex-minerset-controller, onex-miner-controller
  onex::test::toyblc
  onex::test::onexctl
  onex::test::fakeserver
  onex::test::cacheserver

  onex::log::info "$(echo -e ${C_GREEN}===========\> all test passed!${C_NORMAL})"
}

# 幂等创建 colin 用户
onex::test::ensure_user()
{
  tester="colin"
  if [[ "$1" != "" ]];then
    tester="$1"
  fi

  curl -s -XPOST ${Header} http://${ONEX_USERCENTER_ADDR}/v1/users \
    -d'{"username":"'${tester}'","password":"onex(#)666","nickname":"colin","realName":"孔令飞","email":"colin404@foxmail.com","phone":"1812884xxxx"}'  | grep -q 'User already exists' && return 0
}

# 默认 admin 用户登录
# 如果是 `tester` 用户则会先创建 `tester` 用户
# 返回的是 refreshToken
onex::test::refresh_token()
{
  user="admin"
  if [[ "$1" != "" ]];then
    user=$1
  fi

  if [[ $user == "tester" ]];then
    onex::test::ensure_user tester
  fi

  ${CCURL} http://${ONEX_USERCENTER_ADDR}/v1/auth/login -d'{"username":"'$user'","password":"onex(#)666"}' | grep -Po 'refreshToken[" :]+\K[^"]+'
}

# 类似 onex::test::refresh_token，但返回的是 accessToken
onex::test::access_token()
{
  user="admin"
  if [[ "$1" != "" ]];then
    user=$1
  fi

  if [[ $user == "tester" ]];then
    onex::test::ensure_user tester
  fi

  ${CCURL} http://${ONEX_USERCENTER_ADDR}/v1/auth/login -d'{"username":"'$user'","password":"onex(#)666"}' | grep -Po 'accessToken[" :]+\K[^"]+'
}

# 测试 onex-usercenter 组件 user 资源
onex::test::usercenter::user()
{
  admin_token_header="-HAuthorization: Bearer $(onex::test::refresh_token)"

  # 1. 如果有 colin、mark、john 用户先清空
  ${DCURL} "${admin_token_header}" http://${ONEX_USERCENTER_ADDR}/v1/users/colin; echo
  ${DCURL} "${admin_token_header}" http://${ONEX_USERCENTER_ADDR}/v1/users/mark; echo
  ${DCURL} "${admin_token_header}" http://${ONEX_USERCENTER_ADDR}/v1/users/john; echo

  # 2. 创建 colin、mark、john 用户
  ${CCURL} http://${ONEX_USERCENTER_ADDR}/v1/users \
    -d'{"username":"colin","password":"onex(#)666","nickname":"colin","realName":"孔令飞","email":"colin404@ foxmail.com","phone":"1812884xxxx"}'; echo

  # 3. 列出所有用户
  ${RCURL} "${admin_token_header}" "http://${ONEX_USERCENTER_ADDR}/v1/users?offset=0&limit=10"; echo

  # 4. 获取 colin 用户的详细信息
  colin_token_header="-HAuthorization: Bearer $(onex::test::refresh_token colin)"
  ${RCURL} "${colin_token_header}" http://${ONEX_USERCENTER_ADDR}/v1/users/colin; echo

  # 5. 修改 colin 用户
  ${UCURL} "${colin_token_header}" http://${ONEX_USERCENTER_ADDR}/v1/users/colin \
    -d'{"nickname":"colin","email":"colin_modified@foxmail.com","phone":"1812884xxxx"}'; echo

  # 6. 修改用户密码
  ${UCURL} "${colin_token_header}" http://${ONEX_USERCENTER_ADDR}/v1/users/colin/update-password \
    -d'{"oldPassword":"onex(#)666","newPassword":"onex(#)888"}'; echo

  # 7. 删除 colin 用户
  ${DCURL} "${colin_token_header}" http://${ONEX_USERCENTER_ADDR}/v1/users/colin; echo
  onex::log::info "$(echo -e ${C_GREEN}===========\> /v1/user test passed!${C_NORMAL})"
}

# 测试 onex-usercenter auth 资源
onex::test::usercenter::auth()
{
  # Login
  tester_token_header="-HAuthorization: Bearer $(onex::test::refresh_token tester)"

  # RefreshToken
  token_string=$(${CCURL} "${tester_token_header}" http://${ONEX_USERCENTER_ADDR}/v1/auth/refresh-token)
  refresh_token=$(echo ${token_string} | jq -r .refreshToken)
  access_token=$(echo ${token_string} | jq -r .accessToken)
  refresh_token_header="-HAuthorization: Bearer ${refresh_token}"
  access_token_header="-HAuthorization: Bearer ${access_token}"

  # Authenticate
  ${CCURL} "${access_token_header}" http://${ONEX_USERCENTER_ADDR}/v1/auth/authenticate \
    -d'{"token":"'${access_token}'"}'; echo

  # /v1/auth/authorize 不需要认证
  ${CCURL} http://${ONEX_USERCENTER_ADDR}/v1/auth/authorize -d'{"sub":"tester","obj":"tester","act":"delete"}'; echo

  # Auth
  ${CCURL} "${access_token_header}" http://${ONEX_USERCENTER_ADDR}/v1/auth/auth \
    -d'{"token":"'${access_token}'","obj":"tester","act":"delete"}'; echo

  # Logout
  #${CCURL} "${refresh_token_header}" http://${ONEX_USERCENTER_ADDR}/v1/auth/logout; echo

  onex::log::info "$(echo -e ${C_GREEN}===========\> /v1/auth test passed!${C_NORMAL})"
}

# 测试 onex-usercenter secret 资源
onex::test::usercenter::secret()
{
  admin_token_header="-HAuthorization: Bearer $(onex::test::refresh_token admin)"
  tester_token_header="-HAuthorization: Bearer $(onex::test::refresh_token tester)"

  # 1. 如果有 secret0 密钥先清空
  ${DCURL} "${tester_token_header}" http://${ONEX_USERCENTER_ADDR}/v1/secrets/secret0; echo

  # 2. 创建 secret0 密钥
  ${CCURL} "${tester_token_header}" http://${ONEX_USERCENTER_ADDR}/v1/secrets \
    -d'{"name":"secret0","expires":0,"description":"tester secret"}'; echo

  # 3. 列出所有密钥
  ${RCURL} "${admin_token_header}" http://${ONEX_USERCENTER_ADDR}/v1/secrets; echo

  # 4. 获取 secret0 密钥的详细信息
  ${RCURL} "${tester_token_header}" http://${ONEX_USERCENTER_ADDR}/v1/secrets/secret0; echo

  # 5. 修改 secret0 密钥
  ${UCURL} "${tester_token_header}" http://${ONEX_USERCENTER_ADDR}/v1/secrets/secret0 \
    -d'{"expires":4072326717,"description":"tester secret(modified)"}'; echo

  # 6. 获取 secret0 密钥的详细信息
  ${RCURL} "${tester_token_header}" http://${ONEX_USERCENTER_ADDR}/v1/secrets/secret0; echo

  # 7. 删除 secret0 密钥
  ${DCURL} "${tester_token_header}" http://${ONEX_USERCENTER_ADDR}/v1/secrets/secret0; echo
  onex::log::info "$(echo -e ${C_GREEN}===========\> /v1/secret test passed!${C_NORMAL})"
}

# 注意：这里要sleep等待controller调和资源
# 包含了 onex-controller-manager, onex-minerset-controller, onex-miner-controller 3 个 controller 的测试
onex::test::controller()
{
  NS=user-admin
  # 先幂等删除
  kubectl delete -f ${ONEX_ROOT}/manifests/sample/onex/chain.yaml &>/dev/null || true
  kubectl delete -f ${ONEX_ROOT}/manifests/sample/onex/minerset.yaml &>/dev/null || true
  kubectl delete -f ${ONEX_ROOT}/manifests/sample/onex/miner.yaml &>/dev/null || true

  # 等待 2s 等 controller 删除资源
  sleep 2

  # 创建一个私有链
  kubectl create -f ${ONEX_ROOT}/manifests/sample/onex/chain.yaml
  kubectl -n kube-system get chain --no-headers|grep -q genesis
  sleep 1
  kubectl -n kube-system get miner | egrep -q 'genesis.*Running'

  # 创建一个矿机池
  kubectl create -f ${ONEX_ROOT}/manifests/sample/onex/minerset.yaml
  sleep 2
  kubectl -n ${NS} get minerset | grep -q test
  kubectl -n ${NS} get miner | egrep test-.*Running

  # 创建一个游离的矿机
  kubectl create -f ${ONEX_ROOT}/manifests/sample/onex/miner.yaml
  sleep 1
  kubectl -n ${NS} get miner | egrep freeminer.*Running

  onex::log::info "$(echo -e ${C_GREEN}===========\> onex controller test passed!${C_NORMAL})"
}

# 测试 onex-apiserver 组件
onex::test::apiserver()
{
  kubectl api-resources | egrep -q 'apps.onex.io'
}

# 测试 onex-gateway 组件
onex::test::gateway()
{

  onex::test::gateway::chain
  onex::test::gateway::minerset
  onex::test::gateway::miner

  onex::log::info "$(echo -e ${C_GREEN}===========\> onex-gateway test passed!${C_NORMAL})"
}


# 确保创世区块存在
onex::test::gateway::ensure_genesis_chain()
{
  kubectl create -f ${ONEX_ROOT}/manifests/sample/onex/chain.yaml || true
  kubectl -n kube-system get chain genesis
}

# 测试 onex-gateway chacin 资源
onex::test::gateway::chain()
{
  # 创建一个创世区块链
  kubectl create -f ${ONEX_ROOT}/manifests/sample/onex/chain.yaml || true
  kubectl -n kube-system get chain genesis

  onex::log::info "$(echo -e ${C_GREEN}===========\> /v1/chains test passed!${C_NORMAL})"
}

# 测试 onex-fakeserver
onex::test::fakeserver()
{
  # Leave to you to implement it.
  onex::log::info "$(echo -e ${C_GREEN}===========\> onex-fakeserver test passed!${C_NORMAL})"
}

# 测试 onex-cacheserver 组件
onex::test::cacheserver()
{
  # Leave to you to implement it.
  onex::log::info "$(echo -e ${C_GREEN}===========\> onex-cacheserver test passed!${C_NORMAL})"
}

# 测试 onex-gateway 组件
onex::test::gateway::minerset()
{
  # 确保依赖的创世链存在
  onex::test::gateway::ensure_genesis_chain

  # Login
  tester_token_header="-HAuthorization: Bearer $(onex::test::access_token tester)"

  # 先幂等删除
  ${DCURL} "${tester_token_header}" http://${ONEX_GATEWAY_ADDR}/v1/minersets/minerset0; echo

  # 获取幂等 token
  idempotent_token=$(${RCURL} "${tester_token_header}" http://${ONEX_GATEWAY_ADDR}/v1/idempotents | jq -r .token)

  # 创建 minerset0 矿机池
  ${CCURL} "${tester_token_header}" -H"X-Idempotent-ID: ${idempotent_token}" http://${ONEX_GATEWAY_ADDR}/v1/minersets \
    -d'{"apiVersion":"apps.onex.io/v1beta1","kind":"MinerSet","metadata":{"name":"minerset0"},"spec":{"deletePolicy":"Random","displayName":"test-minerset","replicas":2,"template":{"spec":{"chainName":"genesis","minerType":"M1.MEDIUM2"}}}}'; echo

  # 矿机池列表
  ${RCURL} "${tester_token_header}" "http://${ONEX_GATEWAY_ADDR}/v1/minersets?offset=0&limit=10"; echo

  # 获取矿机池详情
  ${RCURL} "${tester_token_header}" "http://${ONEX_GATEWAY_ADDR}/v1/minersets/minerset0"; echo

  # 更新矿机池
  ${UCURL} "${tester_token_header}" http://${ONEX_GATEWAY_ADDR}/v1/minersets -d'{"metadata":{"name":"minerset0"},"spec":{"replicas":2,"selector":{},"template":{"metadata":{},"spec":{"metadata":{},"minerType":"M1.MEDIUM2","chainName":"genesis"}},"displayName":"test-minerset-modified"}}'; echo

  # 扩缩容矿机池
  # 等待 1 s，等待上一次操作调和完成
  sleep 1
  ${UCURL} "${tester_token_header}" http://${ONEX_GATEWAY_ADDR}/v1/minersets/minerset0/scale -d'{"replicas":3}'; echo

  # 删除矿机池
  ${DCURL} "${tester_token_header}" http://${ONEX_GATEWAY_ADDR}/v1/minersets/minerset0; echo

  onex::log::info "$(echo -e ${C_GREEN}===========\> /v1/minersets test passed!${C_NORMAL})"
}

# 测试 onex-miner-controller
onex::test::gateway::miner()
{
  # 确保依赖的创世链存在
  onex::test::gateway::ensure_genesis_chain

  # Login
  tester_token_header="-HAuthorization: Bearer $(onex::test::access_token tester)"

  # 先幂等删除
  ${DCURL} "${tester_token_header}" http://${ONEX_GATEWAY_ADDR}/v1/miners/miner0; echo

  # 获取幂等 token
  idempotent_token=$(${RCURL} "${tester_token_header}" http://${ONEX_GATEWAY_ADDR}/v1/idempotents | jq -r .token)

  # 创建 miner0 矿机池
  ${CCURL} "${tester_token_header}" -H"X-Idempotent-ID: ${idempotent_token}" http://${ONEX_GATEWAY_ADDR}/v1/miners \
    -d'{"metadata":{"name":"miner0"},"spec":{"chainName":"genesis","minerType":"M1.MEDIUM2"},"apiVersion":"apps.onex.io/v1beta1","kind":"Miner"}'; echo

  # 矿机池列表
  ${RCURL} "${tester_token_header}" "http://${ONEX_GATEWAY_ADDR}/v1/miners?offset=0&limit=10"; echo

  # 获取矿机池详情
  ${RCURL} "${tester_token_header}" "http://${ONEX_GATEWAY_ADDR}/v1/miners/miner0"; echo

  # 更新矿机池
  ${UCURL} "${tester_token_header}" http://${ONEX_GATEWAY_ADDR}/v1/miners \
    -d'{"metadata":{"name":"miner0"},"spec":{"chainName":"genesis","minerType":"M1.MEDIUM2","displayName":"test-for-gateway"},"apiVersion":"apps.onex.io/v1beta1","kind":"Miner"}'; echo

  # 删除矿机池
  ${DCURL} "${tester_token_header}" http://${ONEX_GATEWAY_ADDR}/v1/miners/miner0; echo

  onex::log::info "$(echo -e ${C_GREEN}===========\> /v1/miners test passed!${C_NORMAL})"
}

# 测试 onex-usercenter 组件
onex::test::usercenter()
{
  onex::test::usercenter::user
  onex::test::usercenter::secret
  onex::test::usercenter::auth
  onex::log::info "$(echo -e ${C_GREEN}===========\> onex-usercenter test passed!${C_NORMAL})"
}

# 测试 onex-pump 组件
onex::test::pump()
{
  ${RCURL} http://${ONEX_PUMP_ADDR}${ONEX_PUMP_HEALTH_CHECK_PATH} | egrep 'status.*ok'
  onex::log::info "$(echo -e ${C_GREEN}===========\> onex-pump test passed!${C_NORMAL})"
}

# 测试 onex-nightwatch 组件
onex::test::nightwatch()
{
  ${RCURL} http://${ONEX_NIGHTWATCH_ADDR}${ONEX_NIGHTWATCH_HEALTH_CHECK_PATH} | egrep 'status.*ok'
  onex::log::info "$(echo -e ${C_GREEN}===========\> onex-nightwatch test passed!${C_NORMAL})"
}

# 测试 onex-toyblc 组件
onex::test::toyblc()
{
  ${RCURL} -u${ONEX_TOYBLC_USERNAME}:${ONEX_TOYBLC_PASSWORD} http://${ONEX_TOYBLC_ADDR}/v1/blocks; echo
  ${RCURL} -u${ONEX_TOYBLC_USERNAME}:${ONEX_TOYBLC_PASSWORD} http://${ONEX_TOYBLC_ADDR}/v1/peers; echo

  onex::log::info "$(echo -e ${C_GREEN}===========\> onex-toyblc test passed!${C_NORMAL})"
}

# 测试 onexctl 组件
onex::test::onexctl()
{
  # 创建一个测试矿机池
  kubectl delete -f ${ONEX_ROOT}/manifests/sample/onex/minerset.yaml &>/dev/null || true
  kubectl create -f ${ONEX_ROOT}/manifests/sample/onex/minerset.yaml
  ${ONEX_BIN_DIR}/onexctl --config ${ONEX_CONFIG_DIR}/onexctl.yaml minerset list | grep -q testminerset

  onex::log::info "$(echo -e ${C_GREEN}===========\> onexctl test passed!${C_NORMAL})"
}

if [[ "$*" =~ onex::test:: ]]; then
  eval $*
fi
