```bash
├── .air.toml # air 工具配置，https://github.com/cosmtrek/air
├── api/ # OpenAPI/Swagger 规范，JSON 模式文件，协议定义文件
│   ├── openapi/ # OpenAPI Spec 文件
│   │   ├── gateway/ # gateway 组件的 OpenAPI 文件
│   │   │   └── v1/ # v1 版本
│   │   │       ├── errors.swagger.json # 错误码定义
│   │   │       └── gateway.swagger.json # 接口定义
│   │   ├── openapi.yaml # OpenAPI 文档聚合
│   │   ├── ... # 其他组件的 OpenAPI 定义，文档结构和内容同 gateway
│   └── swagger/ # Swagger Spec 文件
│       └── swagger.yaml
├── manifests/ # 用于存储配置文件、部署清单或其他描述项目配置和部署信息的文件。例如：Dockerfile、Kubernetes部署清单、配置文件模板等
│   ├── chain.yaml # Chain 模板 YAML
│   ├── genesis/ # 创建创世链的 K8S 定义文件
│   │   ├── genesis.yaml
│   │   └── node-1.yaml
│   ├── template/ # 一些部署模板文件
│   ├── installation/ # 组件部署相关的配置文件、资源定义文件
│   │   ├── kubernetes/ # 存放安装 kind 集群需要的配置文件
│   │   │   └── kind-onex.yaml
│   │   └── traefik/ # 存放安装 traefik 需要的 K8S 资源定义文件
│   │       ├── depoyed-deployment.yaml # traefik 部署之后的完整 deployment 文件，主要用来作参数对比
│   │       ├── traefik-values.yaml # traefik helm value 文件
│   │       ├── whoami-ingress.yaml # traefik 部署后的测试程序的 YAML 定义
│   │       ├── whoami-services.yaml
│   │       └── whoami.yaml
│   ├── minerset.yaml # MinerSet 模板 YAML
│   ├── miner.yaml # Miner 模板 YAML
│   └── namespace.yaml # Namespace 模板 YAML
├── build/ # 存放构建相关的文件
│   ├── ci/ # 存放 CI(travis, circle, drone) 配置和脚本
│   ├── docker/ # 存放 OneX 项目各组件的的 Dockerfile 文件
│   │   ├── onex-apiserver/  # onex-apiserver 组件的 Dockerfile 存放目录
│   │   │   ├── Dockerfile # 单阶段构建 Dockerfile 文件
│   │   │   └── Dockerfile.multistage # 多阶段构建 Dockerfile 文件
│   │   ├── ... # 其他组件 Dockerfile 文件，文件结构、内容同 onex-apiserver
│   ├── package/ # 存放云( AMI )、容器( Docker )、操作系统( deb、rpm、pkg )包配置和脚本
│   └── tools.go
├── CHANGELOG/ # 目录，用来存放记录变更历史的文件
├── .chglog/ # chglog 工具配置文件所在的目录
│   ├── CHANGELOG.tpl.md*
│   └── config.yml*
├── cmd/ # OneX 各组件 main 文件的存放目录
│   ├── gen-docs/ # gen-docs 用来生成 onexctl（单独更新） 命令介绍文档，存放于 docs/guide~/en-US/cmd/onexctl 目录中
│   ├── gen-man/ # gen-man 用来生成 onex 项目各组件的 man 文件，存放于 docs/man/man1目录中
│   ├── gen-swagger-type-docs/ # 用来生成 swagger 文档
│   ├── gen-yaml/ # 用来生成 onexctl 目录中各子命令的参数描述文件（YAML格式）
│   ├── gen-onex-docs/ # gen-onex-docs 用来生成 onex 各组件（包括onexctl）的命令行参数描述文档，存放于 docs/guide~/en-US/cmd 目录中
│   ├── gen-onex-gorm-model/ # 用来生成 gorm model 文件
│   ├── lint-kubelistcheck/ # linter，用来检查是否设置 ResourceVersion 为 0
│   ├── README.md # cmd 目录的 README 文件。onex 项目中会根据需要，在一些目录中存放 README.md 文件，用来介绍所在目录
│   ├── onex/ # onex 工具，onex 项目脚手架工具
│   ├── onex-ai/ # onex 项目 AI 组件
│   │   └── README.md
│   ├── onex-apiserver/ # onex-apiserver 组件目录
│   │   ├── apiserver.go # onex-apiserver main 函数入口
│   │   ├── app/ # onex-apiserver 启动核心逻辑和 Flag
│   │   │   ├── options/ # onex-apiserver 组件命令行 Flag 配置和构建
│   │   │   └── server.go # onex-apiserver 启动核心逻辑所在的文件
│   ├── ... # 其他 onex 组件，结构和内容类似于 onex-apiserver
├── CODE_OF_CONDUCT.md
├── configs/ # 存储项目的配置文件
│   ├── gen.yaml # gen-onex-gorm-model 组件配置文件
│   ├── kafkactl.config # kafkactl 工具配置文件
│   ├── rbac_model.conf # casbin rbac model 配置文件
│   ├── rbac_policy.csv # casbin policy 文件
│   ├── onex-controller-manager.yaml # onex-controller-manager 配置文件
│   ├── onex-create.sql # onex 数据库创建 SQL 语句（记录，方便记忆）
│   ├── onex-gateway.yaml # onex-gateway 配置文件
│   ├── onex-miner-controller.yaml # onex-miner-controller 配置文件
│   ├── onex-minerset-controller.yaml # onex-minerset-controller 配置文件
│   ├── onex-nightwatch.yaml # onex-nightwatch 配置文件
│   ├── onex.sql # onex 数据库创建时导入文件（使用source命令导入）
│   ├── onex-toyblc.yaml # onex-toyblc 配置文件
│   └── onex-usercenter.yaml # onex-usercenter 配置文件
├── CONTRIBUTING.md # 介绍如何给onex项目做贡献
├── docs/ # 项目文档
│   ├── devel/ # 开发文档
│   │   ├── en-US/ # 英文版文档
│   │   │   └── contributors/ # 贡献文档
│   │   │       ├── github-workflow.md # onex 项目 github 工作流
│   │   │       └── git_workflow.png
│   │   └── zh-CN/ # 中文文档
│   │       ├── architecture.md # onex 项目架构介绍文档
│   │       ├── blockchain-installation.md
│   │       ├── code_structure.md
│   │       ├── code_structure.md~
│   │       ├── conversions/ # onex 项目开发规范
│   │       ├── development.md # 介绍如何开发 onex 项目
│   │       ├── feature-design.md # 介绍 onex 项目的功能
│   │       ├── future.md # 介绍 onex 项目待实现的功能
│   │       └── notes.md # 开发笔记
│   ├── .generated_docs
│   ├── guide/ # onex 组件使用文档
│   │   ├── en-US/ # 英文版文档
│   │   │   ├── cmd/
│   │   │   │   ├── onex-apiserver.md # onex-apiserver 组件使用文档
│   │   │   │   ├── ... # 其他组件使用文档
│   │   │   └── yaml/
│   │   │       └── onexctl/ # onexctl 命令参数 YAML 描述
│   │   └── zh-CN/ # 中文文档
│   │       ├── api/ # API 文档
│   │       ├── best-practice/ # 最佳实践
│   │       ├── faq/ # 常见文档
│   │       ├── installation/ # 部署指南
│   │       ├── introduction/ # 项目介绍
│   │       ├── operation-guide/ # 操作指南
│   │       ├── quickstart/ # 快速入门
│   │       ├── sdk/ # OneX Go SDK
│   │       └── tools/ # 工具使用指南
│   ├── images/ # OneX 项目图片存放目录
│   ├── man/ # OneX 项目 man1文件存放目录
│   ├── README.md
│   └── SUMMARY.md
├── examples/ # 应用程序或公共库的示例
├── .git/ # Git 元数据目录
│   ├── COMMIT_EDITMSG # 包含当前正在进行的提交的消息
│   ├── config # 存储了 Git 仓库的配置信息
│   ├── FETCH_HEAD # 用于跟踪上次执行的git fetch命令的结果
│   ├── HEAD # 指向当前所在的分支或提交
│   ├── hooks/ # 包含了 Git 钩子脚本，用于在特定事件发生时执行自定义操作
│   │   ├── commit-msg* # 在提交消息时触发的钩子
│   │   ├── pre-commit* # 在执行提交之前触发的钩子
│   │   ├── pre-push* # 在执行推送之前触发的钩子
│   ├── index # 包含了当前暂存区的内容
│   ├── logs/ # 存储了引用日志，记录了分支和 HEAD 的历史
│   │   ├── HEAD # 
│   │   └── refs/
│   │       ├── heads/
│   │       │   ├── develop
│   │       │   ├── feature/
│   │       │   │   └── refactor
│   │       │   └── master
│   │       ├── remotes/
│   │       │   └── origin/
│   │       │       ├── develop
│   │       │       └── master
│   │       └── stash
│   ├── objects/ # 存储了 Git 对象，包括文件内容和版本历史
│   │   ├── info/
│   │   └── pack/
│   ├── ORIG_HEAD # 在进行某些操作（如合并）时保存了原始的 HEAD 的引用
│   ├── refs/ # 存储了分支、标签等引用的指针
│   │   ├── heads/ # 存储了本地分支的引用
│   │   │   ├── develop
│   │   │   ├── feature/
│   │   │   │   └── refactor
│   │   │   └── master
│   │   ├── remotes/ # 存储了远程分支的引用
│   │   │   └── origin/
│   │   │       ├── develop
│   │   │       └── master
│   │   ├── stash # 存储了暂存的更改
│   │   └── tags/ # 存储了标签的引用
│   │       └── v0.0.1
│   └── test # 用于存储 Git 测试相关的文件
├── .gitattributes
├── githooks/ # githooks，纳入代码版本库管理
├── .github/ # GitHub 配置文件
│   ├── dependabot.yml # 用于配置 Dependabot 自动更新依赖的文件
│   ├── ISSUE_TEMPLATE/ # 包含用于创建问题模板的文件
│   │   ├── bug-report.md # 用于报告问题的模板
│   │   ├── enhancement.md # 用于提出改进建议的模板
│   │   └── workflow.md # 用于描述工作流程的模板
│   ├── OWNERS # 指定了代码库的所有者或相关团队的文件
│   ├── PULL_REQUEST_TEMPLATE.md # 用于创建拉取请求模板的文件
│   └── SECURITY.md # 包含有关代码库安全性的信息和指导的文件
├── .gitignore # 指定了 Git 应该忽略的文件和目录，这些文件和目录不会被 Git 跟踪g提交
├── .gitlint # 包含了 GitLint 工具的配置文件，用于定义代码库中的提交消息规范
├── .golangci.yaml # 包含了 GolangCI-Lint 工具的配置文件，用于定义代码库中的静态代码分析规则
├── go.mod # Go 项目的模块文件，用于定义项目的依赖关系和版本约束
├── go.sum # 包含了 Go 项目的依赖项的版本和哈希值，用于确保依赖项的完整性
├── .gsemver.yaml # 包含了 GSemver 工具的配置文件，用于定义 Go 项目的语义化版本规范
├── init/ # 系统初始化（systemd、upstart、sysv）和进程管理（runit、supervisord）配置
├── internal/ # 该目录用于存放项目的内部代码，这些代码只在项目内部使用，不对外暴露
│   ├── admission/ # 包含了与准入控制相关的代码
│   ├── apiserver/ # 包含了与 API 服务器相关的代码
│   ├── blc/ # 包含了与区块链相关的代码
│   ├── controller/ # 包含了控制器相关的代码
│   │   ├── alias.go # 定义了控制器的别名
│   │   ├── apis/ # 包含了与控制器相关的 API 定义
│   │   │   └── config/ # 包含了与配置相关的 API 定义
│   │   ├── chain/ # 包含了与链相关的控制器代码
│   │   ├── controller_utils.go # 包含了控制器的工具函数
│   │   ├── miner/ # 存放了 miner controller 的实现
│   │   │   ├── apis/ # miner API 定义
│   │   │   │   └── config/ # miner 配置定义
│   │   │   ├── controller.go # miner controller 实现
│   │   │   ├── ...
│   │   ├── minerset/ # 存放了 minerset controller 的实现
│   │   │   ├── apis/
│   │   │   │   └── config/ # minerset 配置定义
│   │   │   ├── controller.go # minerset controller 实现
│   │   │   └── ...
│   │   ├── namespace/ # 存放了 namespace controller 的实现
│   │   │   ├── controller.go # namespace controller 实现
│   │   │   └── ...
│   │   └── sync/ # 包含了一些同步相关的controller
│   │       ├── chain_sync.go # chain 同步控制器
│   │       ├── minerset_sync.go # minerset 同步控制器
│   │       └── miner_sync.go # miner 同步控制器
│   ├── demo/ # demo 程序实现目录
│   ├── gateway/ # 包含了 gateway 的实现代码
│   │   ├── biz/ # 包含了业务逻辑相关的代码
│   │   ├── converter/ # 包含了数据转换相关的代码
│   │   ├── errors/ # 定义了一些通用的错误
│   │   ├── locales/ # 包含了本地化相关的代码
│   │   ├── model/ # 包含了模型相关的代码
│   │   ├── service/ # 包含了服务相关的代码
│   │   ├── store/ # 包含了数据存储相关的代码
│   │   ├── validation/ # 包含了数据验证相关的代码
│   ├── lint/ # 各个 linter 的实现
│   ├── nightwatch/ # 包含了 nightwatch 的实现
│   ├── pkg/ # 包含了项目的公共代码，这些代码可以被项目中的其他模块或服务所共享
│   │   ├── bootstrap/ # 包含了启动相关的代码
│   │   ├── client/ # 包含了客户端相关的代码
│   │   ├── config/ #  配置相关
│   │   ├── contract/ # contract 相关
│   │   ├── core/ # 包含了一些核心的函数、方法等
│   │   ├── feature/ # feature gate
│   │   ├── global/ # 全局变量等
│   │   ├── idempotent/ # 实现幂等
│   │   ├── known/ # 包含了已知的代码，如已知的标签、标注等
│   │   │   ├── annotations.go # 包含了已知的标注常量定义
│   │   │   ├── apiserver/ # 包含了 apiserver 相关的预定义常量
│   │   │   │   └── default.go # 常量定义默认存放文件
│   │   │   ├── known.go # 常量定义默认存放文件
│   │   │   ├── labels.go # 包含了已知的标签常量定义
│   │   │   └── usercenter/ 包含了 usercenter 相关的预定义常量
│   │   ├── meta/ # 内部使用的一些元数据（数据结构等）
│   │   ├── metrics/ # 包含了度量指标相关的代码，用于收集和展示项目的性能指标
│   │   ├── middleware/ # 包含了中间件相关的代码，用于处理请求的前置和后置逻辑
│   │   │   ├── auth/ # 认证、授权
│   │   │   ├── authn/ # 认证
│   │   │   ├── i18n/ # 国际化
│   │   │   ├── idempotent/ # 请求幂等
│   │   │   ├── logging/ # 日志
│   │   │   ├── tracing/ # 调用链追踪
│   │   │   └── validate/ # 验证
│   │   ├── nextid/ # 包含了生成唯一标识符的代码，用于生成下一个可用的 ID
│   │   ├── options/ # 包含了选项相关的代码，用于设置和管理项目的选项
│   │   ├── ports/ # 预定义端口
│   │   ├── printers/ # onex-apiserver 命令行输出相关代码
│   │   ├── util/ # 工具包，里面包含了一些utils
│   │   ├── validation/ # 校验
│   │   ├── onexx/ # onex 项目的 context 定义
│   │   └── zid/ # onex id generator
│   ├── pump/ # pump 组件的实现
│   ├── registry/ # onex-apiserver 的 registry 实现
│   ├── toyblc/ # toyblc 的实现
│   ├── usercenter/ # 包含了 usercenter 组件的代码实现，结构及目录功能类似于 gateway
│   └── onexctl/ # 包含了 onex 项目命令行工具 onexctl 的实现
├── .kube-linter.yaml # kube-linter 工具配置文件
├── LICENSE # 版权声明文件
├── .make/ # Makefile 临时文件存储目录
├── Makefile # 项目主 Makefile
├── _output/ # 构建产物存放目录
│   ├── bin/
│   ├── cert/ # 生成的 CA 证书存放目录
│   ├── config # onex-apiserver 类 kubeconfig 文件
│   ├── platforms/
│   │   └── linux/
│   │       └── amd64/ # onex 项目各组件编译后的二进制文件存放目录，amd64 CPU 架构
│   ├── tmp/ # 临时目录
│   └── tools/ # 工具类
├── pkg/
│   ├── api/
│   │   ├── gateway/
│   │   │   └── v1/ # gateway v1版 protobuf 文件（接口、错误）
│   │   ├── toyblc/
│   │   │   └── v1/ # toyblc v1版 protobuf 文件（接口、错误）
│   │   ├── usercenter/
│   │   │   └── v1/ # usercenter v1版 protobuf 文件（接口、错误）
│   │   └── zerrors/ # onex 项目公共错误定义，protobuf 文件
│   ├── apis/ # kube-apiserver style 的 scheme 文件
│   │   ├── apps/ # group apps
│   │   │   ├── v1beta1/ # 版本 v1beta1
│   │   │   ├── validation/ # 校验库
│   │   ├── coordination/ # group coordination
│   │   │   ├── v1/ # 版本 v1
│   │   │   ├── validation/ # 校验库
│   │   ├── core/ # group core
│   ├── app/ # app包，应用构建框架，用来快速构建一个应用
│   ├── authn/ # 用来实现认证功能的包
│   ├── cli/ # CLI 包
│   ├── config/ # Kube-Configuration 包
│   ├── db/ # 创建 MySQL、Redis 的包
│   ├── encoding/ # 编解码相关的包
│   │   └── json/ # JSON 编解码器
│   ├── errors/ # 预定义错误类型
│   ├── generated/ # K8S 生成文件存放的目录
│   │   ├── clientset/ # client-go
│   │   ├── informers/ # informers
│   │   ├── listers/ # listers
│   │   └── openapi/ # openapi
│   ├── i18n/ # 国际化包
│   ├── id/ # ID 生成包
│   ├── idempotent/ # 请求幂等包
│   ├── log/ # OneX 项目内部日志包
│   ├── options/ # 存放各类 Options 定义文件
│   ├── streams/ # 数据清洗公共包
│   ├── util/ # 工具目录，里面存放了很多 utils
│   │   ├── file/ # 实现文件类操作
│   │   ├── gen/ # 实现代码生成类操作
│   │   ├── ip/ # 实现 IP 操作
│   │   ├── leaderelection/ # 实现选举操作
│   │   ├── lint/ # 实现静态代码检查
│   │   ├── pagination/ # 实现翻页功能
│   │   ├── record/ # Event Recoder
│   │   ├── reflect/ # 实现反射相关功能
│   │   ├── retry/ # 实现重试
│   │   ├── strings/ # 字符串相关功能
│   │   └── version/ # 实现代码版本功能
│   ├── version/
├── README-en.md # 项目英文版 README
├── README.md # 项目中文版 README
├── scripts/ # 脚本文件、makefile脚本
│   ├── ...
│   ├── make-rules/ # Sub-Makefile
├── SECURITY_CONTACTS
├── test/ # 外部测试应用程序和测试数据。随时根据需要构建/test目录。对于较大的项目，有一个数据子目录更好一些
├── third_party/ # 外部辅助工具，fork的代码和其他第三方工具（例如Swagger UI）
├── tools/ # 外部测试应用程序和测试数据。随时根据需要构建/test目录。
└── web/ # Web应用程序特定的组件：静态Web资源，服务器端模板和单页应用
```
