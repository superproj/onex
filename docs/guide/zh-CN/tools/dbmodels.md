## 如何生成数据库 Model

流程：
1. 创建数据库表
2. 使用工具生成 GORM Model


### 使用 `gen-onex-gorm-model` 工具

```bash
$ make build BINS=gen-onex-gorm-model
$ _output/platforms/linux/amd64/gen-onex-gorm-model -a 127.0.0.1:3306 -u onex -p 'onex(#)666' -d onex --model-pkg-path=/tmp/model
$ ls /tmp/model/
miner.gen.go  minerset.gen.go  secret.gen.go  user.gen.go
```

之后，可以将 `miner.gen.go`, `minerset.gen.go`, `secret.gen.go` 等拷贝到需要的目录。例如：

```bash
$ mkdir -p internal/usercenter/model
$ cp /tmp/model/{uc_secret.gen.go,uc_user.gen.go} internal/usercenter/model
$ mkdir -p internal/gateway/model
$ cp /tmp/model/{api_chain.gen.go,api_miner.gen.go,api_minerset.gen.go} internal/gateway/model
```

> 注意：如果需要新增导出的表，需要修改 `cmd/gen-onex-gorm-model/gen_onex_gorm_model.go` 文件。
> GORM 的 `gen` 包/工具会将 `int(x)` 类型导出为 `int32`，将 `bigint(x)` 导出为 `int64`。



### 使用 `db2struct` 工具（不推荐）

1. 安装

```bash
$ make tools.install.db2struct
```

2. 生成 Model
```bash
$ db2struct --gorm --json -H 127.0.0.1 -u onex -p 'onex(#)666' -d onex -t secret --struct=SecretM
package newpackage

type SecretM struct {
	ID          int64         `gorm:"column:id;primary_key" json:"id"`       //
	Username    string        `gorm:"column:username" json:"username"`       //
	SecretID    string        `gorm:"column:secretID" json:"secretID"`       //
	SecretKey   string        `gorm:"column:secretKey" json:"secretKey"`     //
	Status      sql.NullInt64 `gorm:"column:status" json:"status"`           //
	Description string        `gorm:"column:description" json:"description"` //
	CreatedAt   time.Time     `gorm:"column:createdAt" json:"createdAt"`     //
	UpdatedAt   time.Time     `gorm:"column:updatedAt" json:"updatedAt"`     //
	Expires     int64         `gorm:"column:expires" json:"expires"`         //
}

// TableName sets the insert table name for this struct type
func (s *SecretM) TableName() string {
	return "secret"
}
```
