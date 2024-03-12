#!/usr/bin/env bash

# Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
# Use of this source code is governed by a MIT style
# license that can be found in the LICENSE file. The original repo for
# this file is https://github.com/superproj/onex.
#

for top in pkg
do
  for d in $(find $top -type d)
  do
    if [[ "$d" =~ "pkg/generated" ]];then
      continue
    fi

    if [ ! -f $d/doc.go ]; then
      if ls $d/*.go > /dev/null 2>&1; then
        echo $d/doc.go
        echo "package $(basename $d) // import \"github.com/superproj/onex/$d\"" > $d/doc.go
      fi
    fi
  done
done
