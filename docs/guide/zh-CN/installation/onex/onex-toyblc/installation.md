# onex-toyblc 部署指南（容器化）

1. 创建 ConfigMap

```bash
$ cp configs/onex-toyblc.yaml $HOME/.onex/onex-toyblc.yaml
$ kubectl -n onex create configmap onex-toyblc --from-file $HOME/.onex/onex-toyblc.yaml
```

> 注意：创建前，记得修改 `onex-toyblc.yaml` 中相关配置，例如：访问地址、密码等。

2. 创建 Workload

```bash
$ kubectl -n onex apply -f deployments/onex/onex-toyblc
```

3. 测试是否部署成功

```bash
$ curl -H "Host: onex.toyblc.superproj.com" http://127.0.0.1:18080/healthz
```

## 使用

### 1. 查询 peers

```bash
$ curl -H "Host: onex.toyblc.superproj.com" http://127.0.0.1:18080/v1/peers
```

### 2. 查询 blocks

```bash
$ curl -H "Host: onex.toyblc.superproj.com" http://127.0.0.1:18080/v1/blocks
```

> curl http://genesis.kube-system.svc.superproj.com:8080/v1/blocks

### 3. 挖矿 

```bash
$ curl -XPOST -H "Host: onex.toyblc.superproj.com" -H"Content-type: application/json" -d'{"data": "Some data to the first block"}' http://127.0.0.1:18080/v1/blocks
```
