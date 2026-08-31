module github.com/onexstack/onex

go 1.26.0

require (
	github.com/AlecAivazis/survey/v2 v2.3.7
	github.com/MakeNowJust/heredoc v1.0.0
	github.com/MakeNowJust/heredoc/v2 v2.0.1
	github.com/armon/go-socks5 v0.0.0-20160902184237-e75332964ef5
	github.com/blang/semver/v4 v4.0.0
	github.com/caarlos0/env/v8 v8.0.0
	github.com/casbin/casbin/v2 v2.103.0
	github.com/casbin/gorm-adapter/v3 v3.32.0
	github.com/casbin/redis-watcher/v2 v2.5.0
	github.com/cloudwego/kitex v0.12.1
	github.com/cpuguy83/go-md2man/v2 v2.0.6
	github.com/dgraph-io/ristretto v0.2.0
	github.com/distribution/reference v0.6.0
	github.com/duke-git/lancet/v2 v2.3.4
	github.com/envoyproxy/protoc-gen-validate v1.3.3
	github.com/fatih/color v1.19.0
	github.com/gammazero/workerpool v1.1.3
	github.com/ghodss/yaml v1.0.0
	github.com/gin-contrib/pprof v1.5.2
	github.com/gin-gonic/gin v1.10.0
	github.com/go-kratos/kratos/contrib/registry/consul/v2 v2.0.0-20250127095200-20970020a5ef
	github.com/go-kratos/kratos/contrib/registry/etcd/v2 v2.0.0-20250127095200-20970020a5ef
	github.com/go-kratos/kratos/v2 v2.8.3
	github.com/go-kratos/swagger-api v1.0.1
	github.com/go-logr/logr v1.4.4
	github.com/go-nunu/nunu v1.1.0
	github.com/go-resty/resty/v2 v2.16.5
	github.com/golang-jwt/jwt/v4 v4.5.2
	github.com/golang/mock v1.6.0
	github.com/google/wire v0.6.0
	github.com/gorilla/handlers v1.5.2
	github.com/gorilla/websocket v1.5.4-0.20250319132907-e064f32e3674
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/hashicorp/consul/api v1.31.0
	github.com/jinzhu/copier v0.4.0
	github.com/likexian/host-stat-go v0.0.0-20190516151207-c9cf36dd6ce9
	github.com/looplab/fsm v1.0.2
	github.com/minio/minio-go/v7 v7.0.84
	github.com/olekukonko/tablewriter v1.1.4
	github.com/onexstack/onexstack v0.0.0-00010101000000-000000000000
	github.com/onsi/ginkgo/v2 v2.32.0
	github.com/onsi/gomega v1.42.1
	github.com/prometheus/client_golang v1.24.0
	github.com/redis/go-redis/v9 v9.7.0
	github.com/robfig/cron/v3 v3.0.1
	github.com/russross/blackfriday v1.6.0
	github.com/segmentio/kafka-go v0.4.47
	github.com/spf13/cobra v1.10.2
	github.com/spf13/pflag v1.0.10
	github.com/spf13/viper v1.21.0
	github.com/tmc/langchaingo v0.1.12
	go.etcd.io/etcd/client/pkg/v3 v3.7.0
	go.etcd.io/etcd/client/v3 v3.7.0
	go.mongodb.org/mongo-driver v1.17.2
	go.opencensus.io v0.24.0
	go.opentelemetry.io/otel v1.44.0
	go.opentelemetry.io/otel/sdk v1.44.0
	go.opentelemetry.io/otel/trace v1.44.0
	go.uber.org/ratelimit v0.3.1
	google.golang.org/genproto/googleapis/api v0.0.0-20260526163538-3dc84a4a5aaa
	google.golang.org/grpc v1.82.1
	google.golang.org/protobuf v1.36.12-0.20260120151049-f2248ac996af
	gopkg.in/evanphx/json-patch.v4 v4.13.0
	gorm.io/gen v0.3.26
	gorm.io/gorm v1.30.1
	k8s.io/api v0.37.0
	k8s.io/apiextensions-apiserver v0.37.0
	k8s.io/apimachinery v0.37.0
	k8s.io/apiserver v0.37.0
	k8s.io/cli-runtime v0.37.0
	k8s.io/client-go v0.37.0
	k8s.io/code-generator v0.37.0
	k8s.io/component-base v0.37.0
	k8s.io/controller-manager v0.37.0
	k8s.io/klog/v2 v2.140.0
	k8s.io/kube-aggregator v0.0.0
	k8s.io/kube-openapi v0.0.0-20260721132016-d427ff9ee9ad
	k8s.io/kubectl v0.0.0
	k8s.io/kubernetes v0.0.0-00010101000000-000000000000
	k8s.io/utils v0.0.0-20260626114624-be93311217bd
	sigs.k8s.io/cluster-api v1.14.0
	sigs.k8s.io/cluster-api/api v1.14.0
	sigs.k8s.io/controller-runtime v0.24.1
	sigs.k8s.io/randfill v1.0.0
	sigs.k8s.io/structured-merge-diff/v6 v6.4.2
)

require (
	github.com/lufia/plan9stats v0.0.0-20230326075908-cb1d2100619a // indirect
	github.com/power-devops/perfstat v0.0.0-20221212215047-62379fc7944b // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/yusufpapurcu/wmi v1.2.3 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.68.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.69.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.44.0
	go.opentelemetry.io/otel/metric v1.44.0 // indirect
	go.opentelemetry.io/proto/otlp v1.10.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260526163538-3dc84a4a5aaa // indirect
)

require (
	cel.dev/expr v0.25.1 // indirect
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/BurntSushi/toml v1.4.0 // indirect
	github.com/Masterminds/semver/v3 v3.4.0 // indirect
	github.com/antlr4-go/antlr/v4 v4.13.1 // indirect
	github.com/armon/go-metrics v0.4.1 // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/benbjohnson/clock v1.3.0 // indirect
	github.com/bmatcuk/doublestar/v4 v4.6.1 // indirect
	github.com/bradfitz/gomemcache v0.0.0-20230905024940-24af94b03874 // indirect
	github.com/bytedance/gopkg v0.1.1 // indirect
	github.com/bytedance/sonic v1.12.6 // indirect
	github.com/bytedance/sonic/loader v0.2.1 // indirect
	github.com/casbin/govaluate v1.3.0 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/chai2010/gettext-go v1.0.2 // indirect
	github.com/clipperhouse/uax29/v2 v2.6.0 // indirect
	github.com/cloudwego/base64x v0.1.4 // indirect
	github.com/cloudwego/configmanager v0.2.2 // indirect
	github.com/cloudwego/dynamicgo v0.4.7-0.20241220085612-55704ea4ca8f // indirect
	github.com/cloudwego/fastpb v0.0.5 // indirect
	github.com/cloudwego/frugal v0.2.3 // indirect
	github.com/cloudwego/gopkg v0.1.3 // indirect
	github.com/cloudwego/iasm v0.2.0 // indirect
	github.com/cloudwego/localsession v0.1.1 // indirect
	github.com/cloudwego/netpoll v0.6.5 // indirect
	github.com/cloudwego/runtimex v0.1.0 // indirect
	github.com/cloudwego/thriftgo v0.3.18 // indirect
	github.com/dlclark/regexp2 v1.11.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/emicklei/go-restful/v3 v3.13.0 // indirect
	github.com/evanphx/json-patch/v5 v5.9.11 // indirect
	github.com/exponent-io/jsonpath v0.0.0-20210407135951-1de76d718b3f // indirect
	github.com/fatih/camelcase v1.0.0 // indirect
	github.com/fatih/structtag v1.2.0 // indirect
	github.com/fxamacker/cbor/v2 v2.9.1 // indirect
	github.com/gabriel-vasile/mimetype v1.4.7 // indirect
	github.com/gammazero/deque v0.2.0 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/glebarez/go-sqlite v1.20.3 // indirect
	github.com/glebarez/sqlite v1.7.0 // indirect
	github.com/go-errors/errors v1.4.2 // indirect
	github.com/go-ini/ini v1.67.0 // indirect
	github.com/go-kratos/aegis v0.2.0 // indirect
	github.com/go-kratos/grpc-gateway/v2 v2.5.1-0.20210811062259-c92d36e434b1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-openapi/swag/cmdutils v0.27.1 // indirect
	github.com/go-openapi/swag/conv v0.27.1 // indirect
	github.com/go-openapi/swag/fileutils v0.27.1 // indirect
	github.com/go-openapi/swag/jsonutils v0.27.1 // indirect
	github.com/go-openapi/swag/loading v0.27.1 // indirect
	github.com/go-openapi/swag/mangling v0.27.1 // indirect
	github.com/go-openapi/swag/netutils v0.27.1 // indirect
	github.com/go-openapi/swag/pools v0.27.1 // indirect
	github.com/go-openapi/swag/stringutils v0.27.1 // indirect
	github.com/go-openapi/swag/typeutils v0.27.1 // indirect
	github.com/go-openapi/swag/yamlutils v0.27.1 // indirect
	github.com/go-playground/form/v4 v4.2.1 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.23.0 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/go-zookeeper/zk v1.0.4 // indirect
	github.com/goccy/go-json v0.10.4 // indirect
	github.com/golang-sql/civil v0.0.0-20220223132316-b832511892a9 // indirect
	github.com/golang-sql/sqlexp v0.1.0 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/google/cel-go v0.29.2 // indirect
	github.com/google/gnostic-models v0.7.0 // indirect
	github.com/google/pprof v0.0.0-20260402051712-545e8a4df936 // indirect
	github.com/gorilla/mux v1.8.1 // indirect
	github.com/gosuri/uitable v0.0.4 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus v1.1.0 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.3.3 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.29.0 // indirect
	github.com/h2non/filetype v1.1.3 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-hclog v1.5.0 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/serf v0.10.1 // indirect
	github.com/iancoleman/strcase v0.3.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.7.2 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jhump/protoreflect v1.8.2 // indirect
	github.com/jonboulle/clockwork v0.5.0 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/klauspost/compress v1.19.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.9 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de // indirect
	github.com/likexian/gokit v0.25.9 // indirect
	github.com/lithammer/dedent v1.1.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.19 // indirect
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b // indirect
	github.com/microsoft/go-mssqldb v1.6.0 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/moby/spdystream v0.5.1 // indirect
	github.com/modern-go/gls v0.0.0-20220109145502-612d0167dce5 // indirect
	github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00 // indirect
	github.com/montanaflynn/stats v0.7.1 // indirect
	github.com/mxk/go-flowrate v0.0.0-20140419014527-cca7078d478f // indirect
	github.com/nicksnyder/go-i18n/v2 v2.4.1 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pierrec/lz4/v4 v4.1.18 // indirect
	github.com/pkoukk/tiktoken-go v0.1.6 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/rakyll/statik v0.1.7 // indirect
	github.com/redis/go-redis/extra/rediscensus/v9 v9.7.0 // indirect
	github.com/redis/go-redis/extra/rediscmd/v9 v9.7.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230126093431-47fa9a501578 // indirect
	github.com/rs/xid v1.6.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sagikazarmark/locafero v0.11.0 // indirect
	github.com/shirou/gopsutil/v3 v3.23.12 // indirect
	github.com/sony/sonyflake v1.2.0 // indirect
	github.com/sourcegraph/conc v0.3.1-0.20240121214520-5f936abd7ae8 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	github.com/tidwall/gjson v1.18.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.12 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/xlab/treeprint v1.2.0 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.44.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.44.0 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.yaml.in/yaml/v2 v2.4.4 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/arch v0.12.0 // indirect
	golang.org/x/exp v0.0.0-20260410095643-746e56fc9e2f // indirect
	google.golang.org/genproto v0.0.0-20240401170217-c3f982113cda // indirect
	gopkg.in/go-jose/go-jose.v2 v2.6.3 // indirect
	gorm.io/datatypes v1.1.1-0.20230130040222-c43177d3cf8c // indirect
	gorm.io/driver/mysql v1.5.7 // indirect
	gorm.io/driver/postgres v1.5.11 // indirect
	gorm.io/driver/sqlserver v1.5.3 // indirect
	gorm.io/hints v1.1.0 // indirect
	gorm.io/plugin/dbresolver v1.6.2 // indirect
	k8s.io/cloud-provider v0.37.0 // indirect
	k8s.io/component-helpers v0.37.0 // indirect
	k8s.io/endpointslice v0.0.0 // indirect
	k8s.io/gengo/v2 v2.0.0-20260408192533-25e2208e0dc3 // indirect
	k8s.io/kms v0.37.0 // indirect
	k8s.io/kube-controller-manager v0.0.0 // indirect
	k8s.io/kubelet v0.0.0 // indirect
	k8s.io/metrics v0.37.0 // indirect
	k8s.io/streaming v0.37.0 // indirect
	modernc.org/libc v1.22.2 // indirect
	modernc.org/mathutil v1.5.0 // indirect
	modernc.org/memory v1.5.0 // indirect
	modernc.org/sqlite v1.20.3 // indirect
	sigs.k8s.io/kustomize/api v0.21.1 // indirect
	sigs.k8s.io/kustomize/kustomize/v5 v5.8.1 // indirect
	sigs.k8s.io/kustomize/kyaml v0.21.1 // indirect
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20250102033503-faa5f7b0171c // indirect
	github.com/NYTimes/gziphandler v1.1.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/coreos/go-systemd/v22 v22.7.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/evanphx/json-patch v5.7.0+incompatible
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.9.0
	github.com/go-logr/zapr v1.3.0 // indirect
	github.com/go-openapi/jsonpointer v1.0.0 // indirect
	github.com/go-openapi/jsonreference v1.0.0 // indirect
	github.com/go-openapi/swag v0.27.1 // indirect
	github.com/go-sql-driver/mysql v1.8.1 // indirect
	github.com/gobuffalo/flect v1.0.3
	github.com/gogo/protobuf v1.3.2
	github.com/golang/glog v1.2.5 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.4
	github.com/google/go-cmp v0.7.0
	github.com/google/gofuzz v1.2.0
	github.com/google/uuid v1.6.0
	github.com/hashicorp/golang-lru v1.0.2
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/moby/term v0.5.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.70.0
	github.com/prometheus/procfs v0.21.1 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	go.etcd.io/etcd/api/v3 v3.7.0 // indirect
	go.uber.org/automaxprocs v1.6.0
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.28.0 // indirect
	golang.org/x/crypto v0.54.0 // indirect
	golang.org/x/mod v0.37.0 // indirect
	golang.org/x/net v0.57.0
	golang.org/x/oauth2 v0.36.0 // indirect
	golang.org/x/sync v0.22.0
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/term v0.45.0 // indirect
	golang.org/x/text v0.40.0
	golang.org/x/time v0.15.0
	golang.org/x/tools v0.47.0
	gomodules.xyz/jsonpatch/v2 v2.5.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1
	k8s.io/gengo v0.0.0-20250130153323-76c5745d3511
	sigs.k8s.io/apiserver-network-proxy/konnectivity-client v0.36.0 // indirect
	sigs.k8s.io/json v0.0.0-20250730193827-2d320260d730 // indirect
	sigs.k8s.io/yaml v1.6.0
)

// Version conflicts often occur in go.opentelemetry.io. In order to avoid version confusion
// caused by auto-upgrades, here use replace to solid version.
replace (
	go.opentelemetry.io/contrib/instrumentation => go.opentelemetry.io/contrib/instrumentation v0.42.0
	k8s.io/api => k8s.io/api v0.37.0
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.37.0
	// k8s.io/apimachinery => k8s.io/apimachinery v0.37.0
	// k8s.io/apiserver => k8s.io/apiserver v0.37.0
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.37.0
	k8s.io/client-go => k8s.io/client-go v0.37.0
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.37.0
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.37.0
	k8s.io/code-generator => k8s.io/code-generator v0.37.0
	k8s.io/component-base => k8s.io/component-base v0.37.0
	k8s.io/component-helpers => k8s.io/component-helpers v0.37.0
	k8s.io/controller-manager => k8s.io/controller-manager v0.37.0
	k8s.io/cri-api => k8s.io/cri-api v0.37.0
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.37.0
	k8s.io/dynamic-resource-allocation => k8s.io/dynamic-resource-allocation v0.37.0
	k8s.io/endpointslice => k8s.io/endpointslice v0.37.0
	k8s.io/kms => k8s.io/kms v0.37.0
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.37.0
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.37.0
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.37.0
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.37.0
	k8s.io/kubectl => k8s.io/kubectl v0.37.0
	k8s.io/kubelet => k8s.io/kubelet v0.37.0
	// k8s.io/kubernetes 经常会被自动变更为v1.15.0-alpha.0，这里使用replace解决掉
	k8s.io/kubernetes => k8s.io/kubernetes v1.37.0
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.37.0
	k8s.io/metrics => k8s.io/metrics v0.37.0
	k8s.io/mount-utils => k8s.io/mount-utils v0.37.0
	k8s.io/pod-security-admission => k8s.io/pod-security-admission v0.37.0
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.37.0
	sigs.k8s.io/cluster-api => sigs.k8s.io/cluster-api v1.14.0
	sigs.k8s.io/cluster-api/api => sigs.k8s.io/cluster-api/api v1.14.0
)

replace github.com/onexstack/onexstack => ./staging/src/github.com/onexstack/onexstack

// onexctl 仍旧使用 tablewriter v0 的 API（SetHeader/SetBorder/Colors 等），
// v1 是完全重写、不兼容，故固定回 v0。
replace github.com/olekukonko/tablewriter => github.com/olekukonko/tablewriter v0.0.5
