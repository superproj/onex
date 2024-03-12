# naivechain
A naive and simple implementation of blockchains.

### Build And Run

- Download and compile **go get -v github.com/kofj/naivechain**

- Start First Node

  ```bash
  naivechain -peers ""
  ```

- Start Second Node

  ```bash
  naivechain -api :3002 -p2p :6002 -peers ws://localhost:6001
  ```



### HTTP API

- query blocks

  ```
  curl http://localhost:3001/blocks

  ```

- mine block

  ```
  curl -H "Content-type:application/json" --data '{"data" : "Some data to the first block"}' http://localhost:3001/mine_block

  ```

- add peer

  ```
  curl -H "Content-type:application/json" --data '{"peer" : "ws://localhost:6002"}' http://localhost:3001/add_peer

  ```

- query peers

  ```
  curl http://localhost:3001/peers
  ```