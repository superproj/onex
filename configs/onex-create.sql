-- Copyright 2022 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
-- Use of this source code is governed by a MIT style
-- license that can be found in the LICENSE file. The original repo for
-- this file is https://github.com/superproj/onex.
--

-- uc_user

CREATE TABLE `uc_user` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  `user_id` varchar(253) NOT NULL DEFAULT '' COMMENT '用户 ID',
  `username` varchar(253) NOT NULL DEFAULT '' COMMENT '用户名称',
  `status` varchar(64) NOT NULL DEFAULT '' COMMENT '用户状态：registered,active,disabled,blacklisted,locked,deleted',
  `nickname` varchar(253) NOT NULL DEFAULT '' COMMENT '用户昵称',
  `password` varchar(64) NOT NULL DEFAULT '' COMMENT '用户加密后的密码',
  `email` varchar(253) NOT NULL DEFAULT '' COMMENT '用户电子邮箱',
  `phone` varchar(16) NOT NULL DEFAULT '' COMMENT '用户手机号',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_username` (`username`),
  UNIQUE KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB AUTO_INCREMENT=0 DEFAULT CHARSET=utf8mb4 COMMENT='用户表';

-- uc_secret

CREATE TABLE `uc_secret` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  `user_id` varchar(253) NOT NULL DEFAULT '' COMMENT '用户 ID',
  `name` varchar(253) NOT NULL DEFAULT '' COMMENT '密钥名称',
  `secret_id` varchar(36) NOT NULL DEFAULT '' COMMENT '密钥 ID',
  `secret_key` varchar(36) NOT NULL DEFAULT '' COMMENT '密钥 Key',
  `status` tinyint unsigned NOT NULL DEFAULT 1 COMMENT '密钥状态，0-禁用；1-启用',
  `expires` bigint(64) NOT NULL DEFAULT 0 COMMENT '0 永不过期',
  `description` varchar(255) NOT NULL DEFAULT '' COMMENT '密钥描述',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_secret_id` (`secret_id`),
  KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB AUTO_INCREMENT=0 DEFAULT CHARSET=utf8mb4 COMMENT='密钥表';

-- api_chain

CREATE TABLE `api_chain` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  `namespace` varchar(253) NOT NULL DEFAULT '' COMMENT '命名空间',
  `name` varchar(253) NOT NULL DEFAULT '' COMMENT '区块链名',
  `display_name` varchar(253) NOT NULL DEFAULT '' COMMENT '区块链展示名',
  `miner_type` varchar(16) NOT NULL DEFAULT '' COMMENT '区块链矿机机型',
  `image` varchar(253) NOT NULL DEFAULT '' COMMENT '区块链镜像 ID',
  `min_mine_interval_seconds` int(8) NOT NULL DEFAULT 0 COMMENT '矿机挖矿间隔',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_namespace_name` (`namespace`,`name`)
) ENGINE=InnoDB AUTO_INCREMENT=0 DEFAULT CHARSET=utf8mb4 COMMENT='区块链表';


-- api_minerset

CREATE TABLE `api_minerset` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  `namespace` varchar(253) NOT NULL DEFAULT '' COMMENT '命名空间',
  `name` varchar(253) NOT NULL DEFAULT '' COMMENT '矿机池名',
  `replicas` int(8) NOT NULL DEFAULT 0 COMMENT '矿机副本数',
  `display_name` varchar(253) NOT NULL DEFAULT '' COMMENT '矿机池展示名',
  `delete_policy` varchar(32) NOT NULL DEFAULT '' COMMENT '矿机池缩容策略',
  `min_ready_seconds` int(8) NOT NULL DEFAULT 0 COMMENT '矿机 Ready 最小等待时间',
  `fully_labeled_replicas` int(8) NOT NULL DEFAULT 0 COMMENT '所有标签匹配的副本数',
  `ready_replicas` int(8) NOT NULL DEFAULT 0 COMMENT 'Ready 副本数',
  `available_replicas` int(8) NOT NULL DEFAULT 0 COMMENT '可用副本数',
  `failure_reason` longtext DEFAULT NULL COMMENT '失败原因',
  `failure_message` longtext DEFAULT NULL COMMENT '失败信息',
  `conditions` longtext DEFAULT NULL COMMENT '矿机池状态',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_namespace_name` (`namespace`,`name`)
) ENGINE=InnoDB AUTO_INCREMENT=0 DEFAULT CHARSET=utf8mb4 COMMENT='矿机池表';

-- api_miner

CREATE TABLE `api_miner` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  `namespace` varchar(253) NOT NULL DEFAULT '' COMMENT '命名空间',
  `name` varchar(253) NOT NULL DEFAULT '' COMMENT '矿机名',
  `display_name` varchar(253) NOT NULL DEFAULT '' COMMENT '矿机展示名',
  `phase` varchar(45) NOT NULL DEFAULT '' COMMENT '矿机状态',
  `miner_type` varchar(16) NOT NULL DEFAULT '' COMMENT '矿机机型',
  `chain_name` varchar(253) NOT NULL DEFAULT '' COMMENT '矿机所属的区块链名',
  `cpu` int(8) NOT NULL DEFAULT 0 COMMENT '矿机 CPU 规格',
  `memory` int(8) NOT NULL DEFAULT 0 COMMENT '矿机内存规格',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_namespace_name` (`namespace`,`name`),
  KEY `idx_chain_name` (`chain_name`)
) ENGINE=InnoDB AUTO_INCREMENT=0 DEFAULT CHARSET=utf8mb4 COMMENT='矿机表';

-- fs_order

CREATE TABLE `fs_order` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `order_id` longtext DEFAULT NULL,
  `customer` longtext DEFAULT NULL,
  `product` longtext DEFAULT NULL,
  `quantity` bigint(20) DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=0 DEFAULT CHARSET=utf8mb4 COMMENT='订单表';
