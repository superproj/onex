#!/bin/bash

# Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
# Use of this source code is governed by a MIT style
# license that can be found in the LICENSE file. The original repo for
# this file is https://github.com/superproj/onex.
#


set -e
set +u

if [ -z "$MAX_CPUS" ]; then
    MAX_CPUS=1

    case "$(uname -s)" in
    Darwin)
        MAX_CPUS=$(sysctl -n machdep.cpu.core_count)
        ;;
    Linux)
        CFS_QUOTA=$(cat /sys/fs/cgroup/cpu/cpu.cfs_quota_us)
        if [ "$CFS_QUOTA" -ge 100000 ]; then
            MAX_CPUS=$(("$CFS_QUOTA" / 100 / 1000))
        fi
        ;;
    *)
        # Unsupported host OS. Must be Linux or Mac OS X.
        ;;
    esac
fi

echo "$MAX_CPUS"
