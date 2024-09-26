# Kubernetes 集群部署指南

## 安装 Docker

Kubernetes 底层需要一个容器运行时来部署 Pod。所以，首先我们需要部署一个容器运行时，这里我们使用 containerd 作为 Kubernetes 的容器运行时。

安装步骤分为以下几步：
1. 安装 docker 前置条件检查；
2. 安装 docker；
3. docker 安装后配置。

### 1. 安装 docker 前置条件检查。

需要确保 CentOS 系统启用了 `centos-extras` yum 源，默认情况下已经启用，检查方式如下：

```bash
$ cat /etc/yum.repos.d/CentOS-Extras.repo
# Qcloud-Extras.repo


[extras]
name=Qcloud-$releasever - Extras
baseurl=http://mirrors.tencentyun.com/centos/$releasever/extras/$basearch/os/
gpgcheck=1
enabled=1
gpgkey=file:///etc/pki/rpm-gpg/RPM-GPG-KEY-Qcloud-8
```

### 2. 安装 docker

Docker 官方文档 [Install Docker Engine on CentOS](https://docs.docker.com/engine/install/centos/) 提供了 3 种安装方法:
- 通过 Yum 源安装；
- 通过 RPM 包安装；
- 通过脚本安装。

这里，我们选择最简单的安装方式：**通过 Yum 源安装**。它具体又分为下面 3 个步骤。

1) 安装 docker。

命令如下：

```bash
$ sudo yum install -y yum-utils # 1. 安装 `yum-utils` 包，该包提供了 `yum-config-manager` 工具
$ sudo yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo # 2. 安装 `docker-ce.repo` yum 源
$ sudo yum-config-manager --enable docker-ce-nightly docker-ce-test # 3. 启用 `nightly` 和 `test` yum 源
$ sudo yum install -y docker-ce docker-ce-cli containerd.io # 4. 安装最新版本的 docker 引擎和 containerd
```

2) 启动 docker。

可以通过以下命令来启动 docker：

```bash
$ sudo systemctl start docker
```

docker 的配置文件是 `/etc/docker/daemon.json`，这个配置文件默认是没有的，需要我们手动创建：

```bash
$ sudo tee /etc/docker/daemon.json << EOF
{
  "bip": "172.16.0.1/24",
  "registry-mirrors": [],
  "graph": "/data/lib/docker"
}
EOF
```

配置参数说明如下：
- `registry-mirrors`：仓库地址，可以根据需要修改为指定的地址。
- `graph`：镜像、容器的存储路径，默认是 `/var/lib/docker`。如果你的 `/` 目录存储空间满足不了需求，需要设置 `graph` 为更大的目录。
- `bip`：指定容器的 IP 网段。

配置完成后，需要重启 docker：

```bash
$ sudo systemctl restart docker
```

3）测试 docker 是否安装成功。

```bash
$ sudo docker run hello-world
Unable to find image 'hello-world:latest' locally
latest: Pulling from library/hello-world
b8dfde127a29: Pull complete
Digest: sha256:0fe98d7debd9049c50b597ef1f85b7c1e8cc81f59c8d623fcb2250e8bec85b38
Status: Downloaded newer image for hello-world:latest
...
Hello from Docker!
This message shows that your installation appears to be working correctly.
....
```

`docker run hello-world` 命令会下载 `hello-world` 镜像，并启动容器，打印安装成功提示信息后退出。

### 3. docker 安装后配置

安装成功后，我们还需要做一些其他配置。主要有两个，一个是配置 docker，使其可通过 non-root 用户使用；另一个是配置 docker 开机启动。

1) 使用non-root用户操作 docker。

我们在 Linux 系统上操作，为了安全，需要以普通用户的身份登录系统并执行操作。所以，我们需要配置 docker，使它可以被 non-root 用户使用。具体配置方法如下：

```bash
$ sudo groupadd docker # 1. 创建 `docker` 用户组
$ sudo usermod -aG docker $USER # 2. 将当前用户添加到 `docker` 用户组下
$ newgrp docker # 3. 重新加载组成员身份
$ docker run hello-world # 4. 确认能够以普通用户使用 docker
```

如果在执行 `sudo groupadd docker` 时报 `groupadd: group 'docker' already exists` 错误，说明 docker 组已经存在了，可以忽略这个报错。

如果你在将用户添加到 docker 组之前，使用 `sudo` 运行过 docker 命令，你可能会看到以下错误：

```bash
WARNING: Error loading config file: /home/user/.docker/config.json -
stat /home/user/.docker/config.json: permission denied
```

这个错误，我们可以通过删除 `~/.docker/` 目录来解决，或者通过以下命令更改 `~/.docker/` 目录的所有者和权限：

```bash
$ sudo chown "$USER":"$USER" /home/"$USER"/.docker -R
$ sudo chmod g+rwx "$HOME/.docker" -R
```

2) 配置 docker 开机启动。

配置命令如下：

```bash
$ sudo systemctl enable docker.service # 设置 docker 开机启动
$ sudo systemctl enable containerd.service # 设置 containerd 开机启动
```

## 安装 Kind 集群

Kind （Kubernetes in Docker）是一个工具。可以在本地快速创建、删除 Kubernetes 集群。Kind 是我用过的所有 Kubernetes 集群管理工具中最简单的一个。它的学习成本最低，对用户操作也最友好。

安装完 docker 之后，接下来我们就要部署一个 Kubernetes 集群。最简单的方式是通过 [kind](https://kind.sigs.k8s.io/) 来创建一个 kind 集群。此步骤又分为以下几步：
1. 安装 `kind` 和 `kubectl` 命令；
2. 创建 kind 集群；
3. 访问 kind 集群。

### 1. 安装 `kind` 和 `kubectl` 命令

安装命令如下：

```bash
$ curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.18.0/kind-linux-amd64
$ chmod +x ./kind
$ sudo mv ./kind /usr/local/bin/kind
$ ./scripts/add-completion.sh kind bash
$ kind version # 验证 kind
kind v0.18.0 go1.20.2 linux/amd64
```

为了访问 Kind 集群，我们需要安装 `kubectl` 命令，安装命令如下：

```bash
$ curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
$ chmod +x kubectl
$ sudo mv kubectl /usr/bin/
$ ./scripts/add-completion.sh kubectl bash
$ kubectl version --client --output=yaml # 验证 kubectl 是否安装成功
clientVersion:
  buildDate: "2023-03-15T13:40:17Z"
  compiler: gc
  gitCommit: 9e644106593f3f4aa98f8a84b23db5fa378900bd
  gitTreeState: clean
  gitVersion: v1.26.3
  goVersion: go1.19.7
  major: "1"
  minor: "26"
  platform: linux/amd64
kustomizeVersion: v4.5.7
```

### 2. 创建 Kind 集群

安装好 `kind` 命令后，就可以通过 `kind` 命令来创建一个 Kubernetes 集群，创建命令如下：

```bash
$ kind create cluster --config=kind-onex.yaml
$ kind get clusters # 查询 Kind 集群列表
onex
```

> 注意：需要根据需要修改 `apiServerAddress` 为宿主机 IP。

上述命令会成功创建一个 Kubernetes 集群，并将 `kubectl` 命令的 `context` 设置为新创建的 Kind 集群。

相应的你可以通过 `kind delete cluster --name=onex` 来删除所创建的集群。

### 3. 访问 Kind 集群

我们通过 `kubectl cluster-info` 来访问新建的 Kind 集群，以验证集群成功创建。访问命令如下：

```bash
$ kubectl config use-context kind-onex
$ kubectl cluster-info 
Kubernetes control plane is running at https://127.0.0.1:16443
CoreDNS is running at https://127.0.0.1:16443/api/v1/namespaces/kube-system/services/kube-dns:dns/proxy

To further debug and diagnose cluster problems, use 'kubectl cluster-info dump'.
```

### 4. 导入本地镜像到 Kind 集群


1. 先构建镜像，例如：构建 onex-usercenter 镜像

```bash
$ make image IMAGES=onex-usercenter
```


2. 导入镜像到 Kind 集群

```bash
$ kind load docker-image --name onex ccr.ccs.tencentyun.com/superproj/onex-usercenter-amd64:v0.1.0
```

`kind load docker-image` 提供的其他有用参数为：
- ``-n, --name string`: 集群上下文名称 (默认为 "kind")；
- `--nodes strings`: 要加载镜像的节点的逗号分隔列表。

> 官方文档：[Loading an Image Into Your Cluster](https://kind.sigs.k8s.io/docs/user/quick-start/#loading-an-image-into-your-cluster)


## 其他 Kind 集群操作

1. 登录 Kind Node

```bash
$ docker exec -it onex-worker bash
root@onex-worker:/# crictl img
```

2. 清理 `onex-worker` 节点的 dangling image

```bash
$ docker exec -it onex-worker bash
# ctr -n k8s.io i rm `ctr -n k8s.io i ls|awk '{print $1}'`
# crictl img ls # 虽然仍然能够看到 dangling 镜像，但是实际上宿主机硬盘空间已经被释放
```
