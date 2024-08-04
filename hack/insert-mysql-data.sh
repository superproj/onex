#!/bin/bash

# Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
# Use of this source code is governed by a MIT style
# license that can be found in the LICENSE file. The original repo for
# this file is https://github.com/superproj/onex.
#


function casbinrules() {
  mysql -h127.0.0.1 -uonex -p"onex(#)666" -D onex << EOF
INSERT INTO casbin_rule VALUES (id,'p', 'alice', 'data1', 'read', 'allow', '', '');
INSERT INTO casbin_rule VALUES (id,'p', 'bob', 'data2', 'write', 'deny', '', '');
INSERT INTO casbin_rule VALUES (id,'p', 'data2_admin', 'data2', 'read', 'allow', '', '');
INSERT INTO casbin_rule VALUES (id,'p', 'data2_admin', 'data2', 'write', 'allow', '', '');
INSERT INTO casbin_rule VALUES (id,'g', 'alice', 'data2_admin', 'deny', '', '', '');
EOF
}

$*
