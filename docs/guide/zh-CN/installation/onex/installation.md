# OneX 部署指南

- [onex-usercenter 服务安装](onex-usercenter/installation.md)

## 组件部署顺序

1. onex-usercenter
2. onex-apiserver
3. onex-gateway
4. onex-nightwatch
5. onex-pump
6. onex-controller-manager
7. onex-toyblc
8. onex-minerset-controller
9. onex-miner-controller

## 安装前准备

### 1. 创建必要的目录


```bash
$ mkdir -p $HOME/.onex
$ sudo mkdir -p /opt/onex/conf
$ export HOSTIP=x.x.x.x
```

> 假设：你 Kind 集群宿主机 IP 地址为：`x.x.x.x`

### 2. 准备本地文件

```bash
$ make gen.kubeconfig
$ sudo cp -a _output/cert /opt/onex # 组件配置文件中使用，以使本地配置文件和deployment中路径保持一致，便于维护
$ sudo cp -a _output/config /opt/onex/conf/config # 组件配置文件中使用，以使本地配置文件和deployment中路径保持一致，便于维护
$ cp -a _output/config $HOME/.onex/config # 本地访问（kubectl --kubeconfig $HOME/.onex/config get ms）
```

### 3. 创建共用 K8S 资源

```bash
$ kubectl create namespace onex
$ kubectl -n onex create secret generic onex-tls --from-file=_output/cert
$ kubectl -n onex create configmap onex --from-file=_output/config
```

### 4. 配置 hosts

```bash
$ su - root
# cat << EOF >> /etc/hosts
127.0.0.1 onex.usercenter.superproj.com
127.0.0.1 onex.gateway.superproj.com
127.0.0.1 onex.apiserver.superproj.com
127.0.0.1 onex.controllermanager.superproj.com
127.0.0.1 onex.nightwatch.superproj.com
127.0.0.1 onex.miner.superproj.com
127.0.0.1 onex.minerset.superproj.com
127.0.0.1 onex.toyblc.superproj.com
EOF
```

## 按顺序安装组件

前置操作：                                    

``bash
$ export ONEX_ROOT=$GOPATH/src/github.com/superproj/onex
$ cd ${ONEX_ROOT}
`

组件安装：

1. [onex-usercenter](./onex-usercenter/installation.md)
2. [onex-apiserver](./onex-apiserver/installation.md)
3. [onex-gateway](./onex-gateway/installation.md)
4. [onex-nightwatch](./onex-nightwatch/installation.md)
5. [onex-pump](./onex-pump/installation.md)
6. [onex-controller-manager](./onex-controller-manager/installation.md)
7. [onex-toyblc](./onex-toyblc/installation.md)
8. [onex-minerset-controller](./onex-minerset-controller/installation.md)
9. [onex-miner-controller](./onex-miner-controller/installation.md)

## FAQ

### 1. 为什么选择为整个 OneX 应用部署一个 Ingress，而不是每个服务部署一个 Ingress?

这样做主要考虑到将所有的 OneX 路由规则配置到一个 Ingress 中，既方便管理，又能在一定程度上避免路由冲突。

当然，如果集群中有多个微服务，并且每个微服务都需要公开自己的端点，则创建多个 Ingress 对象可能是有意义的选择。这样，每个服务都可以有自己的路由规则。

