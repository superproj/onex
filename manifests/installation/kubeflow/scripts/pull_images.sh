#!/bin/bash

registry_prefix="m.daocloud.io"

# Usage:
# ./pull_images.sh images-origin.list

image_list="$1"

# 读取镜像列表
while IFS= read -r image; do
  # 拼接完整的镜像路径
  full_image="${registry_prefix}/${image}"

  # 执行 Docker 拉取命令
  docker pull "$full_image" |tee -a /tmp/pull_images.log

  # 如果你还需要其他操作，可以在这里添加
done < ${image_list}


