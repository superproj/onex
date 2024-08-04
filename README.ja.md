<p align="center" style="text-align: center">
  <img src="./docs/images/logo.png" width="55%"><br/>
</p>

<p align="center">
  零云实战平台，做最专业的实战项目。
  <br/>
  <br/>
  <a href="https://github.com/superproj/onex/blob/master/LICENSE">
    <img alt="GitHub" src="https://img.shields.io/github/license/superproj/onex"/>
  </a>
  <a href="https://goreportcard.com/report/github.com/superproj/onex">
    <img src="https://goreportcard.com/badge/github.com/superproj/onex" />
  </a>
  <a href="https://pkg.go.dev/github.com/superproj/onex">
    <img src="https://pkg.go.dev/badge/github.com/superproj/onex.svg" alt="Go Reference"/>
  </a>
  <a href="https://github.com/superproj/onex/issues">
    <img src="https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat" alt="CodeFactor" />
  </a>
  <a href="https://app.fossa.com/projects/git%2Bgithub.com%2Fsuperproj%2Fonex?ref=badge_shield" alt="FOSSA Status">
    <img src="https://app.fossa.com/api/projects/git%2Bgithub.com%2Fsuperproj%2Fonex.svg?type=shield" />
  </a>
  <a href="https://github.com/avelino/awesome-go" rel="nofollow">
    <img src="https://cdn.rawgit.com/sindresorhus/awesome/d7305f38d29fed78fa85652e3a63e154dd8e8829/media/badge.svg" alt="Awesome" />
  </a>
  <a href="https://discord.gg/BrRSWTaxVK">
    <img alt="Discord" src="https://dcbadge.vercel.app/api/server/BrRSWTaxVK?style=flat"/>
  </a>
  <br/>
  <a href="https://github.com/superproj/onex/actions/workflows/build-and-test.yml" rel="nofollow">
    <img src="https://img.shields.io/github/actions/workflow/status/superproj/onex/build-and-test.yml?branch=master&logo=Github" alt="Build" />
  </a>
  <a href="https://github.com/superproj/onex/tags" rel="nofollow">
    <img alt="GitHub tag (latest SemVer pre-release)" src="https://img.shields.io/github/v/tag/superproj/onex?include_prereleases&label=version"/>
  </a>
</p>

<div align="center">
<strong>
<samp>

[简体中文](README.md) · [English](README.en.md) · [日本語](README.ja.md) · [한국어](README.ko.md)

</samp>
</strong>
</div>
# OneX

OneX是一个微型的矿机云平台，也是一个优秀的 Go 企业应用开发脚手架，遵循简洁架构，具有 2 大编程范式，代码规范、质量高、功能全。此外onex还具有以下特性：

- 一个包含Kubernetes编程教学的实战项目；
- 一个微型的Kubernetes；
- 一个符合Kubernetes编程哲学的实战项目；
- 一个可用于企业级生产环境的实战项目；
- 一个支持超高并发的实战项目；
- 一个微服务实战项目

## Features

onex设计目的是能够在一个项目中，有机整合 Go 项目开发中用到的几乎所有的核心功能点，以及大部分常用功能点，具体功能点如下：

- 代码规范、质量高：
  - 遵循各种代码开发规范：代码规范、目录规范、日志规范、错误码规范、提交规范、版本规范、文档规范等
  - 采用标准、高效的代码开发流程：
    - 采用结构化的高质量 Makefile 高效管理 Git 大仓；工程化管理包括但不限于以下几点：
      - 代码格式化
      - 代码生成
      - 代码编译
      - 代码测试
      - 代码部署
      - 镜像制作
      - 版权声明
      - 生成Swagger文档
      - 工具安装
      - CA证书制作
      - 静态代码检查
      - license header添加
      - ...
    - 采用GitFlow代码管理模式
    - 采用敏捷开发模式
    - 集成进了CI/CD系统
  - 面向接口编程、Go项目设计哲学
  - 可测试的代码设计
  - 代码遵循了 Go最佳实战，是用来大量的Go设计模式
- 遵循简洁架构：onex-gateway/onex-usercenter 遵循了简洁架构。(service、biz、store三层设计)；
- 常用功能设计：
  - 日志包设计
  - 错误包设计
  - 错误码设计
  - 应用框架构建设计
  - ...
- 包含了众多Linter检查规则：Dockerfile, Helm Chart, Code, linters
- 项目包含了声明式 API 和命令式 API 2 种编程范式；
- Kubernetes编程开发实战：
  - 声明式 API 编程范式（Kubernetes 编程）：
    - client-go 编程
    - Aggregated APIServer
    - kube-apiserver 风格的 apiserver： onex-apiserver；
    - Kubernetes controller（CRD+Controller）：onex-minerset-controller, onex-miner-controller, onex-controller-manager；
    - Operator开发：onex-operator
  - Kubernetes Webhook、准入控制；
  - Kubernetes 集群（Kind集群）搭建和使用；
- 开发阶段全：项目包含了设计、开发、测试、发布全流程
  - 测试：单元测试、性能测试、性能分析、测试框架(testify, GoConvey)、覆盖率、Mock工具(sqlmock, httpmock, bouk/monkey, golang/mock)
- 项目有配套的课程可供学习：极客时间、知识星球、B站免费课程（开发企业级REST API服务）、慕课网（视频课）；
- 采用结构化的高质量 Makefile 高效管理 Git 大仓；
- 代码实现的企业级功能全： 
  - 包含了企业级应用需要的众多核心功能：
    - Web服务(onex-gateway, onex-usercenter)详细功能列表包含但不限于以下列表：
      - RESTful
      - gRPC
      - 路由匹配
      - 路由分组
      - HTTP/HTTPS
      - 认证/授权
      - 路由匹配
      - Protobuf
      - 中间件
      - 跨域
      - ResutID
      - 优雅关停
      - 参数解析
      - 参数校验
      - 逻辑处理
      - 返回结果处理
      - ...
    - 分布式异步任务处理服务：onex-nightwatch；
    - ETL数据抽取：onex-pump
    - HTTP/GRPC SDK
  - 包含了其他辅助功能：
    - 命令行工具：onexctl；
    - 代码检查：lint-xxx
    - 代码生成：gen-xxx
- 自动版本生成、CHANGELOG生成
- 开发中一些常用功能设计：Client SDK设计方案，Options设计方案
- 大量采用了代码生成技术，提高代码开发效率（见hack/make-rules/generate.mk）：
  - 自定义代码生成gen-xxx
  - 自动生成go protobuf文件 
  - swagger.json
  - error code
  - doc.go
  - ca
  - license header
  - 应用使用文档
  - man文件
  - wire 依赖注入
  - k8s相关源码（listers, informers, client-go等）
  - dockerfile 
  - kubeconfig
  - helm docs
  - ...
- 中间件组件及实战：MySQL、Redis、Etcd、Kafka、MongoDB；
- 校验：基于 Tag 的校验、自定义校验逻辑
- 日志(v1.0)：Elasticsearch、Filebeat、Kibana；(Loki，grafana v2.0)
- 监控告警：Prometheus、alertmanager；
- 部署：裸机部署、独立Deployment部署、Helm编排部署；
- 错误码设计
- Event 分库 Etcd
- 认证：基本认证、Token认证、CA
- 授权：RBAC
- 用了大量Go生态中优秀的包：
  - 缓存：go-cache、lru；
  - 中间件：go-redis、gorm；
  - Web框架：grpc、gin(独立Demo)、grpc-gateway；
  - 命令行工具：cobra、pflag、viper；
  - 认证：golang-jwt；
  - 校验：validator；
  - 日志：klog、zap(独立Demo)、logrus(独立Demo)；
  - 定时任务：cron；
  - HTTP客户端：retry；
  - 授权：casbin
  - 微服务框架：kratos
  - 其他常用的包：client-go、uuid、golang-set、segmentio/kafka-go、fatih/color、olekukonko/tablewriter、redsync等；
- 微服务相关功能：
  - 配置中心：viper + configmap、【kubernetes、nacos（v2.0）】
  - 服务治理（服务注册/服务发现）：etcd(v1.0)、、polaris(v2.0)
  - 调用链：OpenTelemetry、Jaeger
- Blockchain 基本原理、私有链搭建和使用；
- 验证码（captcha）
- 其他特性：
  - Protobuf；
  - 限流算法；
  - 依赖注入；
  - 分布式锁；
  - CA证书制作
  - Dockerfile编写
- 云原生应用的部署设计和实战：
  - 云原生部署架构设计
  - 监控告警
  - 分布式日志解决方案
  - 容灾能力建设实战
  - 弹性扩缩容能力设计和实战
  - Helm服务编排
  - 容器化部署
  - 安全能力建设
  - CI/CD
  - 负载均衡
- 更多其他特性...
- 可能实现的功能：
  - WebSocket；
  - Kratos集成Gin；
  - 服务治理用：consul、kubernetes
- 涉及的语言：Go、Shell、Makefile、AWK
- 使用了众多可以提高效率、规范的工具，例如：
  - gsemver, git-chglog, addlicense, kratos, kind, go-apidiff, gotests, cfssl, go-gitlint, kustomize, kafkactl, kube-linter, kubectl, helm-docs, db2struct, gentool, air, swagger, license, helm, kafka, golangci-lint, goimports, wire
- 泛型、模糊测试、多工作区
- I18n - 国际化支持, 简单切换多语言
- Idempotent - 接口幂等性(解决重复点击或提交)
- Redis - 缓存, 内置防缓存穿透/缓存击穿/缓存雪崩示例
- Action - 权限, 基于行为的权限校验
- Proto - proto协议同时开启gRPC & HTTP支持, 只需开发一次接口, 不用写两套
- Swagger - Api文档一键生成, 无需在代码里写注解
- Embed - go 1.16文件嵌入属性, 轻松将静态文件打包到编译后的二进制应用中
- SqlMigrate - 数据库迁移工具, 每次更新平滑迁移
- Asynq - 分布式定时任务(异步任务)
- Gitbook
- 白名单
- 其他内容：
  - Kubernetes集群（Kind集群）搭建和使用
  - Blockchain基本原理、私有链搭建和使用

## Architecture

![架构图](./docs/images/superporj-arch.png)

## Installation

安装步骤如下：

```bash
$ git clone https://github.com/superproj/onex.git

$ cd onex

# 添加缺失的Go包
$ go mod tidy

# 安装依赖的工具
$ make ci

# 生成代码、编译代码等
$ make

# 本地快速部署
$ make deploy
```
    
## Usage/Examples

```javascript
import Component from 'my-project'

function App() {
  return <Component />
}
```

## Documentation

[Documentation](https://linktodocumentation)


## Feedback

If you have any feedback, please reach out to us at colin404@foxmail.com


## Contributing

Contributions are always welcome!

See `CONTRIBUTING.md` for ways to get started.

Please adhere to this project's `code of conduct`.


## Authors

- [@孔令飞](https://www.github.com/colin404)
- 微信群，公众号

## License

[MIT](https://choosealicense.com/licenses/mit/)


## Related

Here are some related projects

- [iam: 企业级的 Go 语言实战项目：认证和授权系统（带配套课程）](https://github.com/superproj/iam)
- [B 站 Go 语言开发免费课程]()
