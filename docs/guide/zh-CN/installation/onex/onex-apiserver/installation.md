# onex-apiserver 部署指南（容器化）

1. 创建 Workload

```bash
$ sed "s/127.0.0.1/$HOSTIP/g" deployments/onex/onex-apiserver/*|kubectl -n onex apply -f -
```

2. 测试是否部署成功

```bash
$ kubectl -s https://onex.apiserver.superproj.com:18443 --kubeconfig=$HOME/.onex/config get ms
```
