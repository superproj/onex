# MySQL 常用操作

## `mysqldump` 命令使用指南

常用参数介绍：

- `--no-data`: 只导出表结构不导出数据
- `--routines`: 导出存储过程和自定义函数

### 1. 导出所有数据库

```bash
mysqldump -uroot -proot --databases onex > /tmp/onex.sql
```

### 2. 导出 `onex` 数据库的所有数据

```bash
mysqldump -uroot -proot --databases onex > /tmp/onex.sql
```

### 3. 导出初始化 `onex数据库的 SQL 语句

```bash
mysqldump -hxxx.xx.xx.xxx -uonex --databases onex -p'onex(#)666' onex --add-drop-database --add-drop-table --add-drop-trigger --add-locks --no-data > /tmp/onex.sql
```

## 登录 MySQL

```bash
mysql -h127.0.0.1 -uonex -p'onex(#)666' -D onex
```


## 创建用户并授权

1. 授权给指定 IP

```sql
grant all on onex.* TO 'onex'@'localhost' identified by 'onex(#)666' with grant option;
flush privileges;
```

2. 授权给所有 IP

```sql
grant all on onex.* TO 'onex'@'%' identified by 'onex(#)666' with grant option;
flush privileges;
```

## 确认用户“onex”已经被授予访问 MySQL 服务器的权限

```sql
show grants for onex;
```

## 删除用户

```sql
drop user onex;
```

## 创建数据库和表

```sql
CREATE DATABASE  IF NOT EXISTS `superproj`;
CREATE TABLE `user` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `username` varchar(253) DEFAULT NULL,
  `nickname` varchar(253) NOT NULL,
  `password` varchar(64) NOT NULL,
  `email` varchar(253) NOT NULL,
  `phone` int(20) NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
  `updated_at` timestamp NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_username` (`username`)
) ENGINE=InnoDB AUTO_INCREMENT=91 DEFAULT CHARSET=utf8;
```

## 修改 MySQL `root` 密码

```bash
mysql -uroot -p
```

## 在某列之后添加一列

```sql
alter table `miner` add column `displayName` varchar(253) not null after `name`;

```
