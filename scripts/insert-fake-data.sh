#!/bin/bash

# Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
# Use of this source code is governed by a MIT style
# license that can be found in the LICENSE file. The original repo for
# this file is https://github.com/superproj/onex.
#


function usage() {
  cat << EOF

Usage: $0 <fake-user-number>
Insert fake data into onex database.

Example: ./insert-fake-data.sh 500

Reprot bugs to <colin404@foxmail.com>.
EOF
}


# 该函数用来向 onex 数据库中，插入 $1 个假用户，每个用户创建 4 个假密钥对
function insert_fake_data() {
  current_time=$(date +%s)

  # 先删除已有的假数据，防止因为 UNIQUE KEY 报错
  mysql -h127.0.0.1 -uonex -p"onex(#)666" -D onex << EOF
delete from uc_user where nickname='fakedata';
delete from uc_secret where description='fakedata';
EOF

  for n in $(seq 1 1 $1)
  do
    mysql -h127.0.0.1 -uonex -p"onex(#)666" -D onex << EOF
insert into uc_user values(id,'user-$n','user$n',1,'fakedata','fakepassword','fake@qq.com','xxx',now(),now());
insert into uc_secret values(id,'user-$n','fakesecret','secret1$n','b1',1,0,'fakedata',now(),now());
insert into uc_secret values(id,'user-$n','fakesecret','secret2$n','b2',1,0,'fakedata',now(),now());
insert into uc_secret values(id,'user-$n','fakesecret','secret3$n','b3',1,0,'fakedata',now(),now());
insert into uc_secret values(id,'user-$n','fakesecret','secret4$n','b4',1,0,'fakedata',now(),now());
EOF
  done
}

while getopts "h" opt;do
  case ${opt} in
    h)
      usage
      exit 0
      ;;
    ?)
      usage
      exit 0
      ;;
  esac
done

insert_fake_data $1
