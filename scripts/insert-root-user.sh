#!/bin/bash

# Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
# Use of this source code is governed by a MIT style
# license that can be found in the LICENSE file. The original repo for
# this file is https://github.com/superproj/onex.
#


mysql -h127.0.0.1 -P3306 -uroot -p'onex(#)666' << EOF
grant all on *.* TO 'onex'@'%' identified by "onex(#)666";
flush privileges;
EOF
