# Kubernetes 通用特性迁移评估（onex-apiserver / controller-manager / job-controller）

本文档评估 Kubernetes 仓库中可迁移到 `cmd/onex-apiserver`、`cmd/onex-controller-manager`、`cmd/onex-job-controller` 的通用特性/设计，并记录本次已落地的实现方案。

## 1. 背景与现状

这三个二进制是对 kube-apiserver / kube-controller-manager 的忠实移植，已直接依赖上游模块：

- `k8s.io/apimachinery`（类型/转换/版本/Scheme）
- `k8s.io/apiserver`（genericapiserver + 存储/注册/admission）
- `k8s.io/client-go`（informer/workqueue/leaderelection）
- `k8s.io/component-base`、`k8s.io/controller-manager`
- `sigs.k8s.io/controller-runtime`（Reconciler 运行时）

并生成了自己的 typed client/informer/lister（`pkg/generated`）。因此「迁移」不是「引入哪些包」，而是**把 k8s 中通用机制补齐到现有实现，并修复移植不到位/有缺陷之处**。

## 2. k8s 通用可迁移特性 → 现状映射

| k8s 通用能力 | 上游来源 | 现状 | 本次动作 |
|---|---|---|---|
| 类型/转换/Scheme/GVK(GVR) | `apimachinery/runtime`、`runtime/schema` | ✅ 已用 | 无 |
| ObjectMeta/ListMeta/Status 线格式 | `apimachinery/apis/meta/v1` | ✅ 已用 | 无 |
| 乐观并发 resourceVersion + CAS | `apiserver/pkg/storage` + etcd3 | ✅ 已用 | 无 |
| 通用 CRUD Store + Strategy | `apiserver/pkg/registry/generic/registry` | ✅ 已用 | 无 |
| REST 动词契约 | `apiserver/pkg/registry/rest` | ✅ 已用 | 无 |
| 错误/状态码模型 | `apimachinery/api/errors` | ✅ 已用 | 无 |
| admission 链 | `apiserver/pkg/admission` | ✅ 已用 | 无 |
| watch/cache/informer | `client-go/tools/cache` + ctrl cache | ✅ 已用 | 修复 resync 硬编码 |
| workqueue 限流重试 | `client-go/util/workqueue` | ✅ 已有 | 修复注释 |
| leader election（Lease） | `client-go/tools/leaderelection` + ctrl | ✅ 已用 | 无 |
| **ControllerExpectations（写后读一致性）** | `pkg/controller/controller_utils.go` | ❌ 缺失 | **新增（阶段 B）** |
| **backoffStore（失败指数退避）** | `pkg/controller/job/backoff_utils.go` | ❌ 缺失 | **新增（阶段 B）** |
| **ControllerRefManager（领养/弃养）** | `pkg/controller/controller_ref_manager.go` | ❌ 缺失 | **新增（阶段 B）** |
| 非 cron Job 控制器 | `pkg/controller/job` | ❌ 仅 README 存根 | **新增（阶段 B）** |
| 认证/授权（Authorizer） | `apiserver/pkg/endpoints/filters` | ✅ 已用（AlwaysAllow/Deny/RBAC） | **启用（阶段 C，含 RBAC）** |
| **Webhook admission（Validating/Mutating）** | `apiserver/pkg/admission/plugin/webhook` | ❌ 缺失 | **新增（阶段 G）** |
| **委派授权（Webhook authorizer，对接 casbin）** | `apiserver/pkg/authorization/authorizerfactory` + `plugin/pkg/authorizer/webhook` | ❌ 缺失 | **新增（阶段 G）** |
| OpenTelemetry tracing（controller 侧） | otel sdk/http | ⚠️ 仅 apiserver 有 | **接入（阶段 D）** |

## 3. 现状深度分析（9 维要点）

- **架构**：三层均采用 k8s 原生范式（genericapiserver 聚合链 / controller-runtime Reconciler + ControllerDescriptor 注册表），分层清晰。controller-manager 混用 `k8s.io/controller-manager` 注册表与 controller-runtime 运行时，属可控取舍。
- **性能**：`ResyncPeriod` 硬编码 1s 导致 informer 每秒 relist、空耗 apiserver（已修复）。
- **稳定性**：job-controller 的 healthz/readyz 名称反了（已修复）；CronJob 遗留调试输出与未完成的领养逻辑（已修复）。
- **扩展性**：reconciler 均通过 interface 注入控制面（`jobControl`/`provider`），可替换 provider 实现。
- **功能完备性**：非 cron Job 控制器、RBAC 已补齐（阶段 B/C）；调度器仍空缺。
- **代码质量/可维护/可读/简洁**：修正 rate limiter 注释与实现不符、`encodeConfig` 冗余 switch、无用常量、死代码等。

## 4. 已落地实现

### 阶段 A — 快速修复
- `cmd/onex-job-controller/app/controller.go`：修复 setupChecks 中 healthz/readyz 名称颠倒，并修正 cronjob 错误日志里的 "minerset"。
- `cmd/onex-controller-manager/app/controllermanager.go`：`ResyncPeriod` 由硬编码 `1s` 改为 `wait.Jitter(Generic.SyncPeriod.Duration, 1.0)`（默认 10h，带抖动）。
- `internal/controller/job/cronjob/controller.go` / `utils.go`：删除 18 处 `fmt.Println` 调试；实现 `getCronJobsForJob`；将 `shouldExcludeJob`/`adoptOrphan` 接入 `getJobsToBeReconciled`；删除无用 `MaxConcurrency`；移除已 GA 的 `CronJobsScheduledAnnotation` 特性门（该常量在 k8s v1.37 已删，原代码无法编译）。
- `internal/pkg/util/ratelimiter/ratelimiter.go`：修正与实现不符的注释。
- 两个 `configfile.go`：简化 `encodeConfig` 冗余 switch。

### 阶段 B — 非 cron Job 控制器（`internal/controller/job/job/`，忠实移植 k8s `pkg/controller/job` 机制）
- `expectations.go`：移植 `ControllerExpectations`/`ControlleeExpectations`（原子计数器 + TTL 过期 + `SatisfiedExpectations` 门控）。
- `backoff_utils.go`：移植 `backoffRecord.getRemainingTime` 指数退避。
- `controller_ref_manager.go`：移植 ownerRef 领养/弃养（`ClaimObject`）。
- `injection.go`：`jobControlInterface`/`providerControlInterface` + real/fake（测试缝）。
- `controller.go`：状态机 Pending→Running→Succeeded/Failed，支持 Suspend、ActiveDeadlineSeconds、Conditions 管理与 provider 抽象。
- `controller_test.go`：覆盖启动/挂起/超时/完成/退避/expectations 门控。
- 已接线到 `cmd/onex-job-controller/app/controller.go`。

### 阶段 C — 启用 authn/authz
- 新增 `internal/controlplane/apiserver/authorizer.go`：`BuildAuthorizer` 支持 `AlwaysAllow`（默认）/`AlwaysDeny`/`RBAC`（复刻 `k8s.io/kubernetes/plugin/pkg/auth/authorizer/rbac` + RBAC informer），`system:masters` 无条件放行，本地运行、无需委派 k8s 集群。
- `internal/controlplane/apiserver/options/options.go`：新增 `--authorization-mode` flag（默认 `AlwaysAllow`）。
- `internal/controlplane/apiserver/config.go`：移除过期的注释块，接入 `BuildAuthorizer`，覆盖 `genericConfig.Authorization.Authorizer`/`RuleResolver`。认证仍由 `RecommendedOptions.ApplyTo` 用 client-cert/token + 匿名回退完成。

### 阶段 D — controller 接入 OTel tracing
- 新增 `internal/pkg/util/tracing/tracing.go`：`Setup` 初始化全局 TracerProvider 并返回 `otelhttp` transport 包装器。
- `onex-controller-manager` / `onex-job-controller` 的 `Run` 中，按环境变量（`ONEX_OTEL_ENDPOINT` / `ONEX_OTEL_SERVICE_NAME`）开关：设置后对 `Kubeconfig` 的 transport 注入 span；未设置则默认关闭。`go.mod` 已将 `otel/sdk`、`otelhttp`、`otlptracegrpc` 提升为直接依赖。

### 阶段 E — 仓库健康 + 基础设施主干
- 修复 `internal/controller` 既有编译错误（`alias.go` 重复 import、`bak/` 备份目录），使 `go build ./...` 全绿。
- 统一 clock 注入（`job`/`cronjob` controller 用 `k8s.io/utils/clock`），消除测试中的 `time.Sleep` 依赖。
- 对齐 `.go-version`（1.23.3 → 1.26.0）、`go.mod` pin 依赖、clientset 生成代码与新 k8s v1.37 接口。

### 阶段 F — Job provider 语义补齐
- `internal/controller/job/job/provider.go`：实现真实 `kubernetes` provider（单 Pod 语义），`ProviderSpec.Raw` 承载 JSON 编码的 `corev1.PodSpec`，`Start`/`Stop`/`Status`/`Children` 管理子 Pod。
- `injection.go` 的 `realProviderControl` 改为按 `spec.type` 派发；其它 type 返回「不支持」错误。
- ControllerRef adopt/release、expectations 门控、backoff 退避作用于该 Pod 子对象（`controller_test.go` / `provider_test.go` 覆盖）。

### 阶段 G — apiserver 扩展机制
- **Webhook admission**：`options/plugins.go` 注册 `MutatingAdmissionWebhook` / `ValidatingAdmissionWebhook`（默认关闭，`--enable-admission-plugins` 开启）；`config.go` 接入标准 apiserver initializer（外部 informer/clientset）+ webhook initializer（`serviceResolver`=ClusterIP+loopback、`authInfoResolverWrapper`）。
- **委派授权（casbin webhook）**：`authorizer.go` 的 `BuildAuthorizer` 新增 `Webhook` 模式，通过 `authorizerfactory.DelegatingAuthorizerConfig` 委派到远端 `SubjectAccessReview` 服务（如 casbin）；新增 `--authorization-webhook-config-file` / `--authorization-webhook-cache-authorized-ttl` / `--authorization-webhook-cache-unauthorized-ttl` flag。

## 5. 假设与待办

- **Job 控制器的 provider 子对象**：`Job.Spec.Type == "kubernetes"`（或空，默认 kubernetes）时，已实现 `kubernetesProvider`——单 Pod 语义，`ProviderSpec.Raw` 承载 JSON 编码的 `corev1.PodSpec`，`Start` 建 Pod、`Status` 映射 Pod phase、`Children`/`Stop` 按 label 管理子对象；`realProviderControl` 按 `spec.type` 派发，其它 type 返回「不支持」错误，待接入对应后端。`ControllerExpectations`/`ControllerRefManager` 作用于该 Pod 子对象。
- **RBAC**：`BuildAuthorizer` 已实现 RBAC authorizer（`rbac.New` + RBAC informer），并新增 `cmd/onex-controller-manager/app/rbac.go`（clusterrole-aggregation 控制器）。
- **Webhook admission / 委派授权**：webhook admission 默认关闭（需 `--enable-admission-plugins`），其 `service:` 引用经 ClusterIP resolver 解析、`url:` 引用直连；`--authorization-mode=Webhook` 配合 `--authorization-webhook-config-file` 指向 casbin 等远端 `SubjectAccessReview` 服务。二者均复用 `InternalVersionedInformers`（post-start hook 已启动），无需额外 informer 生命周期管理。
- **tracing 配置**：当前用环境变量开关（OTel 社区惯用方式），未下沉为 component-config 字段（避免触发 versioned config 的 zz_generated 重构）。

## 6. 验证

- `go build ./cmd/onex-apiserver/... ./cmd/onex-controller-manager/... ./cmd/onex-job-controller/...` ✅
- `go build ./...` ✅（全仓绿）
- `go vet ./internal/controlplane/... ./internal/controller/job/job/... ./internal/pkg/util/tracing/...` ✅
- `go test ./internal/controlplane/apiserver/...` ✅（含 RBAC、AlwaysAllow/Deny、Webhook 委派授权 e2e 用例）
- `go test ./internal/controller/job/job/...` ✅（controller + provider 用例）
- `go mod tidy` 产出最小 diff（OTel 依赖提升为 direct）✅

> 注：`internal/controller/apis/config/v1beta1/bak`、`internal/controller/apis/config/validation`、`internal/controller/alias.go` 在本次改动前就已存在编译错误（备份目录/重复 import 等），是仓库既有问题，已在阶段 E 修复并纳入范围。