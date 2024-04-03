# onex-minerset-controller 部署指南（容器化）

1. 创建 ConfigMap

```bash
$ cp configs/onex-minerset-controller.yaml $HOME/.onex/onex-minerset-controller.yaml
$ kubectl -n onex create configmap onex-minerset-controller --from-file $HOME/.onex/onex-minerset-controller.yaml
```

> 注意：创建前，记得修改 `onex-minerset-controller.yaml` 中相关配置，例如：访问地址、密码等。

2. 创建 Workload

```bash
$ kubectl -n onex apply -f deployments/onex/onex-minerset-controller
```

3. 测试是否部署成功

```bash
$ curl -H "Host: onex.minerset.superproj.com" http://127.0.0.1:18080/healthz
```
