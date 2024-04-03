## 日志规范

### 常用日志规范

- 为了跟K8S组件、生态兼容，所有日志均使用k8s.io/klog/v2
- 最原始 报错位置打印一条日志即可
- 统一使用 `k8s.io/klog/v2` 记录日志
- 使用klog结构化日志方法记录日志：`klog.InfoS`, `klog.ErrorS`. Example: `klog.InfoS("Received HTTP request", "method", "GET", "URL", "/metrics", "latency", time.Second)`
- 日志均以大写开头，结尾不跟 `.`，例如：`klog.InfoS("Received HTTP request")`
- 使用过去时，例如：`Could not delete B` 而不是 `Cannot delete B`
- 当提到一个对象时，说明它是什么类型的对象。例如：`Deleted pod` 而不是 `Deleted`
- 遵循日志级别规范：
  - 所有错误统一使用`klog.ErrorS`
  - Warning级别的日志使用 `klog.V(1).InfoS`
  - Info级别的日志使用 `klog.V(2).InfoS`
  - Debug级别的日志使用 `klog.V(4/5/6).InfoS`
    - Debug时，经常需要查看的日志使用 `klog.V(4).InfoS`
    - Debug时，不经常需要查看的日志使用 `klog.V(5).InfoS`
    - Debug时，偶尔需要查看的日志使用 `klog.V(6).InfoS`

> Notice: klog.V(0).InfoS = klog.InfoS


### 日志 `-v` 设置

- 生产环境：`-v=2`
- 开发测试环境：`-v=4`

### 其他规范可参考（recommended to follow）

- [Kubernetes Logging Conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/logging.md)
- [Structured and Contextual Logging migration instructions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-instrumentation/migration-to-structured-logging.md#structured-logging-in-kubernetes)
