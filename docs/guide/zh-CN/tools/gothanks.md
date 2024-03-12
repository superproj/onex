## gothanks 使用指南

GitHub Repo: https://github.com/psampaz/gothanks

1. 安装

```bash
$ go install github.com/psampaz/gothanks@latest
```

或者：

```bash
$ make  tools.install.gothanks
```

2. 使用

```bash
$ gothanks -github-token=xxxxxx
```

或者

```bash
$ export GITHUB_TOKEN=xxxxx
$ gothanks
```

3. 示例


```bash
$ ./gothanks -y
Welcome to GoThanks :)

Sending your love..

Repository github.com/golang/go is already starred!
Repository github.com/google/go-github is already starred!
Repository github.com/sirkon/goproxy is already starred!
Repository github.com/golang/crypto is already starred!
Repository github.com/golang/net is already starred!
Repository github.com/golang/oauth2 is already starred!

Thank you!
```
