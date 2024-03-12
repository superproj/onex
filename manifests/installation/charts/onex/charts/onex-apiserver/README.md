# onex-apiserver

![Version: 0.32.0](https://img.shields.io/badge/Version-0.32.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.0.1](https://img.shields.io/badge/AppVersion-0.0.1-informational?style=flat-square)

A Helm chart for onex-apiserver

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| args.clientCAFile | string | `"/opt/onex/cert/ca.pem"` |  |
| args.disableAdmissionPlugins | string | `"Whitelist"` |  |
| args.enableAdmissionPlugins | string | `"RollingUpgrade,ResourceQuota,Cluster,Native,NamespaceAuthorize"` |  |
| args.etcdCAFile | string | `"/root/etcdcert/ca.pem"` |  |
| args.etcdCertFile | string | `"/root/etcdcert/client.pem"` |  |
| args.etcdKeyFile | string | `"/root/etcdcert/client-key.pem"` |  |
| args.etcdServers | string | `"https://127.0.0.1:2379"` |  |
| args.maxMutatingRequestsInflight | int | `2000` |  |
| args.maxRequestsInflight | int | `5000` |  |
| args.securePort | int | `8443` |  |
| args.tlsCertFile | string | `"/opt/onex/cert/onex-apiserver.pem"` |  |
| args.tlsPrivateKeyFile | string | `"/opt/onex/cert/onex-apiserver-key.pem"` |  |
| etcdCerts.ca | string | `"xxx"` |  |
| etcdCerts.cert | string | `"xxx"` |  |
| etcdCerts.key | string | `"xxx"` |  |
| image | string | `"ccr.ccs.tencentyun.com/superproj/onex-apiserver-amd64:v1.0.0"` |  |
| imagePullPolicy | string | `"Always"` |  |
| replicas | int | `1` |  |
| securityContext.capabilities.drop[0] | string | `"ALL"` |  |
| securityContext.readOnlyRootFilesystem | bool | `true` |  |
| securityContext.runAsGroup | int | `10000` |  |
| securityContext.runAsNonRoot | bool | `true` |  |
| securityContext.runAsUser | int | `10000` |  |
| testEnv | bool | `false` |  |

