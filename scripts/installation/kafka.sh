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
ONEX_KAFKA_HOST=${ONEX_KAFKA_HOST:-127.0.0.1}
ONEX_KAFKA_PORT=${ONEX_KAFKA_PORT:-4317}

# Install kafka using containerization.
# Refer to https://www.baeldung.com/ops/kafka-docker-setup
onex::kafka::docker::install()
{
  onex::common::network
  docker run -d --restart always --name onex-zookeeper --network onex -p 2181:2181 -t wurstmeister/zookeeper
  docker run -d --name onex-kafka --link onex-zookeeper:zookeeper \
    --restart always \
    --network onex \
    --restart=always \
    -v /etc/localtime:/etc/localtime \
    -p ${ONEX_KAFKA_HOST}:${ONEX_KAFKA_PORT}:9092 \
    --env KAFKA_ZOOKEEPER_CONNECT=zookeeper:2181 \
    --env KAFKA_ADVERTISED_HOST_NAME=${ONEX_KAFKA_HOST} \
    --env KAFKA_ADVERTISED_PORT=${ONEX_KAFKA_PORT} \
    wurstmeister/kafka

  echo "Sleeping to wait for onex-kafka container to complete startup ..."
  sleep 5
  onex::kafka::status || return 1
  onex::kafka::info
  onex::log::info "install kafka successfully"
}

# Uninstall the docker container.
onex::kafka::docker::uninstall()
{
  docker rm -f onex-zookeeper &>/dev/null
  docker rm -f onex-kafka &>/dev/null
  onex::log::info "uninstall kafka successfully"
}

# Install the kafka step by step.
# sbs is the abbreviation for "step by step".
# Refer to https://kafka.apache.org/documentation/#quickstart
onex::kafka::sbs::install()
{
  onex::kafka::docker::install
  onex::log::info "install kafka successfully"
}

# Uninstall the kafka step by step.
onex::kafka::sbs::uninstall()
{
  onex::kafka::docker::uninstall
  onex::log::info "uninstall kafka successfully"
}

# Print necessary information after docker or sbs installation.
onex::kafka::info()
{
  echo -e ${C_GREEN}kafka has been installed, here are some useful information:${C_NORMAL}
  cat << EOF | sed 's/^/  /'
Kafka brokers is: ${ONEX_KAFKA_HOST}:${ONEX_KAFKA_PORT}
EOF
}

# Status check after docker or sbs installation.
onex::kafka::status()
{
  onex::util::telnet ${ONEX_KAFKA_HOST} ${ONEX_KAFKA_PORT} || return 1
}

if [[ $* =~ onex::kafka:: ]]; then
  eval $*
fi
