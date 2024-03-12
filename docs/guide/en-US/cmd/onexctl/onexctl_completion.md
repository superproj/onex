## onexctl completion

Output shell completion code for the specified shell (bash, zsh, fish, or powershell)

### Synopsis

Output shell completion code for the specified shell (bash or zsh). The shell code must be evaluated to provide interactive completion of onexctl commands.  This can be done by sourcing it from the .bash_profile.

 Detailed instructions on how to do this are available here: http://github.com/superproj/onex/docs/installation/onexctl.md#enabling-shell-autocompletion

 Note for zsh users: [1] zsh completions are only supported in versions of zsh >= 5.2

```
onexctl completion SHELL
```

### Examples

```
  # Installing bash completion on macOS using homebrew
  ## If running Bash 3.2 included with macOS
  brew install bash-completion
  ## or, if running Bash 4.1+
  brew install bash-completion@2
  ## If onexctl is installed via homebrew, this should start working immediately.
  ## If you've installed via other means, you may need add the completion to your completion directory
  onexctl completion bash > $(brew --prefix)/etc/bash_completion.d/onexctl
  
  
  # Installing bash completion on Linux
  ## If bash-completion is not installed on Linux, please install the 'bash-completion' package
  ## via your distribution's package manager.
  ## Load the onexctl completion code for bash into the current shell
  source <(onexctl completion bash)
  ## Write bash completion code to a file and source if from .bash_profile
  onexctl completion bash > ~/.onex/onexctl.completion.bash.inc
  printf "
  # OneX shell completion
  source '$HOME/.onex/onexctl.completion.bash.inc'
  " >> $HOME/.bash_profile
  source $HOME/.bash_profile
  
  # Load the onexctl completion code for zsh[1] into the current shell
  source <(onexctl completion zsh)
  # Set the onexctl completion code for zsh[1] to autoload on startup
  onexctl completion zsh > "${fpath[1]}/_onexctl"
  
  # Load the onexctl completion code for fish[2] into the current shell
  onexctl completion fish | source
  # To load completions for each session, execute once:
  onexctl completion fish > ~/.config/fish/completions/onexctl.fish
  
  # Load the onexctl completion code for powershell into the current shell
  onexctl completion powershell | Out-String | Invoke-Expression
  # Set onexctl completion code for powershell to run on startup
  ## Save completion code to a script and execute in the profile
  onexctl completion powershell > $HOME\.onex\completion.ps1
  Add-Content $PROFILE "$HOME\.onex\completion.ps1"
  ## Execute completion code in the profile
  Add-Content $PROFILE "if (Get-Command onexctl -ErrorAction SilentlyContinue) {
  onexctl completion powershell | Out-String | Invoke-Expression
  }"
  ## Add completion code directly to the $PROFILE script
  onexctl completion powershell >> $PROFILE
```

### Options

```
  -h, --help   help for completion
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

* [onexctl](onexctl.md)	 - onexctl controls the onex cloud platform

###### Auto generated by spf13/cobra on 20-Jan-2024
