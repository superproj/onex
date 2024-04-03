# Kafka 部署指南

安装文档：https://kafka.apache.org/documentation/#quickstart

## 手动安装

```bash
wget https://dlcdn.apache.org/kafka/3.3.1/kafka_2.13-3.3.1.tgz
$ tar -xzf kafka_2.13-3.3.1.tgz
$ cd kafka_2.13-3.3.1
```

## Docker 安装（推荐）

Docker安装文档：
- 可用：http://events.jianshu.io/p/b60afa35303a
- 参考：https://towardsdatascience.com/how-to-install-apache-kafka-using-docker-the-easy-way-4ceb00817d8b

1. 下载镜像

```bash
docker pull wurstmeister/zookeeper  
docker pull wurstmeister/kafka
```

2. 启动zookeeper

```bash
docker run -d --name zookeeper -p 2181:2181 -t wurstmeister/zookeeper
```

3. 启动kafka

```bash
docker run -d --name kafka --publish 9092:9092 --link zookeeper --env KAFKA_ZOOKEEPER_CONNECT=zookeeper:2181 --env KAFKA_ADVERTISED_HOST_NAME=127.0.0.1 --env KAFKA_ADVERTISED_PORT=9092 wurstmeister/kafka
```
4. 创建主题

```bash
docker exec -it kafka /bin/bash
cd opt/kafka_2.12-2.3.0/
bin/kafka-topics.sh --create --zookeeper zookeeper:2181 --replication-factor 1 --partitions 1 --topic mykafka
```

5. 启动消息发送方

```bash
docker exec -it kafka /bin/bash
cd opt/kafka_2.12-2.3.0/
./bin/kafka-console-producer.sh --broker-list localhost:9092 --topic mykafka
```

6. 启动消息接收方

```bash
docker exec -it kafka /bin/bash
cd opt/kafka_2.12-2.3.0/
./bin/kafka-console-consumer.sh --bootstrap-server localhost:9092 --topic mykafka --from-beginning
```

7. 主题列表

```bash
docker exec -it kafka /bin/bash
cd opt/kafka_2.12-2.3.0/
bin/kafka-topics.sh --list --zookeeper zookeeper:2181
```

8. 查看topic的状态

```bash
docker exec -it kafka /bin/bash
cd opt/kafka_2.12-2.3.0/
bin/kafka-topics.sh --describe --zookeeper zookeeper:2181 --topic mykafka
```

9. 安装 Kafka 客户端工具

仓库地址：https://github.com/deviceinsight/kafkactl

```bash
go install github.com/deviceinsight/kafkactl@latest
```

> 提示：`kafkactl` 工具默认配置文件：`$HOME/.config/kafkactl/config.yml`
