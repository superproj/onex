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
ONEX_JAEGER_HOST=${ONEX_JAEGER_HOST:-127.0.0.1}
ONEX_JAEGER_PORT=${ONEX_JAEGER_PORT:-4317}

# Install Jaeger using containerization.
onex::jaeger::docker::install()
{
  onex::common::network
  docker run -d --name onex-jaeger \
    --restart always \
    --network onex \
    -e COLLECTOR_ZIPKIN_HOST_PORT=:9411 \
    -p ${ONEX_ACCESS_HOST}:6831:6831/udp \
    -p ${ONEX_ACCESS_HOST}:6832:6832/udp \
    -p ${ONEX_ACCESS_HOST}:5778:5778 \
    -p ${ONEX_ACCESS_HOST}:16686:16686 \
    -p ${ONEX_ACCESS_HOST}:${ONEX_JAEGER_PORT}:4317 \
    -p ${ONEX_ACCESS_HOST}:4318:4318 \
    -p ${ONEX_ACCESS_HOST}:14250:14250 \
    -p ${ONEX_ACCESS_HOST}:14268:14268 \
    -p ${ONEX_ACCESS_HOST}:14269:14269 \
    -p ${ONEX_ACCESS_HOST}:9411:9411 \
    jaegertracing/all-in-one:1.52

  sleep 2
  onex::jaeger::status || return 1
  onex::jaeger::info
  onex::log::info "install jaeger successfully"
}

# Uninstall the docker container.
onex::jaeger::docker::uninstall()
{
  docker rm -f onex-jaeger &>/dev/null
  onex::log::info "uninstall jaeger successfully"
}

# Install the jaeger step by step.
# sbs is the abbreviation for "step by step".
onex::jaeger::sbs::install()
{
  onex::jaeger::docker::install
  onex::log::info "install jaeger successfully"
}

# Uninstall the jaeger step by step.
onex::jaeger::sbs::uninstall()
{
  onex::jaeger::docker::uninstall
  onex::log::info "uninstall jaeger successfully"
}

# Print necessary information after docker or sbs installation.
onex::jaeger::info()
{
  echo -e ${C_GREEN}Jaeger has been installed, here are some useful information:${C_NORMAL}
  cat << EOF | sed 's/^/  /'
OpenTelemetry Protocol (OTLP) over gRPC Endpoint: ${ONEX_JAEGER_HOST}:${ONEX_JAEGER_PORT}
EOF
}

# Status check after docker or sbs installation.
onex::jaeger::status()
{
  onex::util::telnet ${ONEX_JAEGER_HOST} ${ONEX_JAEGER_PORT} || return 1
}

if [[ "$*" =~ onex::jaeger:: ]]; then
  eval $*
fi
