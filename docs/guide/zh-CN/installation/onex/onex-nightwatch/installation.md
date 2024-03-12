# onex-nightwatch 部署指南（容器化）

1. 创建 ConfigMap

```bash
$ sed "s/127.0.0.1/$HOSTIP/g" configs/onex-nightwatch.yaml > $HOME/.onex/onex-nightwatch.yaml
$ kubectl -n onex create configmap onex-nightwatch --from-file $HOME/.onex/onex-nightwatch.yaml
```

> 注意：创建前，记得修改 `onex-nightwatch.yaml` 中相关配置，例如：访问地址、密码等。

2. 创建 Workload

```bash
$ kubectl -n onex apply -f deployments/onex/onex-nightwatch
```

3. 测试是否部署成功

```bash
$ curl -H "Host: onex.nightwatch.superproj.com" http://127.0.0.1:18080/healthz
```
