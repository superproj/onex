#!/usr/bin/env bash

# Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
# Use of this source code is governed by a MIT style
# license that can be found in the LICENSE file. The original repo for
# this file is https://github.com/superproj/onex.
#


set -o errexit

ONEX_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
source "${ONEX_ROOT}/scripts/common.sh"

OVERSION=${OVERSION:-$(uplift tag --current --silent)+$(date +'%Y%m%d%H%M%S')}
COMPONENTS=()
DEPLOY=false
IMAGE=false
PUSH=false
LOAD=false
ONEX_KIND_CLUSTER_NAME=${ONEX_KIND_CLUSTER_NAME:-onex}

# Get file names from COMMAND LINE arguments
getcomponents() {
  for f in "$@"; do
    COMPONENTS[${#COMPONENTS[*]}]=${ONEX_ALL_COMPONENTS[$1]:-$1}
  done
}

# Load docker images to kind cluster
load_docker_image() {
  # Only kind cluster need to load image.
  if ! kubectl config view | grep -q 'cluster: kind-';then
    return
  fi

  NODES=${KIND_LOAD_NODES:-$(kubectl get nodes|awk '/Ready/ && !/SchedulingDisabled/{nodes=nodes$1","} END{gsub(/,$/,"",nodes);print nodes}')}

  local comp="$1"

  # In the SemVer versioning specification, the use of the "+" sign is possible, but container image tag names
  # do not support the "+" character. Therefore, it is necessary to replace the "+" with "-" in the version number.
  # For example, the version number "v0.18.0+20240121235656" should be transformed into "v0.18.0-20240121235656" for
  # use as a container tag name.
  kind load docker-image --name ${ONEX_KIND_CLUSTER_NAME} --nodes ${NODES} \
    ccr.ccs.tencentyun.com/superproj/${comp}-amd64:$(echo ${OVERSION} | sed 's/+/-/')
}

# Build docker images
build_image() {
  local cmd=image
  [[ ${PUSH} == true ]] && cmd=push

  for comp in "${COMPONENTS[@]}"
  do
    make -C ${ONEX_ROOT} ${cmd} IMAGES=${comp} VERSION=${OVERSION} MULTISTAGE=0
    [[ "$LOAD" == true ]] && load_docker_image ${comp}
  done
}

# Only build component
build() {
  for comp in "${COMPONENTS[@]}"
  do
    make -C ${ONEX_ROOT} build BINS=${comp} VERSION=${OVERSION}
  done
}

# Build docker images and deploy them
deploy() {
  for comp in "${COMPONENTS[@]}"
  do
    make -C ${ONEX_ROOT} deploy DEPLOYS=${comp} VERSION=${OVERSION}
    load_docker_image ${comp}
    kubectl rollout restart deployment ${comp}
  done
}

# Print usage infomation
usage()
{
  readonly PROG=${0##*/}
  cat << EOF

Usage: ${PROG} [ OPTIONS ] SHORTNAME [-d]
build suger script.

  SHORTNAME              short name for onex component.

OPTIONS:
  -h, --help             usage information.
  -d, --deploy           whether to deploy component to kind cluster (build image and deploy).
  -i, --image            build image only.
      --load             load docker image to kind cluster. Only work when \`-i\` options is specified.
  -v, --version          build or deploy version.

Reprot bugs to <colin404@foxmail.com>.
EOF
}

# Print message to standerr
die()
{
  echo "$@" >&2
  exit 1
}

# Check the argument associate with a option
requiredarg()
{
  [ -z "$2" -o "$(echo $2 | awk '$0~/^-/{print 1}')" == "1" ] && die "$0: option $1 requires an argument"
  ((args++))
}


### read cli options
# separate groups of short options. replace --foo=bar with --foo bar
while [[ -n $1 ]]; do
  case "$1" in
    -- )
      for arg in "$@"; do
        ARGS[${#ARGS[*]}]="$arg"
      done
      break
      ;;
    --*=?* )
      ARGS[${#ARGS[*]}]="${1%%=*}"
      ARGS[${#ARGS[*]}]="${1#*=}"
      ;;
    --* )
      #die "$0: option $1 requires a value"
      ARGS[${#ARGS[*]}]="$1"
      ;;
    -* )
      for shortarg in $(sed -e 's|.| -&|g' <<< "${1#-}"); do
        ARGS[${#ARGS[*]}]="$shortarg"
      done
      ;;
    * )
      ARGS[${#ARGS[*]}]="$1"
  esac
  shift
done

# set the separated options as input options.
set -- "${ARGS[@]}"

while [[ -n $1 ]]; do
  ((args=1))
  case "$1" in
    -h | --help )
      usage
      exit 0
      ;;
    -d | --deploy )
      DEPLOY="true"
      ;;
    -i | --image )
      IMAGE="true"
      ;;
    -p | --push)
      IMAGE="true"
      PUSH="true"
      ;;
    --load )
      LOAD="true"
      ;;
    -v | --version )
      requiredarg "$@"
      OVERSION="$2"
      ;;
    -* )
      die "$0: unrecognized option '$1'"
      ;;
    *)
      getcomponents "$1"
      ;;
  esac
  shift $args
done

if [  "${#COMPONENTS[@]}" -eq 0 ];then
  COMPONENTS=("${ONEX_ALL_COMPONENTS[@]}")
fi

if [ "${DEPLOY}" == true ];then
  deploy
elif [ "${IMAGE}" == true ];then
  build_image
else
  build
fi
