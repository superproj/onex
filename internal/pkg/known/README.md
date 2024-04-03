因为是大仓，并且每个子项目的 known 列表规模无法预估，所以可能存在某个子项目的 known 列表很大。

为了规范 `internal/pkg/known` 目录下的 known 列表，规定 `internal/pkg/known` 目录存放内容遵循以下规范：

- `internal/pkg/known` 存放更应用相关的，通用的 known 列表，例如：`annotations.go`、`labels.go`；
- `internal/pkg/known/{apiserver,usercenter}` 存放 `apiserver`、`usercenter` 私有的 known 列表。



这里没有采用 `internal/pkg/known/quota.go`、`internal/pkg/known/usercenter.go` 这种组织方式，原因如下：
1. `quota.go` 可能被多个项目使用，导致里面内容杂乱无章，难以阅读；
2. `internal/pkg/known/usercenter.go` 起到了文件级别的物理隔离，但 `usercenter` 项目其实还有很多其他 known 列表，不适合全集中在 `usercenter.go` 文件中。如果单独存放一个文件，文件名可能会跟其他项目重合；
3. 所有项目的 known 列表都放在 `internal/pkg/known` 中，每个子项目随意创建 known 列表会造成 `internal/pkg/known` 难以阅读；不随意，又丧失了自由定制的优势。
