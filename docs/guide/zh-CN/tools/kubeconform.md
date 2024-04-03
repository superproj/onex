## kubeconform 使用指南

kubeconform 是一个 Kubernetes 清单验证工具。可以将其合并到您的 CI 中，或在本地使用它来验证您的 Kubernetes 配置！

> onex 暂时没用到；核心资源都通过helm来安装，helm chart 可以通过 `kube-lint` 来检查。

### 安装

```bash
$ go install github.com/yannh/kubeconform/cmd/kubeconform@latest
```

或者：

```bash
$ make tools.install.kubeconform
```

### 使用

参考：https://github.com/yannh/kubeconform

