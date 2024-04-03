#!/usr/bin/env bash

# Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
# Use of this source code is governed by a MIT style
# license that can be found in the LICENSE file. The original repo for
# this file is https://github.com/superproj/onex.
#


version="${VERSION}"
if [ "${version}" == "" ];then
  version=v`gsemver bump`
fi

if [ -z "`git tag -l ${version}`" ];then
  git tag -a -m "release version ${version}" ${version}
fi
