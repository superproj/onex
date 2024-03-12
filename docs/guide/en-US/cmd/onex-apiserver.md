## onex-apiserver

Launch a onex API server

### Synopsis

The OneX API server validates and configures data
for the api objects which include miners, minersets, configmaps, and
others. The API Server services REST operations and provides the frontend to the
onex's shared state through which all other components interact.

```
onex-apiserver [flags]
```

### Options

```
      --admission-control-config-file string               File with admission control configuration.
      --advertise-address ip                               The IP address on which to advertise the apiserver to members of the cluster. This address must be reachable by the rest of the cluster. If blank, the --bind-address will be used. If --bind-address is unspecified, the host's default interface will be used.
      --allow-metric-labels stringToString                 The map from metric-label to value allow-list of this label. The key's format is <MetricName>,<LabelName>. The value's format is <allowed_value>,<allowed_value>...e.g. metric1,label1='v1,v2,v3', metric1,label2='v1,v2,v3' metric2,label1='v1,v2,v3'. (default [])
      --allow-metric-labels-manifest string                The path to the manifest file that contains the allow-list mapping. The format of the file is the same as the flag --allow-metric-labels. Note that the flag --allow-metric-labels will override the manifest file.
      --audit-log-batch-buffer-size int                    The size of the buffer to store events before batching and writing. Only used in batch mode. (default 10000)
      --audit-log-batch-max-size int                       The maximum size of a batch. Only used in batch mode. (default 1)
      --audit-log-batch-max-wait duration                  The amount of time to wait before force writing the batch that hadn't reached the max size. Only used in batch mode.
      --audit-log-batch-throttle-burst int                 Maximum number of requests sent at the same moment if ThrottleQPS was not utilized before. Only used in batch mode.
      --audit-log-batch-throttle-enable                    Whether batching throttling is enabled. Only used in batch mode.
      --audit-log-batch-throttle-qps float32               Maximum average number of batches per second. Only used in batch mode.
      --audit-log-compress                                 If set, the rotated log files will be compressed using gzip.
      --audit-log-format string                            Format of saved audits. "legacy" indicates 1-line text format for each event. "json" indicates structured json format. Known formats are legacy,json. (default "json")
      --audit-log-maxage int                               The maximum number of days to retain old audit log files based on the timestamp encoded in their filename.
      --audit-log-maxbackup int                            The maximum number of old audit log files to retain. Setting a value of 0 will mean there's no restriction on the number of files.
      --audit-log-maxsize int                              The maximum size in megabytes of the audit log file before it gets rotated.
      --audit-log-mode string                              Strategy for sending audit events. Blocking indicates sending events should block server responses. Batch causes the backend to buffer and write events asynchronously. Known modes are batch,blocking,blocking-strict. (default "blocking")
      --audit-log-path string                              If set, all requests coming to the apiserver will be logged to this file.  '-' means standard out.
      --audit-log-truncate-enabled                         Whether event and batch truncating is enabled.
      --audit-log-truncate-max-batch-size int              Maximum size of the batch sent to the underlying backend. Actual serialized size can be several hundreds of bytes greater. If a batch exceeds this limit, it is split into several batches of smaller size. (default 10485760)
      --audit-log-truncate-max-event-size int              Maximum size of the audit event sent to the underlying backend. If the size of an event is greater than this number, first request and response are removed, and if this doesn't reduce the size enough, event is discarded. (default 102400)
      --audit-log-version string                           API group and version used for serializing audit events written to log. (default "audit.k8s.io/v1")
      --audit-policy-file string                           Path to the file that defines the audit policy configuration.
      --audit-webhook-batch-buffer-size int                The size of the buffer to store events before batching and writing. Only used in batch mode. (default 10000)
      --audit-webhook-batch-max-size int                   The maximum size of a batch. Only used in batch mode. (default 400)
      --audit-webhook-batch-max-wait duration              The amount of time to wait before force writing the batch that hadn't reached the max size. Only used in batch mode. (default 30s)
      --audit-webhook-batch-throttle-burst int             Maximum number of requests sent at the same moment if ThrottleQPS was not utilized before. Only used in batch mode. (default 15)
      --audit-webhook-batch-throttle-enable                Whether batching throttling is enabled. Only used in batch mode. (default true)
      --audit-webhook-batch-throttle-qps float32           Maximum average number of batches per second. Only used in batch mode. (default 10)
      --audit-webhook-config-file string                   Path to a kubeconfig formatted file that defines the audit webhook configuration.
      --audit-webhook-initial-backoff duration             The amount of time to wait before retrying the first failed request. (default 10s)
      --audit-webhook-mode string                          Strategy for sending audit events. Blocking indicates sending events should block server responses. Batch causes the backend to buffer and write events asynchronously. Known modes are batch,blocking,blocking-strict. (default "batch")
      --audit-webhook-truncate-enabled                     Whether event and batch truncating is enabled.
      --audit-webhook-truncate-max-batch-size int          Maximum size of the batch sent to the underlying backend. Actual serialized size can be several hundreds of bytes greater. If a batch exceeds this limit, it is split into several batches of smaller size. (default 10485760)
      --audit-webhook-truncate-max-event-size int          Maximum size of the audit event sent to the underlying backend. If the size of an event is greater than this number, first request and response are removed, and if this doesn't reduce the size enough, event is discarded. (default 102400)
      --audit-webhook-version string                       API group and version used for serializing audit events written to webhook. (default "audit.k8s.io/v1")
      --authentication-kubeconfig string                   kubeconfig file pointing at the 'core' kubernetes server with enough rights to create tokenreviews.authentication.k8s.io.
      --authentication-skip-lookup                         If false, the authentication-kubeconfig will be used to lookup missing authentication configuration from the cluster.
      --authentication-token-webhook-cache-ttl duration    The duration to cache responses from the webhook token authenticator. (default 10s)
      --authentication-tolerate-lookup-failure             If true, failures to look up missing authentication configuration from the cluster are not considered fatal. Note that this can result in authentication that treats all requests as anonymous.
      --bind-address ip                                    The IP address on which to listen for the --secure-port port. The associated interface(s) must be reachable by the rest of the cluster, and by CLI/web clients. If blank or an unspecified address (0.0.0.0 or ::), all interfaces and IP address families will be used. (default 0.0.0.0)
      --cert-dir string                                    The directory where the TLS certs are located. If --tls-cert-file and --tls-private-key-file are provided, this flag will be ignored. (default "_output/certificates")
      --client-ca-file string                              If set, any request presenting a client certificate signed by one of the authorities in the client-ca-file is authenticated with an identity corresponding to the CommonName of the client certificate.
  -c, --config FILE                                        Read configuration from specified FILE, support JSON, TOML, YAML, HCL, or Java properties formats.
      --contention-profiling                               Enable block profiling, if profiling is enabled
      --cors-allowed-origins strings                       List of allowed origins for CORS, comma separated. An allowed origin can be a regular expression to support subdomain matching. If this list is empty CORS will not be enabled. Please ensure each expression matches the entire hostname by anchoring to the start with '^' or including the '//' prefix, and by anchoring to the end with '$' or including the ':' port separator suffix. Examples of valid expressions are '//example\.com(:|$)' and '^https://example\.com(:|$)'
      --debug-socket-path string                           Use an unprotected (no authn/authz) unix-domain socket for profiling with the given path
      --delete-collection-workers int                      Number of workers spawned for DeleteCollection call. These are used to speed up namespace cleanup. (default 1)
      --disable-admission-plugins strings                  admission plugins that should be disabled although they are in the default enabled plugins list (NamespaceAutoProvision, NamespaceLifecycle). Comma-delimited list of admission plugins: AlwaysAdmit, AlwaysDeny, NamespaceAutoProvision, NamespaceExists, NamespaceLifecycle. The order of plugins in this flag does not matter.
      --disabled-metrics strings                           This flag provides an escape hatch for misbehaving metrics. You must provide the fully qualified metric name in order to disable it. Disclaimer: disabling metrics is higher in precedence than showing hidden metrics.
      --egress-selector-config-file string                 File with apiserver egress selector configuration.
      --enable-admission-plugins strings                   admission plugins that should be enabled in addition to default enabled ones (NamespaceAutoProvision, NamespaceLifecycle). Comma-delimited list of admission plugins: AlwaysAdmit, AlwaysDeny, NamespaceAutoProvision, NamespaceExists, NamespaceLifecycle. The order of plugins in this flag does not matter.
      --enable-garbage-collector                           Enables the generic garbage collector. MUST be synced with the corresponding flag of the kube-controller-manager. (default true)
      --enable-priority-and-fairness                       If true, replace the max-in-flight handler with an enhanced one that queues and dispatches with priority and fairness (default true)
      --encryption-provider-config string                  The file containing configuration for encryption providers to be used for storing secrets in etcd
      --encryption-provider-config-automatic-reload        Determines if the file set by --encryption-provider-config should be automatically reloaded if the disk contents change. Setting this to true disables the ability to uniquely identify distinct KMS plugins via the API server healthz endpoints.
      --etcd-cafile string                                 SSL Certificate Authority file used to secure etcd communication.
      --etcd-certfile string                               SSL certification file used to secure etcd communication.
      --etcd-compaction-interval duration                  The interval of compaction requests. If 0, the compaction request from apiserver is disabled. (default 5m0s)
      --etcd-count-metric-poll-period duration             Frequency of polling etcd for number of resources per type. 0 disables the metric collection. (default 1m0s)
      --etcd-db-metric-poll-interval duration              The interval of requests to poll etcd and update metric. 0 disables the metric collection (default 30s)
      --etcd-healthcheck-timeout duration                  The timeout to use when checking etcd health. (default 2s)
      --etcd-keyfile string                                SSL key file used to secure etcd communication.
      --etcd-prefix string                                 The prefix to prepend to all resource paths in etcd. (default "/registry/onex.io")
      --etcd-readycheck-timeout duration                   The timeout to use when checking etcd readiness (default 2s)
      --etcd-servers strings                               List of etcd servers to connect with (scheme://ip:port), comma separated.
      --etcd-servers-overrides strings                     Per-resource etcd servers overrides, comma separated. The individual override format: group/resource#servers, where servers are URLs, semicolon separated. Note that this applies only to resources compiled into this server binary. 
      --event-ttl duration                                 Amount of time to retain events. (default 1h0m0s)
      --external-hostname string                           The hostname to use when generating externalized URLs for this master (e.g. Swagger API Docs or OpenID Discovery).
      --feature-gates mapStringBool                        A set of key=value pairs that describe feature gates for alpha/experimental features. Options are:
                                                           APIResponseCompression=true|false (BETA - default=true)
                                                           APIServerIdentity=true|false (BETA - default=true)
                                                           APIServerTracing=true|false (BETA - default=true)
                                                           AdmissionWebhookMatchConditions=true|false (BETA - default=true)
                                                           AggregatedDiscoveryEndpoint=true|false (BETA - default=true)
                                                           AllAlpha=true|false (ALPHA - default=false)
                                                           AllBeta=true|false (BETA - default=false)
                                                           AnyVolumeDataSource=true|false (BETA - default=true)
                                                           AppArmor=true|false (BETA - default=true)
                                                           CPUManagerPolicyAlphaOptions=true|false (ALPHA - default=false)
                                                           CPUManagerPolicyBetaOptions=true|false (BETA - default=true)
                                                           CPUManagerPolicyOptions=true|false (BETA - default=true)
                                                           CRDValidationRatcheting=true|false (ALPHA - default=false)
                                                           CSIMigrationPortworx=true|false (BETA - default=false)
                                                           CSIVolumeHealth=true|false (ALPHA - default=false)
                                                           CloudControllerManagerWebhook=true|false (ALPHA - default=false)
                                                           CloudDualStackNodeIPs=true|false (BETA - default=true)
                                                           ClusterTrustBundle=true|false (ALPHA - default=false)
                                                           ClusterTrustBundleProjection=true|false (ALPHA - default=false)
                                                           ComponentSLIs=true|false (BETA - default=true)
                                                           ConsistentListFromCache=true|false (ALPHA - default=false)
                                                           ContainerCheckpoint=true|false (ALPHA - default=false)
                                                           ContextualLogging=true|false (ALPHA - default=false)
                                                           CronJobsScheduledAnnotation=true|false (BETA - default=true)
                                                           CrossNamespaceVolumeDataSource=true|false (ALPHA - default=false)
                                                           CustomCPUCFSQuotaPeriod=true|false (ALPHA - default=false)
                                                           DevicePluginCDIDevices=true|false (BETA - default=true)
                                                           DisableCloudProviders=true|false (BETA - default=true)
                                                           DisableKubeletCloudCredentialProviders=true|false (BETA - default=true)
                                                           DisableNodeKubeProxyVersion=true|false (ALPHA - default=false)
                                                           DynamicResourceAllocation=true|false (ALPHA - default=false)
                                                           ElasticIndexedJob=true|false (BETA - default=true)
                                                           EventedPLEG=true|false (BETA - default=false)
                                                           GracefulNodeShutdown=true|false (BETA - default=true)
                                                           GracefulNodeShutdownBasedOnPodPriority=true|false (BETA - default=true)
                                                           HPAContainerMetrics=true|false (BETA - default=true)
                                                           HPAScaleToZero=true|false (ALPHA - default=false)
                                                           HonorPVReclaimPolicy=true|false (ALPHA - default=false)
                                                           ImageMaximumGCAge=true|false (ALPHA - default=false)
                                                           InPlacePodVerticalScaling=true|false (ALPHA - default=false)
                                                           InTreePluginAWSUnregister=true|false (ALPHA - default=false)
                                                           InTreePluginAzureDiskUnregister=true|false (ALPHA - default=false)
                                                           InTreePluginAzureFileUnregister=true|false (ALPHA - default=false)
                                                           InTreePluginGCEUnregister=true|false (ALPHA - default=false)
                                                           InTreePluginOpenStackUnregister=true|false (ALPHA - default=false)
                                                           InTreePluginPortworxUnregister=true|false (ALPHA - default=false)
                                                           InTreePluginvSphereUnregister=true|false (ALPHA - default=false)
                                                           JobBackoffLimitPerIndex=true|false (BETA - default=true)
                                                           JobPodFailurePolicy=true|false (BETA - default=true)
                                                           JobPodReplacementPolicy=true|false (BETA - default=true)
                                                           KubeProxyDrainingTerminatingNodes=true|false (ALPHA - default=false)
                                                           KubeletCgroupDriverFromCRI=true|false (ALPHA - default=false)
                                                           KubeletInUserNamespace=true|false (ALPHA - default=false)
                                                           KubeletPodResourcesDynamicResources=true|false (ALPHA - default=false)
                                                           KubeletPodResourcesGet=true|false (ALPHA - default=false)
                                                           KubeletSeparateDiskGC=true|false (ALPHA - default=false)
                                                           KubeletTracing=true|false (BETA - default=true)
                                                           LegacyServiceAccountTokenCleanUp=true|false (BETA - default=true)
                                                           LoadBalancerIPMode=true|false (ALPHA - default=false)
                                                           LocalStorageCapacityIsolationFSQuotaMonitoring=true|false (ALPHA - default=false)
                                                           LogarithmicScaleDown=true|false (BETA - default=true)
                                                           LoggingAlphaOptions=true|false (ALPHA - default=false)
                                                           LoggingBetaOptions=true|false (BETA - default=true)
                                                           MatchLabelKeysInPodAffinity=true|false (ALPHA - default=false)
                                                           MatchLabelKeysInPodTopologySpread=true|false (BETA - default=true)
                                                           MaxUnavailableStatefulSet=true|false (ALPHA - default=false)
                                                           MemoryManager=true|false (BETA - default=true)
                                                           MemoryQoS=true|false (ALPHA - default=false)
                                                           MinDomainsInPodTopologySpread=true|false (BETA - default=true)
                                                           MultiCIDRServiceAllocator=true|false (ALPHA - default=false)
                                                           NFTablesProxyMode=true|false (ALPHA - default=false)
                                                           NewVolumeManagerReconstruction=true|false (BETA - default=true)
                                                           NodeInclusionPolicyInPodTopologySpread=true|false (BETA - default=true)
                                                           NodeLogQuery=true|false (ALPHA - default=false)
                                                           NodeSwap=true|false (BETA - default=false)
                                                           OpenAPIEnums=true|false (BETA - default=true)
                                                           PDBUnhealthyPodEvictionPolicy=true|false (BETA - default=true)
                                                           PersistentVolumeLastPhaseTransitionTime=true|false (BETA - default=true)
                                                           PodAndContainerStatsFromCRI=true|false (ALPHA - default=false)
                                                           PodDeletionCost=true|false (BETA - default=true)
                                                           PodDisruptionConditions=true|false (BETA - default=true)
                                                           PodHostIPs=true|false (BETA - default=true)
                                                           PodIndexLabel=true|false (BETA - default=true)
                                                           PodLifecycleSleepAction=true|false (ALPHA - default=false)
                                                           PodReadyToStartContainersCondition=true|false (BETA - default=true)
                                                           PodSchedulingReadiness=true|false (BETA - default=true)
                                                           ProcMountType=true|false (ALPHA - default=false)
                                                           QOSReserved=true|false (ALPHA - default=false)
                                                           RecoverVolumeExpansionFailure=true|false (ALPHA - default=false)
                                                           RotateKubeletServerCertificate=true|false (BETA - default=true)
                                                           RuntimeClassInImageCriApi=true|false (ALPHA - default=false)
                                                           SELinuxMountReadWriteOncePod=true|false (BETA - default=true)
                                                           SchedulerQueueingHints=true|false (BETA - default=false)
                                                           SecurityContextDeny=true|false (ALPHA - default=false)
                                                           SeparateTaintEvictionController=true|false (BETA - default=true)
                                                           ServiceAccountTokenJTI=true|false (ALPHA - default=false)
                                                           ServiceAccountTokenNodeBinding=true|false (ALPHA - default=false)
                                                           ServiceAccountTokenNodeBindingValidation=true|false (ALPHA - default=false)
                                                           ServiceAccountTokenPodNodeInfo=true|false (ALPHA - default=false)
                                                           SidecarContainers=true|false (BETA - default=true)
                                                           SizeMemoryBackedVolumes=true|false (BETA - default=true)
                                                           StableLoadBalancerNodeSet=true|false (BETA - default=true)
                                                           StatefulSetAutoDeletePVC=true|false (BETA - default=true)
                                                           StatefulSetStartOrdinal=true|false (BETA - default=true)
                                                           StorageVersionAPI=true|false (ALPHA - default=false)
                                                           StorageVersionHash=true|false (BETA - default=true)
                                                           StructuredAuthenticationConfiguration=true|false (ALPHA - default=false)
                                                           StructuredAuthorizationConfiguration=true|false (ALPHA - default=false)
                                                           TopologyAwareHints=true|false (BETA - default=true)
                                                           TopologyManagerPolicyAlphaOptions=true|false (ALPHA - default=false)
                                                           TopologyManagerPolicyBetaOptions=true|false (BETA - default=true)
                                                           TopologyManagerPolicyOptions=true|false (BETA - default=true)
                                                           TranslateStreamCloseWebsocketRequests=true|false (ALPHA - default=false)
                                                           UnauthenticatedHTTP2DOSMitigation=true|false (BETA - default=true)
                                                           UnknownVersionInteroperabilityProxy=true|false (ALPHA - default=false)
                                                           UserNamespacesPodSecurityStandards=true|false (ALPHA - default=false)
                                                           UserNamespacesSupport=true|false (ALPHA - default=false)
                                                           ValidatingAdmissionPolicy=true|false (BETA - default=false)
                                                           VolumeAttributesClass=true|false (ALPHA - default=false)
                                                           VolumeCapacityPriority=true|false (ALPHA - default=false)
                                                           WatchList=true|false (ALPHA - default=false)
                                                           WinDSR=true|false (ALPHA - default=false)
                                                           WinOverlay=true|false (BETA - default=true)
                                                           WindowsHostNetwork=true|false (ALPHA - default=true)
                                                           ZeroLimitedNominalConcurrencyShares=true|false (BETA - default=false)
      --goaway-chance float                                To prevent HTTP/2 clients from getting stuck on a single apiserver, randomly close a connection (GOAWAY). The client's other in-flight requests won't be affected, and the client will reconnect, likely landing on a different apiserver after going through the load balancer again. This argument sets the fraction of requests that will be sent a GOAWAY. Clusters with single apiservers, or which don't use a load balancer, should NOT enable this. Min is 0 (off), Max is .02 (1/50 requests); .001 (1/1000) is a recommended starting point.
  -h, --help                                               help for onex-apiserver
      --http2-max-streams-per-connection int               The limit that the server gives to clients for the maximum number of streams in an HTTP/2 connection. Zero means to use golang's default. (default 1000)
      --lease-reuse-duration-seconds int                   The time in seconds that each lease is reused. A lower value could avoid large number of objects reusing the same lease. Notice that a too small value may cause performance problems at storage layer. (default 60)
      --livez-grace-period duration                        This option represents the maximum amount of time it should take for apiserver to complete its startup sequence and become live. From apiserver's start time to when this amount of time has elapsed, /livez will assume that unfinished post-start hooks will complete successfully and therefore return true.
      --log-flush-frequency duration                       Maximum number of seconds between log flushes (default 5s)
      --logging-format string                              Sets the log format. Permitted formats: "text". (default "text")
      --max-mutating-requests-inflight int                 This and --max-requests-inflight are summed to determine the server's total concurrency limit (which must be positive) if --enable-priority-and-fairness is true. Otherwise, this flag limits the maximum number of mutating requests in flight, or a zero value disables the limit completely. (default 200)
      --max-requests-inflight int                          This and --max-mutating-requests-inflight are summed to determine the server's total concurrency limit (which must be positive) if --enable-priority-and-fairness is true. Otherwise, this flag limits the maximum number of non-mutating requests in flight, or a zero value disables the limit completely. (default 400)
      --min-request-timeout int                            An optional field indicating the minimum number of seconds a handler must keep a request open before timing it out. Currently only honored by the watch request handler, which picks a randomized value above this number as the connection timeout, to spread out load. (default 1800)
      --permit-address-sharing                             If true, SO_REUSEADDR will be used when binding the port. This allows binding to wildcard IPs like 0.0.0.0 and specific IPs in parallel, and it avoids waiting for the kernel to release sockets in TIME_WAIT state. [default=false]
      --permit-port-sharing                                If true, SO_REUSEPORT will be used when binding the port, which allows more than one instance to bind on the same address and port. [default=false]
      --profiling                                          Enable profiling via web interface host:port/debug/pprof/ (default true)
      --request-timeout duration                           An optional field indicating the duration a handler must keep a request open before timing it out. This is the default request timeout for requests but may be overridden by flags such as --min-request-timeout for specific types of requests. (default 1m0s)
      --requestheader-allowed-names strings                List of client certificate common names to allow to provide usernames in headers specified by --requestheader-username-headers. If empty, any client certificate validated by the authorities in --requestheader-client-ca-file is allowed.
      --requestheader-client-ca-file string                Root certificate bundle to use to verify client certificates on incoming requests before trusting usernames in headers specified by --requestheader-username-headers. WARNING: generally do not depend on authorization being already done for incoming requests.
      --requestheader-extra-headers-prefix strings         List of request header prefixes to inspect. X-Remote-Extra- is suggested. (default [x-remote-extra-])
      --requestheader-group-headers strings                List of request headers to inspect for groups. X-Remote-Group is suggested. (default [x-remote-group])
      --requestheader-username-headers strings             List of request headers to inspect for usernames. X-Remote-User is common. (default [x-remote-user])
      --secure-port int                                    The port on which to serve HTTPS with authentication and authorization. If 0, don't serve HTTPS at all. (default 443)
      --show-hidden-metrics-for-version string             The previous version for which you want to show hidden metrics. Only the previous minor version is meaningful, other values will not be allowed. The format is <major>.<minor>, e.g.: '1.16'. The purpose of this format is make sure you have the opportunity to notice if the next release hides additional metrics, rather than being surprised when they are permanently removed in the release after that.
      --shutdown-delay-duration duration                   Time to delay the termination. During that time the server keeps serving requests normally. The endpoints /healthz and /livez will return success, but /readyz immediately returns failure. Graceful termination starts after this delay has elapsed. This can be used to allow load balancer to stop sending traffic to this server.
      --shutdown-send-retry-after                          If true the HTTP Server will continue listening until all non long running request(s) in flight have been drained, during this window all incoming requests will be rejected with a status code 429 and a 'Retry-After' response header, in addition 'Connection: close' response header is set in order to tear down the TCP connection when idle.
      --shutdown-watch-termination-grace-period duration   This option, if set, represents the maximum amount of grace period the apiserver will wait for active watch request(s) to drain during the graceful server shutdown window.
      --storage-backend string                             The storage backend for persistence. Options: 'etcd3' (default).
      --storage-media-type string                          The media type to use to store objects in storage. Some resources or storage backends may only support a specific media type and will ignore this setting. Supported media types: [application/json, application/yaml, application/vnd.kubernetes.protobuf] (default "application/json")
      --strict-transport-security-directives strings       List of directives for HSTS, comma separated. If this list is empty, then HSTS directives will not be added. Example: 'max-age=31536000,includeSubDomains,preload'
      --tls-cert-file string                               File containing the default x509 Certificate for HTTPS. (CA cert, if any, concatenated after server cert). If HTTPS serving is enabled, and --tls-cert-file and --tls-private-key-file are not provided, a self-signed certificate and key are generated for the public address and saved to the directory specified by --cert-dir.
      --tls-cipher-suites strings                          Comma-separated list of cipher suites for the server. If omitted, the default Go cipher suites will be used. 
                                                           Preferred values: TLS_AES_128_GCM_SHA256, TLS_AES_256_GCM_SHA384, TLS_CHACHA20_POLY1305_SHA256, TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA, TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256, TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA, TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384, TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305, TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256, TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA, TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256, TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA, TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384, TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305, TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256, TLS_RSA_WITH_AES_128_CBC_SHA, TLS_RSA_WITH_AES_128_GCM_SHA256, TLS_RSA_WITH_AES_256_CBC_SHA, TLS_RSA_WITH_AES_256_GCM_SHA384. 
                                                           Insecure values: TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256, TLS_ECDHE_ECDSA_WITH_RC4_128_SHA, TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA, TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256, TLS_ECDHE_RSA_WITH_RC4_128_SHA, TLS_RSA_WITH_3DES_EDE_CBC_SHA, TLS_RSA_WITH_AES_128_CBC_SHA256, TLS_RSA_WITH_RC4_128_SHA.
      --tls-min-version string                             Minimum TLS version supported. Possible values: VersionTLS10, VersionTLS11, VersionTLS12, VersionTLS13
      --tls-private-key-file string                        File containing the default x509 private key matching --tls-cert-file.
      --tls-sni-cert-key namedCertKey                      A pair of x509 certificate and private key file paths, optionally suffixed with a list of domain patterns which are fully qualified domain names, possibly with prefixed wildcard segments. The domain patterns also allow IP addresses, but IPs should only be used if the apiserver has visibility to the IP address requested by a client. If no domain patterns are provided, the names of the certificate are extracted. Non-wildcard matches trump over wildcard matches, explicit domain patterns trump over extracted names. For multiple key/certificate pairs, use the --tls-sni-cert-key multiple times. Examples: "example.crt,example.key" or "foo.crt,foo.key:*.foo.com,foo.com". (default [])
      --tracing-config-file string                         File with apiserver tracing configuration.
  -v, --v Level                                            number for the log level verbosity
      --version version[=true]                             Print version information and quit
      --vmodule pattern=N,...                              comma-separated list of pattern=N settings for file-filtered logging (only works for text log format)
      --watch-cache                                        Enable watch caching in the apiserver (default true)
      --watch-cache-sizes strings                          Watch cache size settings for some resources (pods, nodes, etc.), comma separated. The individual setting format: resource[.group]#size, where resource is lowercase plural (no version), group is omitted for resources of apiVersion v1 (the legacy core API) and included for others, and size is a number. This option is only meaningful for resources built into the apiserver, not ones defined by CRDs or aggregated from external servers, and is only consulted if the watch-cache is enabled. The only meaningful size setting to supply here is zero, which means to disable watch caching for the associated resource; all non-zero values are equivalent and mean to not disable watch caching for that resource
```

###### Auto generated by spf13/cobra on 20-Jan-2024
