# Jaeger 部署指南

官方安装文档：https://www.jaegertracing.io/docs/1.43/getting-started/

常用安装方式如下：
- 使用 Docker 进行安装（测试开发时可用）
- Operator安装（适合生产环境使用）

## 使用 Docker 进行安装

安装命令如下：

```bash
docker run -d --name jaeger \
  -e COLLECTOR_ZIPKIN_HOST_PORT=:9411 \
  -e COLLECTOR_OTLP_ENABLED=true \
  -p 6831:6831/udp \
  -p 6832:6832/udp \
  -p 5778:5778 \
  -p 16686:16686 \
  -p 4317:4317 \
  -p 4318:4318 \
  -p 14250:14250 \
  -p 14268:14268 \
  -p 14269:14269 \
  -p 9411:9411 \
  jaegertracing/all-in-one:1.43
```

安装后，可访问 Jaeger UI：`http://127.0.0.1:16686`

## Operator 安装

安装文档参考：https://www.jaegertracing.io/docs/1.43/operator/
