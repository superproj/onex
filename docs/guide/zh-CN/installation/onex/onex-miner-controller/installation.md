# onex-miner-controller 部署指南（容器化）

1. 创建 ConfigMap

```bash
$ sed "s/127.0.0.1/$HOSTIP/g" configs/onex-miner-controller.yaml > $HOME/.onex/onex-miner-controller.yaml
$ kubectl -n onex create configmap onex-miner-controller --from-file $HOME/.onex/onex-miner-controller.yaml --from-file config.kind=$HOME/.kube/config
```

> 注意：创建前，记得修改 `onex-miner-controller.yaml` 中相关配置，例如：访问地址、密码等。

2. 创建 Workload

```bash
$ kubectl -n onex apply -f deployments/onex/onex-miner-controller
```

3. 测试是否部署成功

```bash
$ curl -H "Host: onex.miner.superproj.com" http://127.0.0.1:18080/healthz
```
