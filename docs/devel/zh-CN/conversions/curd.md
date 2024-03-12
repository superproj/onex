## 业务 CURD 规范

- K8S资源，例如：MinerSet/Miner具有以下行为：
  - Create：返参是v1beta1.MinerXXX
  - Update: 返参是v1beta1.MinerXXX
  - Get: 返参是v1beta1.MinerXXX
  - List: 入参是ListMinerXXXRequest，返回参数是ListMinerXXXResponse(考虑到性能问题)
  - Delete: ...
