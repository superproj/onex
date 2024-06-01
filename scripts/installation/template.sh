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
ONEX_TEMPLATE_HOST=${ONEX_TEMPLATE_HOST:-127.0.0.1}
ONEX_TEMPLATE_PORT=${ONEX_TEMPLATE_PORT:-xxxx}

# Install template using containerization.
onex::template::docker::install()
{
  # docker run -d --restart always --name onex-xxx ...
  onex::template::status || return 1
  onex::template::info
  onex::log::info "install template successfully"
}

# Uninstall the docker container.
onex::template::docker::uninstall()
{
  # docker rm -f onex-xxx &>/dev/null
  onex::log::info "uninstall template successfully"
}

# Install the template step by step.
# sbs is the abbreviation for "step by step".
onex::template::sbs::install()
{
  onex::log::info "install template successfully"
}

# Uninstall the template step by step.
onex::template::sbs::uninstall()
{
  onex::log::info "uninstall template successfully"
}

# Print necessary information after docker or sbs installation.
onex::template::info()
{
  echo -e ${C_GREEN}template has been installed, here are some useful information:${C_NORMAL}
  cat << EOF | sed 's/^/  /'
some useful information
EOF
}

# Status check after docker or sbs installation.
onex::template::status()
{
  onex::util::telnet ${ONEX_TEMPLATE_HOST} ${ONEX_TEMPLATE_PORT} || return 1
}

if [[ "$*" =~ onex::template:: ]];then
  eval $*
fi
