# 项目笔记

> 注意：所有操作均在 `${ONEX_ROOT}` 目录下进行。

本项目所有密码均为：`onex(#)666`。

## 组件名、数据库、密码

为了能够追踪请求源，方便以后排障，这里针对每一个组件申请一个账号：

| 组件名 | 用户名 | 密码 | 备注 |
| ----------- | ----------- | ----------- | ----------- |
| - | onex | onex(#)666 | root account |
| onex-usercenter | usercenter | onex(#)666 | |
| onex-gateway | gateway | onex(#)666 | |
| onex-apiserver | apiserver | onex(#)666 | |
| onex-nightwatch | nightwatch | onex(#)666 | |

## 启动 GitBook

```bash
$ cd docs
$ gitbook serve # 访问 http://127.0.0.1:4000/
```

## 组件启动命令

1. 命令行直接启动

```bash
_output/platforms/linux/amd64/onex-usercenter --db.host=127.0.0.1 --db.username=onex --db.password='onex(#)666' --db.database=onex --redis.addr=127.0.0.1:6379 --redis.password='onex(#)666' --etcd.endpoints=127.0.0.1:2379 --kafka.brokers=localhost:9092 --http.addr=0.0.0.0:38443 --grpc.addr=0.0.0.0:39090

_output/platforms/linux/amd64/onex-usercenter --db.host=127.0.0.1 --db.username=onex --db.password='onex(#)666' --db.database=onex --redis.addr=127.0.0.1:6379 --redis.password='onex(#)666' --etcd.endpoints=127.0.0.1:2379 --kafka.brokers=localhost:9092 --http.addr=0.0.0.0:38443 --grpc.addr=0.0.0.0:39090 --tls.use-tls=true --tls.cert=/home/colin/.onex/cert/onex-usercenter.pem --tls.key=/home/colin/.onex/cert/onex-usercenter-key.pem

_output/platforms/linux/amd64/onex-gateway --db.host=127.0.0.1 --db.username=onex --db.password='onex(#)666' --db.database=onex --etcd.endpoints=127.0.0.1:2379 --insecure.addr=0.0.0.0:38443 --secure.bind-address=0.0.0.0 --secure.bind-port=39090 --grpc.addr=0.0.0.0:51020 --kubeconfig=$HOME/.onex/config --usercenter.server=127.0.0.1:38443

_output/platforms/linux/amd64/onex-apiserver --etcd-servers=127.0.0.1:2379 --secure-port=31443 --bind-address=0.0.0.0 --client-ca-file=/home/colin/.onex/cert/ca.pem --tls-cert-file=/home/colin/.onex/cert/onex-apiserver.pem --tls-private-key-file=/home/colin/.onex/cert/onex-apiserver-key.pem

 _output/platforms/linux/amd64/onex-controller-manager --kubeconfig /home/colin/.onex/config --mysql-database=onex --mysql-host=127.0.0.1:3306 --mysql-username=onex --mysql-password='onex(#)666'

_output/platforms/linux/amd64/onex-controller-manager --kubeconfig ~/.onex/config --config configs/onex-controller-manager.yaml

_output/platforms/linux/amd64/onex-nightwatch --kubeconfig /home/colin/.onex/config --db.host=127.0.0.1 --db.username=onex --db.password='onex(#)666' --db.database=onex --redis.addr=127.0.0.1:6379 --redis.password='onex(#)666' --redis.database=1
_output/platforms/linux/amd64/onex-nightwatch --config ~/.onex/onex-nightwatch.yaml

```

2. Kubernetes 部署


## 其他

1. 转发

```bash
kubectl port-forward -n kube-system --address 0.0.0.0 $(kubectl get pods -n kube-system --selector "app.kubernetes.io/name=traefik" --output=name) 8000:9000
kubectl port-forward -n onex --address 0.0.0.0 services/onex-apiserver 52010:https
```

机器学习资料：
https://github.com/microsoft/AI-System
https://openmlsys.github.io/index.html
