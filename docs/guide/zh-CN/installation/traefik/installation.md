## Traefik 部署指南

官方部署文档：https://doc.traefik.io/traefik/getting-started/install-traefik/#use-the-helm-chart

## 操作（Helm部署）

Helm Release 版本：v2.9.9
Helm Chart：traefik-22.0.0

### 1. 安装

安装命令如下：

```bash
helm repo add traefik https://traefik.github.io/charts
helm repo update
helm install traefik traefik/traefik --namespace kube-system --values traefik-values.yaml
kubectl port-forward -n kube-system --address 0.0.0.0 $(kubectl get pods -n kube-system --selector "app.kubernetes.io/name=traefik" --output=name) 7000:9000
```

#### traefik-values.yaml 配置解析

- asDefault: 如果一个服务没有明确指定入口点，那么启用此入口点作为默认入口点；
- port: traefik 后端服务监听的端口;
- hostPort: hostPort 是将 pod 的端口映射到宿主机上；
- hostIP: traefik 后端服务监听端口；
- expose:  将入口点公开到外部网络；
- exposedPort: Kubernetes 集群中 traefik 服务的端口；
- nodePort: nodePort 是将 service 的端口映射到集群中的每个宿主机上；
### 2. 访问 dashbaord

浏览器访问 `http://127.0.0.1:7000/dashboard/` 即可

部署后的适配 onex 组件的 traefik deployment请参考：[depoyed-deployment.yaml](depoyed-deployment.yaml)


## 使用

部署测试程序：

```bash
kubectl apply -f whoami.yaml -f whoami-services.yaml -f whoami-ingress.yaml
```

访问：

```bash
curl http://127.0.0.1:18080
```

网络流量为：kind host's hostPort -> kind node's hostPort -> traefik pod expose port。注意：流量没有经过 traefik service.（TODO：完善走 service 方案）

> 提示：这里的 `18080` 端口是在 [Kubernetes 集群部署](../kubernetes/kubernetes.md) `extraPortMappings[0].hostPort` 部分设置的。

## 注意

traefik deployment args 中需要添加以下参数：
```bash
- --serversTransport.insecureSkipVerify=true
```

否则会出现以下错误：

```bash
500 Internal Server Error' caused by: tls: failed to verify certificate: x509: certificate is valid for 127.0.0.1
```
