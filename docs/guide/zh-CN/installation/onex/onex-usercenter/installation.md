# onex-usercenter 部署指南

本安装指南，包含以下 3 中安装方式：
- 终端启动
- Systemd 托管
- Kubernetes YAML 安装
- Helm Chart 安装

## 1. 终端启动

**通过命令行参数启动**

```bash
$ _output/platforms/linux/amd64/onex-usercenter --db.host=127.0.0.1 --db.username=onex --db.password='onex(#)666' --db.database=onex --redis.addr=127.0.0.1:6379 --etcd.endpoints=127.0.0.1:2379 --kafka.brokers=localhost:9092 --http.addr=0.0.0.0:38443 --grpc.addr=0.0.0.0:39090
```

**通过配置文件启动**

```bash
_output/platforms/linux/amd64/onex-usercenter --config configs/onex-usercenter.yaml
```

## 2. Systemd 托管


## Kubernetes YAML 安装（推荐）

1. 创建 ConfigMap

```bash
$ sed "s/127.0.0.1/$HOSTIP/g" configs/onex-usercenter.yaml > $HOME/.onex/onex-usercenter.yaml
$ kubectl -n onex create configmap onex-usercenter --from-file $HOME/.onex/onex-usercenter.yaml
```

> 注意：创建前，记得修改 `onex-usercenter.yaml` 中相关配置，例如：访问地址、密码等。

2. 创建 Workload

```bash
$ kubectl -n onex apply -f deployments/onex/onex-usercenter
```

3. 测试是否部署成功

```bash
$ curl -H "Host: onex.usercenter.superproj.com" http://127.0.0.1:18080/metrics
```

## Helm Chart 安装

<TODO>
