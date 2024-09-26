# TBB 区块链系统搭建指南

## 启动创世节点

```
$ ./scripts/get-genesis-account.sh $HOME/.tbb/genesis
$ tbb run --datadir=$HOME/.tbb/genesis --disable-ssl --ip=0.0.0.0 --port=8080 --bootstrap-ip=0.0.0.0 --bootstrap-port=8080 --miner=0x210d9eD12CEA87E33a98AA7Bcb4359eABA9e800e
```

> 当--ip = --bootstrap-ip & --port = --bootstrap-port，tbb会认为这是一个genesis节点。参考：`github.com/web3coach/the-blockchain-bar/node/sync.go.doSync()`


## 创建节点 X

```bash
$ tbb wallet new-account --datadir=$HOME/.tbb/nodex
$ tbb run --datadir=$HOME/.tbb/nodex --disable-ssl --ip=0.0.0.0 --port=8081 --bootstrap-ip=127.0.0.1 --bootstrap-port=8080 --miner=<address>
```
- *<account>:* `tbb wallet new-account`命令生成的account

## 创建节点 Y

```bash
$ tbb wallet new-account --datadir=$HOME/.tbb/nodey
$ tbb run --datadir=$HOME/.tbb/nodey --disable-ssl --ip=0.0.0.0 --port=8081 --bootstrap-ip=127.0.0.1 --bootstrap-port=8080 --miner=<address>
```
- *<account>:* `tbb wallet new-account`命令生成的account

## 转账触发挖矿

```bash
$ curl -XPOST http://127.0.0.1:8080/tx/add -d'{"from":"0x210d9eD12CEA87E33a98AA7Bcb4359eABA9e800","to":"0x210d9eD12CEA87E33a98AA7Bcb4359eABA9e800","value":10,"from_pwd":"onex(#)666"}' -H 'Accept: application/json'
```

## 查看钱包余额

```bash
$ curl -XPOST http://127.0.0.1:8080/balances/list
```
