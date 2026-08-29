# onex-controller-manager 编译/运行修复报告

本报告记录了 `onex-controller-manager` 组件从编译到运行、再到功能验证的完整过程，以及过程中发现并修复的问题。

## 1. 编译

编译命令：

```bash
make build BINS=onex-controller-manager
```

编译输出（首次即成功，无需修复编译问题）：

```
===========> Building binary onex-controller-manager v0.2.1-11-gb56fecd-dirty for linux amd64
```

产物路径：

```
_output/platforms/linux/amd64/onex-controller-manager
```

> 说明：本机为 `linux/amd64`，因此实际产物路径为 `_output/platforms/linux/amd64/`，而非任务描述中的
> `_output/platforms/darwin/arm64/`（后者是 macOS ARM 平台的路径，仅在不做跨平台编译时才会由 `make build`
> 按宿主平台生成对应目录）。

## 2. 运行

运行命令（按本机平台调整产物路径）：

```bash
_output/platforms/linux/amd64/onex-controller-manager \
  --kubeconfig _output/config \
  --mysql-database=onex \
  --mysql-host=43.139.4.14:3306 \
  --mysql-username=onex \
  --mysql-password='onex(#)666'
```

## 3. 发现并修复的问题

首次运行失败，报错如下：

```
E0829 ... run.go:72] "command failed" err="failed to start metrics server: failed to create listener: listen tcp :8080: bind: address already in use"
...
"Warning: controller is disabled" controller="garbage-collector-controller"
"Warning: controller is disabled" controller="namespaced-resource-deleter-controller"
```

共定位到 3 个问题，修复如下。

### 3.1 修复 metrics 服务器端口绑定错误（导致启动失败的直接原因）

**问题**：`ctrl.NewManager` 中 `Metrics` 配置被注释掉了，导致 controller-runtime 使用默认的
`:8080` 地址启动 metrics 服务器。而宿主机 8080 端口已被占用，导致 metrics 服务器绑定失败、
manager 启动失败、进程退出。

**修复**：将 metrics 服务器绑定地址接入组件配置 `MetricsBindAddress`（默认 `127.0.0.1:20251`）。

文件：`cmd/onex-controller-manager/app/controllermanager.go`

```go
mgr, err := ctrl.NewManager(c.Kubeconfig, ctrl.Options{
	Scheme: scheme,
	Metrics: metricsserver.Options{
		BindAddress: c.ComponentConfig.Generic.MetricsBindAddress,
	},
	LeaderElection:             c.ComponentConfig.Generic.LeaderElection.LeaderElect,
	...
})
```

同时新增导入：

```go
metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
```

> 注意：controller-runtime v0.24.1 的 `ctrl.Options.Metrics` 类型为 `metricsserver.Options`
> （结构体），而不是字符串，因此不能直接写成 `Metrics: c.ComponentConfig.Generic.MetricsBindAddress`。

### 3.2 修复默认控制器列表为空导致所有控制器被禁用

**问题**：`RecommendedDefaultGenericControllerManagerConfiguration` 缺少对 `Controllers` 的默认值设置。
上游 `k8s.io/controller-manager` 在 `len(obj.Controllers) == 0` 时会默认 `Controllers = []string{"*"}`
（启用所有默认开启的控制器）。onex 版本漏掉了这行，导致 `Generic.Controllers` 为空，而
`IsControllerEnabled` 在列表为空且无 `*` 时返回 `false`，于是 `garbage-collector-controller` 和
`namespaced-resource-deleter-controller` 全部被禁用。

**修复**：补齐默认值。

文件：`pkg/config/v1beta1/defaults.go`

```go
// Enable all on-by-default controllers by default.
if len(obj.Controllers) == 0 {
	obj.Controllers = []string{"*"}
}
```

该函数同时被 `onex-controller-manager`、区块链（blockchain）控制器、job 控制器的配置所复用，
补齐默认值与上游行为一致，对其它控制器管理器同样正确。

### 3.3 清理遗留的调试输出

**问题**：`pkg/config/options/generic.go` 的 `ApplyTo` 中残留了一行调试 `fmt.Println`：

```go
fmt.Println("111111111111111111111111-999", cfg.Controllers)
```

每次启动都会向标准输出打印一行 `111111111111111111111111-999 []`。

**修复**：删除该调试输出。

文件：`pkg/config/options/generic.go`

## 4. 修复后的运行结果

重新编译并运行后，启动日志如下（关键行）：

```
I0829 ... "Register controller" controller="garbage-collector-controller"
I0829 ... "Register controller" controller="namespaced-resource-deleter-controller"
I0829 ... "Serving metrics server" ... bindAddress="127.0.0.1:20251" secure=false
I0829 ... "Successfully acquired lease" lock="kube-system/onex-controller-manager"
I0829 ... "Starting EventSource" controller="controller-manager.namespace" ...
I0829 ... "Garbage collector: all resource monitors have synced"
I0829 ... "Proceeding to collect garbage"
I0829 ... "Starting workers" controller="controller-manager.namespace" ... worker count=16
```

两个控制器均已正常注册并运行，metrics 服务监听 `127.0.0.1:20251`，健康检查监听 `:20250`。

健康检查 / metrics 端点验证：

```bash
$ curl -s http://127.0.0.1:20250/readyz   # ok
$ curl -s http://127.0.0.1:20250/healthz  # ok
$ curl -s http://127.0.0.1:20251/metrics  # 正常输出 Prometheus 指标
```

> 关于启动时的一行 `E0829 ... The manifest file is empty, ignoring.`：这是 `k8s.io/component-base`
> `metrics.Options.Apply()` 在未配置 `--allow-metric-labels-manifest` 时的标准提示，kube-controller-manager
> 同样会输出，属良性告警，非 onex 缺陷，无需处理。

## 5. 功能验证（各 controller）

`onex-controller-manager` 当前注册并内置了 2 个控制器（见 `cmd/onex-controller-manager/README.md`）：

- `garbage-collector-controller`：垃圾回收（GC）
- `namespaced-resource-deleter-controller`：命名空间清理

### 5.1 namespaced-resource-deleter-controller

**验证场景**：在自定义命名空间中创建资源（包括核心资源 ConfigMap 与自定义 onex 资源 Miner），
删除命名空间，验证控制器能清理命名空间内所有资源并移除 finalizer，使命名空间最终被完整删除。

```bash
kubectl --kubeconfig _output/config create ns del-test
kubectl --kubeconfig _output/config -n del-test create configmap cm1 --from-literal=a=b
kubectl --kubeconfig _output/config apply -f - <<'EOF'
apiVersion: apps.onex.io/v1beta1
kind: Miner
metadata:
  name: test-miner
  namespace: del-test
spec:
  displayName: testminer
  chainName: genesis
  minerType: M1.MEDIUM2
EOF

kubectl --kubeconfig _output/config delete ns del-test --timeout=30s
# namespace "del-test" deleted
kubectl --kubeconfig _output/config get ns del-test
# Error from server (NotFound): namespaces "del-test" not found
```

**结果**：✅ 通过。命名空间内的 ConfigMap 与自定义 Miner 资源均被清理，finalizer 被移除，命名空间完整删除
（未停留在 Terminating 状态）。

### 5.2 garbage-collector-controller

**验证场景**：创建一个作为 owner 的 ConfigMap，再创建一个带 `ownerReferences` 指向该 ConfigMap 的 Secret，
删除 owner 后验证 GC 能够回收“孤儿”的依赖对象。

```bash
kubectl --kubeconfig _output/config create ns gc-test
kubectl --kubeconfig _output/config -n gc-test create configmap owner-cm --from-literal=a=b
CM_UID=$(kubectl --kubeconfig _output/config -n gc-test get cm owner-cm -o jsonpath='{.metadata.uid}')

kubectl --kubeconfig _output/config apply -f - <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: dependent-secret
  namespace: gc-test
  ownerReferences:
  - apiVersion: v1
    kind: ConfigMap
    name: owner-cm
    uid: $CM_UID
type: Opaque
stringData:
  k: v
EOF

kubectl --kubeconfig _output/config -n gc-test delete cm owner-cm
sleep 5
kubectl --kubeconfig _output/config -n gc-test get secret dependent-secret
# Error from server (NotFound): secrets "dependent-secret" not found
```

**结果**：✅ 通过。删除 owner ConfigMap 后，带 ownerReference 的依赖 Secret 被 GC 自动回收。

### 5.3 指标佐证

```bash
$ curl -s http://127.0.0.1:20251/metrics | grep -E 'reconcile_errors_total|reconcile_total\{'
controller_runtime_reconcile_errors_total{controller="controller-manager.namespace"} 0
controller_runtime_reconcile_total{controller="controller-manager.namespace",result="success"} 12
controller_runtime_reconcile_errors_total  …  0（无任何 reconcile 错误）
```

## 6. 修改文件清单

| 文件 | 修改内容 |
| --- | --- |
| `cmd/onex-controller-manager/app/controllermanager.go` | 接入 metrics 服务器绑定地址，新增 `metricsserver` 导入 |
| `pkg/config/v1beta1/defaults.go` | 补齐 `Controllers` 默认值为 `["*"]` |
| `pkg/config/options/generic.go` | 移除遗留调试输出 `fmt.Println` |

## 7. 结论

`onex-controller-manager` 现可正常编译、正常启动并稳定运行（无报错），内置的
`garbage-collector-controller` 与 `namespaced-resource-deleter-controller` 两个控制器功能均验证通过。
