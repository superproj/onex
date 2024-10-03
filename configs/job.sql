CREATE TABLE `nw_cronjob` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  `cronjob_id` varchar(100) NOT NULL DEFAULT '' COMMENT 'CronJob ID',
  `user_id` varchar(100) NOT NULL DEFAULT '' COMMENT '创建人',
  `scope` varchar(256) NOT NULL DEFAULT 'default' COMMENT 'CronJob 作用域',
  `name` varchar(255) NOT NULL DEFAULT '' COMMENT 'CronJob 名称',
  `description` varchar(256) NOT NULL DEFAULT '' COMMENT 'CronJob 描述',
  `schedule` varchar(100) NOT NULL DEFAULT '' COMMENT 'Quartz 格式的调度时间描述',
  `status` longtext COMMENT 'CronJob 任务状态',
  `concurrency_policy` tinyint NOT NULL DEFAULT '1' COMMENT '作业处理方式（1 串行，2 并行，3 替换）',
  `suspend` tinyint NOT NULL DEFAULT '0' COMMENT '是否挂起（1 挂起，0 不挂起）',
  `job_template` longtext COMMENT 'Job 模版',
  `success_history_limit` tinyint NOT NULL DEFAULT '10' COMMENT '要保留的成功完成作业的数量。值必须是非负整数',
  `failed_history_limit` tinyint NOT NULL DEFAULT '5' COMMENT '要保留的失败完成作业的数量。值必须是非负整数。',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_scope` (`scope`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='Cron 任务表';

---

CREATE TABLE `nw_job` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  `job_id` varchar(100) NOT NULL DEFAULT '' COMMENT 'Job ID',
  `user_id` varchar(100) NOT NULL DEFAULT '' COMMENT '创建人',
  `scope` varchar(256) NOT NULL DEFAULT 'default' COMMENT 'Job 作用域',
  `name` varchar(255) NOT NULL DEFAULT '' COMMENT 'Job 名称',
  `description` varchar(256) NOT NULL DEFAULT '' COMMENT 'Job 描述',
  `cronjob_id` varchar(100) DEFAULT NULL COMMENT 'CronJob ID，可选',
  `watcher` varchar(255) NOT NULL DEFAULT '' COMMENT 'eam-nightwatch watcher 名字',
  `suspend` tinyint NOT NULL DEFAULT '0' COMMENT '是否挂起（1 挂起，0 不挂起）',
  `params` longtext COMMENT 'Job 参数',
  `results` longtext COMMENT 'Job 执行结果',
  `status` varchar(32) NOT NULL DEFAULT 'Pending' COMMENT 'Job 状态',
  `conditions` longtext COMMENT 'Job 运行状态',
  `started_at` datetime NOT NULL DEFAULT '0000-00-00 00:00:00' COMMENT 'Job 开始时间',
  `ended_at` datetime NOT NULL DEFAULT '0000-00-00 00:00:00' COMMENT 'Job 结束时间',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_task_id` (`cronjob_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_scope` (`scope`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='任务表';
