# OneX 核心组件测试用例指南（decl-test2）

> 本文档是 `decl-test.md` 的**超集**：在完整保留其「重试/退避机制 + 重试用例 + 资源操作示例」的基础上，进一步补充了 **CRUD / 字段校验 / status & scale 子资源 / 乐观锁 / watch / 健康检查 / Leader 选举 / 控制器开关** 等常见功能与健壮性测试用例，并按组件重新组织。
>
> 涉及组件：
>
> - `cmd/onex-apiserver`：OneX 声明式 API 服务端（存储到 etcd）。
> - `cmd/onex-controller-manager`：回收类控制器（GC / namespace finalizer / clusterrole-aggregation）。
> - `cmd/onex-job-controller`：Job 与 CronJob 控制器。
>
> 所有操作目录为 `${PROJ_ROOT_DIR}`，密码统一为 `onex(#)666`。

---

## 1. 测试对象与视角

OneX 是一套类 Kubernetes 的声明式系统，核心闭环：

```
声明期望状态(Spec) ──▶ onex-apiserver 持久化到 etcd ──▶ 控制器 watch/reconcile ──▶ 驱动真实资源走向 Spec
                                                              ▲                    │
                                                              └── 失败则重试（backoff）──┘
```

本指南从四个视角组织测试：

| 视角 | 说明 | 对应分组 |
| --- | --- | --- |
| 功能测试 | CRUD / 字段校验 / 子资源 / 持久化 / watch | A（apiserver） |
| 组件测试 | 健康检查 / Leader 选举 / 控制器开关 / GC / namespace 级联删除 | B（controller-manager） |
| 生命周期测试 | Job/CronJob 状态机、suspend、超时、并发策略 | C（job-controller） |
| 重试/退避测试 | 失败后的指数退避重试、冲突重试、最终收敛 | D（贯穿 A/B/C 的重试用例汇总） |

## 2. 组件与资源速览

| 组件 | 职责 | 关键资源/控制器 | 关键源码 |
| --- | --- | --- | --- |
| `onex-apiserver` | 声明式 API 服务端：认证、admission、校验、etcd 持久化 | `apps.onex.io/v1beta1`、`batch.onex.io/v1beta1` | `cmd/onex-apiserver/app/server.go`、`internal/apiserver/registry/*/rest` |
| `onex-controller-manager` | 回收类控制器 | `garbage-collector`、`namespaced-resource-deleter`、`clusterrole-aggregation` | `cmd/onex-controller-manager/app/*`、`internal/controller/namespace/controller.go` |
| `onex-job-controller` | 任务生命周期控制器 | `job-controller`、`cronjob-controller` | `internal/controller/job/job/controller.go`、`internal/controller/job/cronjob/controller.go` |

### 2.1 资源清单

| Kind | Group/Version | 资源名 | 子资源 | 作用域 |
| --- | --- | --- | --- | --- |
| Chain | `apps.onex.io/v1beta1` | `chains` | `/status` | Namespaced（**仅限 `kube-system`**） |
| Miner | `apps.onex.io/v1beta1` | `miners` | `/status` | Namespaced |
| MinerSet | `apps.onex.io/v1beta1` | `minersets` | `/status`、`/scale` | Namespaced |
| Job | `batch.onex.io/v1beta1` | `jobs` | `/status` | Namespaced |
| CronJob | `batch.onex.io/v1beta1` | `cronjobs` | `/status` | Namespaced |

### 2.2 关键校验规则（服务端 enforcement，测试字段校验用）

> 来源：`pkg/apis/apps/validation/`、`pkg/apis/batch/validation/`、`internal/apiserver/admission/plugin/minerset/`。

| 资源 | 规则 |
| --- | --- |
| Chain | name 需 DNS 子域；**namespace 必须为 `kube-system`**，否则 `Forbidden` |
| Miner / MinerSet | name 需 DNS 子域（`NameIsDNSSubdomain`） |
| Job | name 需 DNS 子域；`activeDeadlineSeconds` 必须 ≥ 0 |
| CronJob | name 需 DNS 子域且 **≤ 52 字符**（控制器会追加 11 字符时间戳，总长 ≤63）；`schedule` 必填且符合 cron 格式；`startingDeadlineSeconds` ≥ 0；`timeZone` 合法 |
| MinerSet（Admission） | 自动注入 `template.spec.displayName`；自动补全 selector；`apps.onex.io/deletion-protection=true` 时禁止删除 |

### 2.3 组件默认参数（Leader 选举 / 健康检查）

> 来源：`pkg/config/v1beta1/defaults.go`。

| 参数 | 默认值 |
| --- | --- |
| healthz | `0.0.0.0:20250` |
| metrics | `0.0.0.0:20251` |
| parallelism | `16` |
| syncPeriod | `10h`（带 jitter） |
| controllers | `["*"]`（全启用） |
| leader election | `leases` 锁，namespace `kube-system`，锁名 `controller-manager`（Lease 15s / Renew 10s / Retry 2s） |

---

## 3. 测试环境准备

```bash
cd ${PROJ_ROOT_DIR}
export ONEX_CONFIG=$HOME/.onex
export KUBECONFIG=$ONEX_CONFIG/config

# 1. etcd
etcd --data-dir=/tmp/etcd &

# 2. onex-apiserver
./bin/onex-apiserver \
  --etcd-servers 127.0.0.1:2379 \
  --secure-port 52443 \
  --client-ca-file=$ONEX_CONFIG/cert/ca.pem \
  --tls-cert-file=$ONEX_CONFIG/cert/onex-apiserver.pem \
  --tls-private-key-file=$ONEX_CONFIG/cert/onex-apiserver-key.pem &

# 3. onex-controller-manager
./bin/onex-controller-manager --kubeconfig $ONEX_CONFIG/config &

# 4. onex-job-controller
./bin/onex-job-controller --kubeconfig $ONEX_CONFIG/config \
  --config $ONEX_CONFIG/onex-controller-manager.yaml &
```

验证资源注册：

```bash
kubectl api-resources | grep onex.io
```

## 4. 观察与验证手段

```bash
# 事件（失败/重试大多会生成 Event）
kubectl get events --all-namespaces --sort-by=.lastTimestamp -w

# 状态与 condition
kubectl get <resource> <name> -o jsonpath='{.status.conditions}{"\n"}{.status.phase}'

# 控制器日志（提高 verbosity）
./bin/onex-job-controller --v=4 ...          # 调度/退避细节
./bin/onex-controller-manager --v=5 ...      # namespace deleteCollection/listCollection 细节

# 健康检查 & metrics
curl -s http://127.0.0.1:20250/healthz        # controller-manager
curl -s http://127.0.0.1:20251/metrics        # controller-manager metrics
kubectl get --raw /healthz                    # apiserver
kubectl get --raw /readyz
```

---

## 5. 测试用例明细（step by step）

> 用例编号规则：`A-*`（apiserver）、`B-*`（controller-manager）、`C-*`（job-controller）、`D-*`（重试/退避专项）。

### 5.1 分组 A：onex-apiserver

#### A-1 启动与健康检查 / 版本 / API 发现

**步骤**：

```bash
# 健康检查
kubectl get --raw /healthz          # 期望 ok
kubectl get --raw /readyz           # 期望 ok

# 版本
kubectl version

# API 发现（含聚合层）
kubectl get --raw /apis/apps.onex.io/v1beta1 | python3 -m json.tool
kubectl get --raw /apis/batch.onex.io/v1beta1 | python3 -m json.tool
```

**预期**：四个端点均正常；`/apis/{group}/v1beta1` 返回该组全部资源（`chains`、`miners`、`minersets` / `jobs`、`cronjobs`）。

---

#### A-2 资源 CRUD（create / get / list / watch / patch / delete）

**步骤**（以 Miner 为例，其余资源同理）：

```bash
# create
kubectl apply -f - <<'EOF'
apiVersion: apps.onex.io/v1beta1
kind: Miner
metadata:
  name: crud-miner
  namespace: default
spec:
  minerType: tiny
  chainName: my-chain
EOF

# get
kubectl get miner crud-miner -o yaml

# list（含 label selector）
kubectl get miners -l apps.onex.io/chain-name=my-chain

# watch
kubectl get miners -w &
kubectl patch miner crud-miner --type=merge -p '{"spec":{"displayName":"renamed"}}'

# delete
kubectl delete miner crud-miner
```

**预期**：四类操作均返回一致结果；watch 能观察到 update 事件。

**审计核对**（分组 A 全部资源的 CRUD）见第 6 节「资源操作示例」。

---

#### A-3 MinerSet Admission 自动注入（Mutate）

**步骤**：

```bash
kubectl apply -f - <<'EOF'
apiVersion: apps.onex.io/v1beta1
kind: MinerSet
metadata:
  name: ms-demo
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: miner
  template:
    metadata:
      labels:
        app: miner
    spec:
      minerType: tiny
EOF

kubectl get minerset ms-demo -o yaml | grep -A2 displayName
```

**预期**：`spec.template.spec.displayName` 被自动置为 `miner-for-ms-demo-minerset`，selector 与 template labels 一致。

---

#### A-4 MinerSet 删除保护（Admission 拒绝）

**步骤**：

```bash
kubectl annotate minerset ms-demo apps.onex.io/deletion-protection=true
kubectl delete minerset ms-demo        # 期望被拒绝
kubectl get minerset ms-demo           # 依旧存在
```

**预期**：删除被 Admission 拒绝，报错含 `deletion protection`。

---

#### A-5 status 子资源读写分离

**步骤**：

```bash
# 尝试在创建时写 status —— 应被忽略
kubectl apply -f - <<'EOF'
apiVersion: apps.onex.io/v1beta1
kind: Miner
metadata:
  name: status-demo
  namespace: default
spec:
  minerType: tiny
status:
  phase: Running
EOF

# 直接写 status 子资源（仅控制器/RBAC 允许）
kubectl patch miner status-demo --subresource=status --type=merge -p '{"status":{"phase":"Running"}}'
```

**预期**：`status` 由系统/控制器独占维护；用户声明式里传的 status 不生效。

---

#### A-6 MinerSet /scale 子资源

**步骤**：

```bash
kubectl scale --replicas=5 minerset ms-demo          # 扩容
kubectl get minerset ms-demo/scale -o yaml            # 读取 scale
```

**预期**：`spec.replicas` 更新为 5，scale 子资源可读。

---

#### A-7 乐观锁（resourceVersion 冲突）

**步骤**：

```bash
kubectl get miner crud-miner -o yaml > /tmp/m.yaml        # 拿旧 resourceVersion
# 手工修改 /tmp/m.yaml（改 spec）后 apply，与此同时另一终端再 patch 一次
kubectl apply -f /tmp/m.yaml
```

**预期**：并发写时返回 `Operation cannot be fulfilled on ... the object has been modified`，不会静默覆盖。

---

#### A-8 字段校验（拒绝非法声明）

**步骤**：

```bash
# 1) Chain 放在非 kube-system → Forbidden
kubectl apply -f - <<'EOF'
apiVersion: apps.onex.io/v1beta1
kind: Chain
metadata:
  name: bad-chain
  namespace: default
spec:
  minerType: tiny
EOF

# 2) CronJob 名字超过 52 字符 → Invalid
kubectl apply -f - <<'EOF'
apiVersion: batch.onex.io/v1beta1
kind: CronJob
metadata:
  name: this-cronjob-name-is-way-too-long-to-be-valid-and-must-be-rejected-by-server
  namespace: default
spec:
  schedule: "* * * * *"
  jobTemplate:
    spec:
      type: kubernetes
EOF

# 3) CronJob 缺 schedule → Required
kubectl apply -f - <<'EOF'
apiVersion: batch.onex.io/v1beta1
kind: CronJob
metadata:
  name: no-schedule
  namespace: default
spec:
  jobTemplate:
    spec:
      type: kubernetes
EOF

# 4) Job activeDeadlineSeconds 为负 → Invalid
kubectl apply -f - <<'EOF'
apiVersion: batch.onex.io/v1beta1
kind: Job
metadata:
  name: neg-deadline
  namespace: default
spec:
  type: kubernetes
  activeDeadlineSeconds: -1
EOF
```

**预期**：全部被服务端拒绝，返回对应 `Forbidden` / `Invalid` / `Required`。

---

#### A-9 etcd 持久化（重启 apiserver 后仍在）

**步骤**：

1. `kubectl apply` 创建若干资源。
2. 重启 `onex-apiserver`。
3. `kubectl get ...` 再次查询。

**预期**：重启后资源完整存在（存储前缀 `/registry/onex.io`）。

---

#### A-10 watch 能力

**步骤**：

```bash
kubectl get miners -w &
# 另一终端 apply/delete 一个 miner，观察 watch 推送 ADDED/MODIFIED/DELETED
```

**预期**：watch 流正确推送事件。

---

### 5.2 分组 B：onex-controller-manager

#### B-1 Leader 选举（多实例高可用）

**前提**：两个 controller-manager 实例连同一 apiserver。

**步骤**：

```bash
# 实例 1 与实例 2 使用相同 --leader-elect 配置（默认 leases 锁）
./bin/onex-controller-manager --kubeconfig $ONEX_CONFIG/config &
./bin/onex-controller-manager --kubeconfig $ONEX_CONFIG/config &

# 观察 leases
kubectl get lease -n kube-system controller-manager -o yaml
```

**预期**：只有一个实例成为 leader 并实际 reconcile；另一个为 standby。kill leader 后 standby 在租约到期（约 15s）内接管。

**验证**：

```bash
kubectl describe lease -n kube-system controller-manager   # HolderIdentity 指向 leader
```

---

#### B-2 健康检查与指标

**步骤**：

```bash
curl -s http://127.0.0.1:20250/healthz   # ok
curl -s http://127.0.0.1:20250/readyz    # ok
curl -s http://127.0.0.1:20251/metrics | grep controller_runtime
```

**预期**：healthz/readyz 返回 ok；metrics 暴露 controller-runtime 指标。

---

#### B-3 控制器启用/禁用（--controllers）

**步骤**：

```bash
# 仅启用类回收控制器，禁用 namespaced-resource-deleter
./bin/onex-controller-manager \
  --kubeconfig $ONEX_CONFIG/config \
  --controllers=garbage-collector-controller,clusterrole-aggregation-controller
```

**预期**：启动日志打印被跳过的控制器 `Warning: controller is disabled`。

---

#### B-4 GC 回收孤儿对象

**前提**：`garbage-collector-controller` 启用。

**步骤**：

1. 创建带 `ownerReferences`（被某父对象拥有）的 Job；随后删除父对象。
2. 观察 GC 依赖图跟踪后，孤儿 Job 被后台回收（`DeletePropagationBackground`）。

**预期**：owner 被删除后，无其他引用者的依赖对象被 GC 回收。

**验证**：`kubectl get job <name>` 最终 not found。

---

#### B-5 Namespace 级联删除（含 finalizer / 冲突重试）

**步骤**：

```bash
kubectl create ns demo
kubectl -n demo apply -f jobs.yaml        # 若干 Job

kubectl delete ns demo
kubectl get ns demo -o jsonpath='{.status.phase}'   # Terminating
```

**预期（`namespaced-resource-deleter` 控制器流程）**：

1. 确认 `deletionTimestamp` 已设置；把 ns 置为 `Terminating`。
2. 通过 discovery 找到所有支持 `delete` 的 GVR，优先 `deletecollection`，不支持则逐条删除。
3. 有 finalizer 残留时返回 `ResourcesRemainingError` 触发重试（estimation 默认 15s）。
4. 清空后移除 `kubernetes` finalizer，ns 彻底删除。

**附加（乐观锁冲突重试，D 组展开）**：并发修改 ns status 制造 `resourceVersion` 冲突，控制器应自动 `retryOnConflictError`。

---

#### B-6 ClusterRole 聚合

**前提**：`clusterrole-aggregation-controller` 启用。

**步骤**：

1. 创建带 `aggregationRule` 的 ClusterRole。

**预期**：控制器把被引用 ClusterRole 的 rules 合并进该 ClusterRole（5 个 worker）。

---

#### B-7 等待 API Server（WaitForAPIServer）

**步骤**：

1. 先不启动 apiserver，直接启动 controller-manager。

**预期**：controller-manager 不会立即退出，而是等待 apiserver 健康（最长 10s）后失败报错 `failed to wait for apiserver being healthy`。

---

### 5.3 分组 C：onex-job-controller

#### C-1 Job 生命周期：Pending → Running → Succeeded

**步骤**：

```bash
kubectl apply -f - <<'EOF'
apiVersion: batch.onex.io/v1beta1
kind: Job
metadata:
  name: lifecycle-demo
  namespace: default
spec:
  type: kubernetes
  providerSpec:
    restartPolicy: Never
    containers:
    - name: main
      image: busybox
      command: ["sh", "-c", "echo done && sleep 3"]
EOF

kubectl get job lifecycle-demo -w
```

**预期**：`phase` 依次 `Pending → Running → Succeeded`，`startedAt`/`endedAt` 填充，`conditions` 出现 `Complete=True`，并派发对应 Pod（`job.onex.io/name=lifecycle-demo`）。

---

#### C-2 suspend 挂起与恢复

**步骤**：

```bash
kubectl patch job lifecycle-demo --type=merge -p '{"spec":{"suspend":true}}'
# 期望：conditions 出现 Suspended=True，phase 回 Pending，startedAt 清空，Pod 被删除
kubectl patch job lifecycle-demo --type=merge -p '{"spec":{"suspend":false}}'
# 期望：重新派发 Pod 并重置 startedAt
```

**验证**：

```bash
kubectl get pods -l job.onex.io/name=lifecycle-demo   # 挂起时 0，恢复后 1
```

---

#### C-3 activeDeadlineSeconds 超时强制失败

**步骤**：创建 `activeDeadlineSeconds: 5`、Pod 执行 `sleep 300` 的 Job，等待约 5s。

**预期**：`phase=Failed`，`conditions` 含 `Failed=True` 且 `reason=DeadlineExceeded`，`errorMessage` 记录超时信息，Pod 被删除。

---

#### C-4 CronJob 定时派发与活动列表（active / lastScheduleTime）

**步骤**：

```bash
kubectl apply -f - <<'EOF'
apiVersion: batch.onex.io/v1beta1
kind: CronJob
metadata:
  name: every-minute
  namespace: default
spec:
  schedule: "* * * * *"
  jobTemplate:
    spec:
      type: kubernetes
      providerSpec:
        restartPolicy: Never
        containers:
        - name: c
          image: busybox
          command: ["sleep", "5"]
EOF

kubectl get cronjob every-minute -o yaml   # 观察 status.active / lastScheduleTime
kubectl get jobs                            # 派生 Job 名称为 every-minute-<hash>
```

**预期**：每分钟派生一个 Job，运行期间入 `active`，完成后移除并更新 `lastScheduleTime`。

---

#### C-5 CronJob 并发策略（Allow / Forbid / Replace）

**步骤**：对一个执行较慢（`sleep 120`）的 CronJob 分别设置：

```bash
kubectl patch cronjob every-minute --type=merge -p '{"spec":{"concurrencyPolicy":"Forbid"}}'   # 跳过下一调度，事件 JobAlreadyActive
kubectl patch cronjob every-minute --type=merge -p '{"spec":{"concurrencyPolicy":"Replace"}}'  # 删除旧 Job 再派发
kubectl patch cronjob every-minute --type=merge -p '{"spec":{"concurrencyPolicy":"Allow"}}'    # 允许并发
```

**验证**：

```bash
kubectl get events --field-selector involvedObject.kind=CronJob,name=every-minute
```

---

#### C-6 CronJob 错过调度窗口 / 历史清理

```bash
# 错过调度（startingDeadlineSeconds 内未执行）
kubectl patch cronjob every-minute --type=merge -p '{"spec":{"startingDeadlineSeconds":10}}'
kubectl get events --field-selector reason=MissSchedule

# 历史清理
kubectl patch cronjob every-minute --type=merge -p '{"spec":{"successfulJobsHistoryLimit":1,"failedJobsHistoryLimit":1}}'
kubectl get jobs   # 数量收敛到上限
```

**预期**：错过调度产生 `MissSchedule`；历史 Job 保留最近 1 成功 + 1 失败。

---

#### C-7 watchFilterValue 选择性 reconcile

**步骤**：

1. job-controller 以 `--watch-filter=foo` 启动。
2. 创建带 `apps.onex.io/watch-filter=foo` 与不带该 label 的 Job。

**预期**：只有带匹配 label 的 Job 被 reconcile。

---

### 5.4 分组 D：重试 / 退避专项（整合自 decl-test.md）

> 其余重试用例已在 A/B/C 中标注对应的最终收敛行为，此处聚焦**退避机制本身**。

#### D-1 重试 / 退避机制速查

| 层次 | 机制 | 参数 | 说明 |
| --- | --- | --- | --- |
| reconcile 队列 | `ratelimiter.DefaultControllerRateLimiter()` | 单项指数退避 `200ms → 1h`；全局限流 5000qps / 突发 10000 | 同一对象反复失败，重试间隔越来越长 |
| Job 启动退避 | `backoffStore` | 基础 5s，`×2^n`，上限 5m；成功即 reset | 派发失败后指数退避重试 |
| 写后读一致性 | `ControllerExpectations` | 未满足则 `RequeueAfter: 1s` | 等 informer 观察到自己的写再继续 |
| Job 运行态轮询 | `providerPollInterval` | 10s | 运行中每 10s 轮询 provider |
| Namespace 冲突重试 | `retryOnConflictError` | 冲突时重新 Get 后重试 | UID 变化则报错 |
| Namespace 资源未清空 | `ResourcesRemainingError` | estimate 秒后重试 | finalizer 残留触发 requeue |
| GC 同步 | syncPeriod | 30s | 定期刷新 RESTMapper 并同步 |
| 启动等待 | `WaitForAPIServer` | 10s | 启动时等待 apiserver 健康 |

#### D-2 Job 首次派发失败 → 指数退避重试 → 修复后收敛（key 用例）

**步骤**：

1. 创建 providerSpec 缺失的 Job（`kubernetes` provider 返回 “has no providerSpec”）。

   ```bash
   kubectl apply -f - <<'EOF'
   apiVersion: batch.onex.io/v1beta1
   kind: Job
   metadata:
     name: retry-spec
     namespace: default
   spec:
     type: kubernetes
     # 故意不填 providerSpec
   EOF
   ```

2. 观察 `StartFailed` 事件；控制器日志（`-v=4`）中退避间隔 5s → 10s → 20s … 上限 5m。

3. 补上 providerSpec 修复声明：

   ```bash
   kubectl patch job retry-spec --type=merge -p \
     '{"spec":{"providerSpec":{"restartPolicy":"Never","containers":[{"name":"c","image":"busybox","command":["sh","-c","sleep 5"]}]}}}'
   ```

4. 观察收敛到 `Succeeded`。

**预期**：失败重试间隔指数增长且封顶 5m；修复后 backoff reset，最终 `Succeeded`。

#### D-3 写后读一致性（expectations）不重复派发

**步骤**：反复 apply 同一 Job，观察控制器在 informer 未观察到自己的写时 `RequeueAfter: 1s`，不重复派发。

**预期**：`job.onex.io/name` 标签下最多一个 Pod。

#### D-4 Namespace resourceVersion 冲突自动重试

**步骤**：ns 级联删除过程中，并发修改 ns status 制造冲突。

**预期**：控制器自动重新 Get 后再 Update；如 UID 变化则报错 `namespace uid has changed across retries` 而非 panic。

#### D-5 重试验证方法论

```bash
# 事件
kubectl get events --all-namespaces --sort-by=.lastTimestamp | tail -50

# 状态
kubectl get <resource> <name> -o jsonpath='{.status.conditions}{"\n"}{.status.phase}'

# 证伪“一次成功”：注入瞬时故障（停 etcd / 删 Pod / 制造冲突），
# 观察系统经多次重试后仍收敛。
```

---

## 6. 资源操作示例（整合自 decl-test.md）

> 下列示例均针对 `onex-apiserver`，方式与 Kubernetes 原生对象一致。

### 6.1 Chain（注意：仅限 kube-system）

```bash
kubectl apply -f - <<'EOF'
apiVersion: apps.onex.io/v1beta1
kind: Chain
metadata:
  name: my-chain
  namespace: kube-system
spec:
  displayName: "Demo Chain"
  minerType: tiny
  image: "onex/example-node:v1"
  minMineIntervalSeconds: 15
EOF

kubectl get chain my-chain -n kube-system -o yaml
kubectl patch chain my-chain -n kube-system --type=merge -p '{"spec":{"minMineIntervalSeconds":30}}'
kubectl patch chain my-chain -n kube-system --subresource=status --type=merge -p '{"status":{"observedGeneration":1}}'
kubectl delete chain my-chain -n kube-system
```

### 6.2 Miner

```bash
kubectl apply -f - <<'EOF'
apiVersion: apps.onex.io/v1beta1
kind: Miner
metadata:
  name: my-miner
  namespace: default
  labels:
    apps.onex.io/chain-name: my-chain
spec:
  displayName: "Demo Miner"
  minerType: tiny
  chainName: my-chain
  restartPolicy: Always
  podDeletionTimeout: 10s
EOF

kubectl get miner my-miner -o wide
kubectl get miner my-miner -o jsonpath='{.status.phase}'
kubectl delete miner my-miner
```

### 6.3 MinerSet（含 /scale）

```bash
kubectl apply -f - <<'EOF'
apiVersion: apps.onex.io/v1beta1
kind: MinerSet
metadata:
  name: my-minerset
  namespace: default
spec:
  replicas: 3
  deletePolicy: Random
  selector:
    matchLabels:
      app: miner
  template:
    metadata:
      labels:
        app: miner
    spec:
      minerType: tiny
      chainName: my-chain
EOF

kubectl scale --replicas=5 minerset my-minerset
kubectl get minerset my-minerset/scale -o yaml
kubectl get minerset my-minerset -o jsonpath='{.status.replicas}/{.status.readyReplicas}'
kubectl delete minerset my-minerset
```

### 6.4 Job

```bash
kubectl apply -f - <<'EOF'
apiVersion: batch.onex.io/v1beta1
kind: Job
metadata:
  name: my-job
  namespace: default
spec:
  type: kubernetes
  activeDeadlineSeconds: 60
  providerSpec:
    # providerSpec 是 runtime.RawExtension，实际内容为 corev1.PodSpec 的 JSON
    restartPolicy: Never
    containers:
    - name: main
      image: busybox
      command: ["sh", "-c", "echo hello && sleep 3"]
EOF

kubectl get job my-job -w                     # Pending -> Running -> Succeeded
kubectl get job my-job -o jsonpath='{.status.phase}'
kubectl delete job my-job
```

### 6.5 CronJob

```bash
kubectl apply -f - <<'EOF'
apiVersion: batch.onex.io/v1beta1
kind: CronJob
metadata:
  name: my-cronjob
  namespace: default
spec:
  schedule: "*/5 * * * *"
  concurrencyPolicy: Forbid
  startingDeadlineSeconds: 30
  successfulJobsHistoryLimit: 3
  failedJobsHistoryLimit: 1
  jobTemplate:
    spec:
      type: kubernetes
      providerSpec:
        restartPolicy: Never
        containers:
        - name: main
          image: busybox
          command: ["sleep", "10"]
EOF

kubectl get cronjob my-cronjob -o yaml        # status.active / lastScheduleTime
kubectl get jobs                              # 派生的 Job
kubectl delete cronjob my-cronjob
```

### 6.6 label selector 查询

```bash
kubectl get miners -l apps.onex.io/chain-name=my-chain
kubectl get miners -l apps.onex.io/minerset-name=my-minerset
kubectl get miners -l apps.onex.io/watch-filter=<filter-value>
```

---

## 7. 附录：关键源码索引

| 主题 | 路径 |
| --- | --- |
| APIServer 主入口 | `cmd/onex-apiserver/apiserver.go` |
| APIServer 命令/启动链 | `cmd/onex-apiserver/app/server.go` |
| APIServer Options | `cmd/onex-apiserver/app/options/options.go` |
| REST 存储注册（apps） | `internal/apiserver/registry/apps/rest/storage_apps.go` |
| REST 存储注册（batch） | `internal/apiserver/registry/batch/rest/storage_batch.go` |
| MinerSet Admission | `internal/apiserver/admission/plugin/minerset/admission.go` |
| 字段校验（apps） | `pkg/apis/apps/validation/` |
| 字段校验（batch） | `pkg/apis/batch/validation/` |
| controller-manager 命令 | `cmd/onex-controller-manager/app/controllermanager.go` |
| GC / ClusterRoleAggregation | `cmd/onex-controller-manager/app/core.go`、`rbac.go` |
| Namespace 级联删除 | `internal/controller/namespace/controller.go` |
| 控制器名称常量 | `cmd/onex-controller-manager/names/controller_names.go` |
| 默认参数 | `pkg/config/v1beta1/defaults.go` |
| job-controller 命令 | `cmd/onex-job-controller/app/controller.go` |
| Job 控制器 | `internal/controller/job/job/controller.go` |
| Job 退避 | `internal/controller/job/job/backoff_utils.go` |
| Job 写后读一致性 | `internal/controller/job/job/expectations.go` |
| Job provider（kubernetes） | `internal/controller/job/job/provider.go` |
| CronJob 控制器 | `internal/controller/job/cronjob/controller.go` |
| 统一 RateLimiter | `internal/pkg/util/ratelimiter/ratelimiter.go` |
| API 类型定义 | `pkg/apis/apps/v1beta1/`、`pkg/apis/batch/v1beta1/` |