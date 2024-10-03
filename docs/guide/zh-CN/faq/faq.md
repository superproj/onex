## 执行 `./scripts/update-codegen.sh protobuf` 时报错

报以下错误：
```
2024/10/01 13:55:15 Unable to clean package k8s.io.apimachinery.pkg.api.resource: remove /data00/.cache/golang/pkg/mod/k8s.io/apimachinery@v0.30.2/pkg/api/resource/generated.proto: permission denied
...
/data00/.cache/golang/pkg/mod/k8s.io/apimachinery@v0.30.2/pkg/apis/meta/v1beta1/generated.pb.go
```

原因：
1. 执行 scripts/update-codegen.sh protobuf 会先删除再生成
2. 


解决方法：

```bash
--apimachinery-packages '-k8s.io/apimachinery/pkg/util/intstr,-k8s.io/apimachinery/pkg/api/resource,-k8s.io/apimachinery/pkg/runtime/schema,-k8s.io/apimachinery/pkg/runtime,-k8s.io/apimachinery/pkg/apis/meta/v1,-k8s.io/apimachinery/pkg/apis/meta/v1beta1,-k8s.io/apimachinery/pkg/apis/testapigroup/v1'
```
