## onexctl jwt sign

Sign a jwt token with given secretID and secretKey

### Synopsis

Sign a jwt token with given secretID and secretKey

```
onexctl jwt sign SECRETID SECRETKEY
```

### Examples

```
  # Sign a token with secretID and secretKey
  onexctl sign tgydj8d9EQSnFqKf iBdEdFNBLN1nR3fV
  
  # Sign a token with expires and sign method
  onexctl sign tgydj8d9EQSnFqKf iBdEdFNBLN1nR3fV --timeout=2h --algorithm=HS256
```

### Options

```
      --algorithm string      Signing algorithm - possible values are HS256, HS384, HS512. (default "HS256")
      --header map            Add additional header params. may be used more than once. (default {})
  -h, --help                  help for sign
      --issuer string         Identifies the principal that issued the JWT. (default "onexctl")
      --not-before duration   Identifies the time before which the JWT MUST NOT be accepted for processing.
      --timeout duration      JWT token expires time. (default 2h0m0s)
```

### Options inherited from parent commands

```
      --alsologtostderr                           log to standard error as well as files
      --config string                             Path to the config file to use for CLI.
      --gateway.address string                    The address and port of the OneX API server
      --gateway.certificate-authority string      Path to a cert file for the certificate authority
      --gateway.insecure-skip-tls-verify          If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
      --gateway.max-retries int                   Maximum number of retries.
      --gateway.retry-interval duration           The interval time between each attempt.
      --gateway.timeout duration                  The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests.
      --kubeconfig string                         Paths to a kubeconfig. Only required if out-of-cluster.
      --log-backtrace-at traceLocations           when logging hits line file:N, emit a stack trace
      --log-dir string                            If non-empty, write log files in this directory
      --log-link string                           If non-empty, add symbolic links in this directory to the log files
      --logbuflevel int                           Buffer log messages logged at this level or lower (-1 means don't buffer; 0 means buffer INFO only; ...). Has limited applicability on non-prod platforms.
      --logtostderr                               log to standard error instead of files
      --profile string                            Name of profile to capture. One of (none|cpu|heap|goroutine|threadcreate|block|mutex) (default "none")
      --profile-output string                     Name of the file to write the profile to (default "profile.pprof")
      --stderrthreshold severityFlag              logs at or above this threshold go to stderr (default 2)
      --user.client-certificate string            Path to a client certificate file for TLS
      --user.client-key string                    Path to a client key file for TLS
      --user.password string                      Password for basic authentication to the API server
      --user.secret-id string                     SecretID for JWT authentication to the API server
      --user.secret-key string                    SecretKey for jwt authentication to the API server
      --user.token string                         Bearer token for authentication to the API server
      --user.username string                      Username for basic authentication to the API server
      --usercenter.address string                 The address and port of the OneX API server
      --usercenter.certificate-authority string   Path to a cert file for the certificate authority
      --usercenter.insecure-skip-tls-verify       If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
      --usercenter.max-retries int                Maximum number of retries.
      --usercenter.retry-interval duration        The interval time between each attempt.
      --usercenter.timeout duration               The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests.
  -v, --v Level                                   log level for V logs
      --version version[=true]                    Print version information and quit
      --vmodule vModuleFlag                       comma-separated list of pattern=N settings for file-filtered logging
      --warnings-as-errors                        Treat warnings received from the server as errors and exit with a non-zero exit code
```

### SEE ALSO

* [onexctl jwt](onexctl_jwt.md)	 - JWT command-line tool

###### Auto generated by spf13/cobra on 20-Jan-2024
