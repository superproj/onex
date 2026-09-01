# LLMTrain 类型 Job / CronJob 测试用例

> 本文档覆盖 `onex-job-controller` 中 `spec.type: LLMTrain` 的 **Job** 与 **CronJob** 生命周期、失败、重试/退避及并发等测试用例，与 `docs/devel/zh-CN/decl-test.md` 的分组 C/D（job-controller）互补。环境准备、组件启动方式请复用 `decl-test.md` 第 3 节；本文只列出与 `LLMTrain` 类型相关的差异与用例。

---

## 1. LLMTrain provider 特性速览

`LLMTrain` 是 `onex-job-controller` 通过 **可插拔 provider 注册表**（`internal/controller/job/job/providers/registry`）挂载的一个示例 Job 类型，用于演示如何基于当前框架新增自定义 JobType。核心实现：`internal/controller/job/job/providers/llmtrain/`。

关键行为（与 `kubernetes` 类型对比）：

| 维度 | `kubernetes` | `LLMTrain` |
| --- | --- | --- |
| 工作负载形态 | 派发一个 Pod（`job.onex.io/name=<job>`） | **进程内 FSM 流水线**，不创建任何 K8s 子资源 |
| 状态载体 | Pod phase | `job.status.providerStatus`（JSON） |
| 进度推进 | Pod 状态变化 | 每 10s 轮询推进**一段** FSM |
| 依赖 | K8s 集群 | **自包含 demo**：fake MinIO + 确定性 embed + 内存 trainer（无任何外部依赖） |
| Job 最终状态 | Succeeded / Failed | Succeeded / Failed（由 provider 终态映射） |

> ⚠️ 自包含 demo 说明：默认 factory 使用 `fakeminio.NewFakeMinioClient`（预置源对象 `llm/test.json`）、确定性的 embedding、以及 `train.TrainManager`（20s 后任务 `Completed`）。因此以下用例**无需外部 MinIO / Ollama / 训练平台即可端到端跑通**。若要接真实平台，只需替换 `newProvider` 的 `minio.IMinio` / `trainer` / `embedder` 三个接口实现。

### 1.1 providerSpec 字段（`spec.providerSpec`，`runtime.RawExtension`）

`LLMTrainProviderSpec` 结构（JSON/YAML 内联到 `providerSpec`）：

| 字段 | 类型 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- | --- |
| `sourceObject` | string | 否 | `llm/test.json` | 原始训练数据的对象存储 key |
| `jobTimeoutSec` | int64 | 否 | `14400`（4h） | 训练最大时长（秒），provider 内部超时判断 |
| `idempotentExecution` | bool | 否 | `false` | 幂等执行开关（见 LT-J-02） |

### 1.2 providerStatus 字段（`status.providerStatus`，`runtime.RawExtension`）

`LLMTrainProviderStatus` 流水线状态（由 provider 每次轮询序列化写回，随 status 子资源持久化，**controller 重启后可从中断 phase 续跑**）：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `phase` | string | 当前 FSM 阶段，取值见下表 |
| `dataPath` | string | 下载段产物：原始数据对象 key |
| `embeddedDataPath` | string | 嵌入段产物：向量数据对象 key |
| `taskID` | string | 训练任务 ID（首次提交后置位，幂等） |
| `resultPath` | string | 训练结果目标对象 key |
| `message` | string | 终态失败原因 |

FSM 阶段（`phase`）：

```
Pending → Downloading → Downloaded → Embedding → Embedded → Training → Trained → Succeeded
                                                                        ↘ Failed
```

- `Downloading/Embedding/Training` 为实际执行段（分别对应 download / embed / train 三段逻辑）。
- `Training` 段会停留在该阶段轮询，直到 trainer 返回 `Completed` 才进入 `Trained`。
- 任一段出错 → `phase=Failed`（**终态失败**，不回退重试，见 §4.1）。

### 1.3 Job 状态机（双状态）

`LLMTrain` Job 仍走标准 Job 状态机，同时 provider 在 `status.providerStatus.phase` 里暴露细粒度进度：

| 观察点 | 取值 |
| --- | --- |
| `status.phase` | `Pending → Running → Succeeded/Failed`（运行期恒为 `Running`） |
| `status.providerStatus.phase` | FSM 七阶段 + `Failed`（运行期逐步推进） |
| `status.conditions` | 终态出现 `Complete=True` 或 `Failed=True` |

---

## 2. 环境准备

```bash
cd ${PROJ_ROOT_DIR}
export KUBECONFIG=$HOME/.onex/config

# etcd / onex-apiserver / onex-job-controller 启动方式同 decl-test.md 第 3 节
# 确认 job 控制器已运行
./bin/onex-job-controller --kubeconfig $KUBECONFIG --config $ONEX_CONFIG/onex-controller-manager.yaml &

# 观察事件与状态
kubectl get events --all-namespaces --sort-by=.lastTimestamp -w
kubectl get job <name> -o jsonpath='{.status.phase}{"\n"}{.status.providerStatus}{"\n"}'
```

---

## 3. Job 用例（LT-J）

### LT-J-01 完整流水线：Pending → Running → Succeeded（核心用例）

**步骤**：

```bash
kubectl apply -f - <<'EOF'
apiVersion: batch.onex.io/v1beta1
kind: Job
metadata:
  name: llmtrain-demo
  namespace: default
spec:
  type: LLMTrain
  providerSpec:
    sourceObject: llm/test.json
    jobTimeoutSec: 14400
    idempotentExecution: false
EOF

kubectl get job llmtrain-demo -w
```

**验证**：

```bash
kubectl get job llmtrain-demo -o jsonpath='{.status.phase}{"\n"}{.status.conditions}{"\n"}'
kubectl get job llmtrain-demo -o jsonpath='{.status.providerStatus}'
```

**预期**：

- `status.phase` 依次 `Pending → Running → Succeeded`（运行期约 1~2 分钟，训练段受 20s 模拟延时会多停留几次轮询）。
- `status.providerStatus.phase` 依次经过 `Downloading → Downloaded → Embedding → Embedded → Training → Trained → Succeeded`。
- **无任何 Pod 派生**：`kubectl get pods -l job.onex.io/name=llmtrain-demo` 恒为空。
- 终态 `providerStatus` 内 `dataPath` / `embeddedDataPath` / `taskID` / `resultPath` 均已填充，形如：

```json
{"phase":"Succeeded","dataPath":"job/llmtrain-demo/data.json","embeddedDataPath":"job/llmtrain-demo/embedded.json","taskID":"task-1","resultPath":"job/llmtrain-demo/result.json"}
```

---

### LT-J-02 幂等执行开关（idempotentExecution）

**步骤**：创建 `idempotentExecution: true` 的 Job，并在 `Downloading` 阶段中断 controller（`kill` 后重启），观察从 `status.providerStatus` 里已完成的段继续、不重复执行已 `True` 的段。

```bash
kubectl apply -f - <<'EOF'
apiVersion: batch.onex.io/v1beta1
kind: Job
metadata:
  name: llmtrain-idem
  namespace: default
spec:
  type: LLMTrain
  providerSpec:
    idempotentExecution: true
EOF

# 观察 providerStatus.phase 推进到某一中间段后，重启 onex-job-controller
kubectl get job llmtrain-idem -o jsonpath='{.status.providerStatus}{"\n"}'
```

**预期**：controller 重启后 job 不从头开始，而是从 `status.providerStatus.phase` 记录的阶段继续；已完成的段（`dataPath` 等已非空）被跳过，最终收敛到 `Succeeded`。

---

### LT-J-03 自定义超时 → Failed（provider 内部超时）

**步骤**：

```bash
kubectl apply -f - <<'EOF'
apiVersion: batch.onex.io/v1beta1
kind: Job
metadata:
  name: llmtrain-timeout
  namespace: default
spec:
  type: LLMTrain
  providerSpec:
    jobTimeoutSec: 5
EOF
```

**预期**：`jobTimeoutSec` 很小，训练流水线无法在 5s 内完成（训练段本身至少 20s）。约 5s 后 `status.phase=Failed`，`status.providerStatus.phase=Failed`，`status.providerStatus.message` 形如 `llm train task exceeded 5 seconds`，`conditions` 出现 `Failed=True`。

> 与 Job 级 `activeDeadlineSeconds` 的区别：`jobTimeoutSec` 是 **provider 内部**超时（在 `Status` 里判断，只作用于 LLMTrain）；`activeDeadlineSeconds` 是 Job 控制器在 `sync` 中**先行强制**的通用超时（见 §4.2），二者可叠加。

---

### LT-J-04 providerSpec 省略 / sourceObject 缺省

**步骤**：创建完全不带 `providerSpec` 的 `LLMTrain` Job。

```bash
kubectl apply -f - <<'EOF'
apiVersion: batch.onex.io/v1beta1
kind: Job
metadata:
  name: llmtrain-default
  namespace: default
spec:
  type: LLMTrain
EOF
```

**预期**：`providerSpec` 缺省 → `decodeSpec` 使用默认值（`sourceObject=llm/test.json`、`jobTimeoutSec=14400`、`idempotentExecution=false`），流水线正常跑通到 `Succeeded`（**不会**因缺失 providerSpec 而失败，与 `kubernetes` 类型不同）。

---

### LT-J-05 type 大小写 / 未知类型 → 未注册 provider

**步骤**：

```bash
# 小写 llmtrain（未注册：注册表 key 为大写 "LLMTrain"）
kubectl apply -f - <<'EOF'
apiVersion: batch.onex.io/v1beta1
kind: Job
metadata:
  name: llmtrain-lowercase
  namespace: default
spec:
  type: llmtrain
EOF
```

**验证**：

```bash
kubectl get events --field-selector involvedObject.name=llmtrain-lowercase
kubectl get job llmtrain-lowercase -o jsonpath='{.status.conditions}'
```

**预期**：`spec.type` 严格区分大小写。小写 `llmtrain` 未注册，派发失败，事件报 `StartFailed ... unsupported job type "llmtrain"`，随后进入指数退避重试（见 §4.3）。同理，`type` 留空会**默认落到 `kubernetes`**，而非 LLMTrain。

---

### LT-J-06 suspend 挂起与恢复（无 Pod 可删）

**步骤**：

```bash
kubectl patch job llmtrain-demo --type=merge -p '{"spec":{"suspend":true}}'
kubectl get job llmtrain-demo -o jsonpath='{.status.conditions}{"\n"}{.status.phase}'
kubectl patch job llmtrain-demo --type=merge -p '{"spec":{"suspend":false}}'
```

**预期**：

- 挂起后：`conditions` 出现 `Suspended=True`，`status.phase` 回 `Pending`，`startedAt` 清空；**无 Pod 需要删除**（对比 `kubernetes` 类型会删 Pod）。
- 恢复后：`startedAt` 重置，provider 重新 `Start`，`status.providerStatus` 被重置为 `phase: Pending`，流水线**从头重跑**。

---

## 4. CronJob 用例（LT-C）

> CronJob 本身是「定时生成 Job」的控制器；`type: LLMTrain` 只影响**派生的 Job** 的行为。CronJob 的调度/并发/历史清理语义不受 provider 类型影响，但派生出的每个 Job 都会独立跑一段 llmtrain 流水线。

### LT-C-01 定时派发 llmtrain Job

**步骤**：

```bash
kubectl apply -f - <<'EOF'
apiVersion: batch.onex.io/v1beta1
kind: CronJob
metadata:
  name: llmtrain-cron
  namespace: default
spec:
  schedule: "* * * * *"
  jobTemplate:
    spec:
      type: LLMTrain
      providerSpec:
        jobTimeoutSec: 14400
EOF

kubectl get cronjob llmtrain-cron -o yaml   # status.active / lastScheduleTime
kubectl get jobs
```

**预期**：每整分钟派生一个名称为 `llmtrain-cron-<11位时间戳>` 的 Job，其 `spec.type=LLMTrain`；该 Job 独立跑完流水线 → `Succeeded`；完成后从 `cronjob.status.active` 移除并更新 `lastScheduleTime`。

---

### LT-C-02 并发策略（Allow / Forbid / Replace）

**步骤**：将 `jobTemplate.spec.providerSpec.jobTimeoutSec` 设大（如 `600`）使 Job 运行时间较长；分别 patch 并发策略。

```bash
kubectl patch cronjob llmtrain-cron --type=merge -p '{"spec":{"concurrencyPolicy":"Forbid"}}'   # 上一轮未完则跳过，事件 JobAlreadyActive
kubectl patch cronjob llmtrain-cron --type=merge -p '{"spec":{"concurrencyPolicy":"Replace"}}'  # 删旧 Job 再派发
kubectl patch cronjob llmtrain-cron --type=merge -p '{"spec":{"concurrencyPolicy":"Allow"}}'    # 允许并发（默认）
```

**验证**：

```bash
kubectl get events --field-selector involvedObject.kind=CronJob,name=llmtrain-cron
kubectl get jobs   # 观察并发派发数量
```

**预期**：`Forbid` 跳过并发调度并产生 `JobAlreadyActive` 事件；`Replace` 删除仍运行的旧 Job 再派新 Job（`jobTemplate` 相同的 llmtrain Job 各自 carry 独立状态）；`Allow` 允许多个 llmtrain Job 同时运行，彼此 `status.providerStatus` 互不影响。

---

### LT-C-03 历史清理与错过调度

```bash
kubectl patch cronjob llmtrain-cron --type=merge -p '{"spec":{"successfulJobsHistoryLimit":1,"failedJobsHistoryLimit":1}}'
kubectl patch cronjob llmtrain-cron --type=merge -p '{"spec":{"startingDeadlineSeconds":10}}'
```

**预期**：`MissSchedule` 事件仅在错过调度窗口时产生；历史派生的 llmtrain Job 收敛为最近「1 成功 + 1 失败」。

---

## 5. 失败 / 重试 / 退避用例（LT-R）

### LT-R-01 段内失败 → 终态 Failed（不重试）

`LLMTrain` 的 FSM 语义与 `kubernetes` 不同：**任一段（download/embed/train）返回错误即标记为终态 `Failed`**，不回退、不重试（对比 `kubernetes` 的 Pod 建失败才触发 `backoffStore` 退避）。

> 内置自包含 demo 的 fake minio / trainer 几乎不会返回错误，此路径主要依赖**自定义 provider（真实 MinIO/Ollama/训练平台）**触发；单测覆盖见 `providers/llmtrain/provider_test.go` 的 `TestPipelineFailsOnError`。触发方式：让 `embed` 或 `train` 段的底层客户端返回错误。

**预期**：`status.providerStatus.phase=Failed`、`message` 记录具体错误、`status.phase=Failed`、`conditions` 出现 `Failed=True`，且**无后续 reconcile 重试**。

---

### LT-R-02 双重超时：activeDeadlineSeconds 与 jobTimeoutSec

| 超时机制 | 位置 | 判定点 | 结果 |
| --- | --- | --- | --- |
| `activeDeadlineSeconds` | `spec` | 控制器 `sync` 先行判定 | `phase=Failed`、`conditions[Failed].reason=DeadlineExceeded`、`errorMessage` 记录、删除子资源（llmtrain 无） |
| `jobTimeoutSec` | `spec.providerSpec` | provider `Status` 内判定 | `providerStatus.phase=Failed`、`message="... exceeded N seconds"` |

**步骤**：

```bash
kubectl apply -f - <<'EOF'
apiVersion: batch.onex.io/v1beta1
kind: Job
metadata:
  name: llmtrain-deadline
  namespace: default
spec:
  type: LLMTrain
  activeDeadlineSeconds: 10
  providerSpec:
    jobTimeoutSec: 14400
EOF
```

**预期**：约 10s（`activeDeadlineSeconds`）后，先于 provider 内部超时，`phase=Failed`、`conditions[Failed].reason=DeadlineExceeded`。若仅设 `jobTimeoutSec`（LT-J-03）则走 provider 内部超时路径。两者可独立验证。

---

### LT-R-03 派发失败 → 指数退避 → 修复后收敛

**步骤**：非法 `providerSpec` JSON 使 `Start`（`decodeSpec`）失败，观察退避；随后修复。

```bash
kubectl apply -f - <<'EOF'
apiVersion: batch.onex.io/v1beta1
kind: Job
metadata:
  name: llmtrain-retry
  namespace: default
spec:
  type: LLMTrain
  providerSpec: '"not-a-json-object"'
EOF

kubectl get events --field-selector involvedObject.name=llmtrain-retry -w   # StartFailed 反复重试
```

修复：

```bash
kubectl patch job llmtrain-retry --type=merge -p \
  '{"spec":{"providerSpec":{"sourceObject":"llm/test.json"}}}'
kubectl get job llmtrain-retry -w
```

**预期**：`StartFailed` 事件后进入 `backoffStore` 指数退避（基础 5s，`×2^n`，上限 5m）；修复后收敛，`phase` 走 `Pending → Running → Succeeded`。

---

## 6. 快速验证清单

| 编号 | 场景 | 关键断言 |
| --- | --- | --- |
| LT-J-01 | 完整流水线 | `phase` Pending→Running→Succeeded；`providerStatus.phase` 七阶段；无 Pod |
| LT-J-02 | 幂等执行 | 重启后从中断 phase 续跑，不重复执行已完成段 |
| LT-J-03 | provider 超时 | `providerStatus.phase=Failed`、`message` 含 `exceeded` |
| LT-J-04 | providerSpec 默认 | 缺省仍跑通成功 |
| LT-J-05 | 大小写/未知类型 | `unsupported job type`；空 type 走 kubernetes |
| LT-J-06 | suspend | 挂起无 Pod 可删；恢复后流水线重置重跑 |
| LT-C-01 | CronJob 定时派发 | 派生 `name-<时间戳>` Job 独立收敛 |
| LT-C-02 | 并发策略 | Forbid 跳过 / Replace 删旧 / Allow 并发 |
| LT-C-03 | 历史清理 | 保留最近 1 成功 + 1 失败 |
| LT-R-01 | 段内失败 | 终态 Failed，不重试 |
| LT-R-02 | 双超时 | `activeDeadlineSeconds`（DeadlineExceeded）/ `jobTimeoutSec`（exceeded N seconds） |
| LT-R-03 | 派发失败退避 | StartFailed → 指数退避 → 修复收敛 |