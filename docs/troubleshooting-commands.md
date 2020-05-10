## tools template get

Prints information about configured templates

### Synopsis

Prints information about configured templates

```
tools template get [flags]
```

### Examples

```

# prints all templates
argocd-notifications tools template get

# print YAML formatted app-sync-succeeded template definition
argocd-notifications tools template get app-sync-succeeded -o=yaml

```

### Options

```
  -h, --help            help for get
  -o, --output string   Output format. One of:json|yaml|wide|name (default "wide")
```

### Options inherited from parent commands

```
      --argocd-repo-server string      Argo CD repo server address (default "argocd-repo-server:8081")
      --as string                      Username to impersonate for the operation
      --as-group stringArray           Group to impersonate for the operation, this flag can be repeated to specify multiple groups.
      --certificate-authority string   Path to a cert file for the certificate authority
      --client-certificate string      Path to a client certificate file for TLS
      --client-key string              Path to a client key file for TLS
      --cluster string                 The name of the kubeconfig cluster to use
      --config-map string              argocd-notifications-cm.yaml file path
      --context string                 The name of the kubeconfig context to use
      --insecure-skip-tls-verify       If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
      --kubeconfig string              Path to a kube config. Only required if out-of-cluster
  -n, --namespace string               If present, the namespace scope for this CLI request
      --password string                Password for basic authentication to the API server
      --request-timeout string         The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests. (default "0")
      --secret string                  argocd-notifications-secret.yaml file path. Use empty secret if provided value is ':empty'
      --server string                  The address and port of the Kubernetes API server
      --token string                   Bearer token for authentication to the API server
      --user string                    The name of the kubeconfig user to use
      --username string                Username for basic authentication to the API server
```

## tools template notify

Generates notification using the specified template and send it to specified recipients

### Synopsis

Generates notification using the specified template and send it to specified recipients

```
tools template notify NAME APPLICATION [flags]
```

### Examples

```

# Trigger notification using in-cluster config map and secret
argocd-notifications tools template notify app-sync-succeeded guestbook --recipient slack:argocd-notifications

# Render notification render generated notification in console
argocd-notifications tools template notify app-sync-succeeded guestbook

```

### Options

```
  -h, --help                    help for notify
      --recipient stringArray   List of recipients (default [console:stdout])
```

### Options inherited from parent commands

```
      --argocd-repo-server string      Argo CD repo server address (default "argocd-repo-server:8081")
      --as string                      Username to impersonate for the operation
      --as-group stringArray           Group to impersonate for the operation, this flag can be repeated to specify multiple groups.
      --certificate-authority string   Path to a cert file for the certificate authority
      --client-certificate string      Path to a client certificate file for TLS
      --client-key string              Path to a client key file for TLS
      --cluster string                 The name of the kubeconfig cluster to use
      --config-map string              argocd-notifications-cm.yaml file path
      --context string                 The name of the kubeconfig context to use
      --insecure-skip-tls-verify       If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
      --kubeconfig string              Path to a kube config. Only required if out-of-cluster
  -n, --namespace string               If present, the namespace scope for this CLI request
      --password string                Password for basic authentication to the API server
      --request-timeout string         The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests. (default "0")
      --secret string                  argocd-notifications-secret.yaml file path. Use empty secret if provided value is ':empty'
      --server string                  The address and port of the Kubernetes API server
      --token string                   Bearer token for authentication to the API server
      --user string                    The name of the kubeconfig user to use
      --username string                Username for basic authentication to the API server
```

## tools trigger get

Prints information about configured triggers

### Synopsis

Prints information about configured triggers

```
tools trigger get [flags]
```

### Examples

```

# prints all triggers
argocd-notifications tools trigger get

# print YAML formatted on-sync-failed trigger definition
argocd-notifications tools trigger get on-sync-failed -o=yaml

```

### Options

```
  -h, --help            help for get
  -o, --output string   Output format. One of:json|yaml|wide|name (default "wide")
```

### Options inherited from parent commands

```
      --argocd-repo-server string      Argo CD repo server address (default "argocd-repo-server:8081")
      --as string                      Username to impersonate for the operation
      --as-group stringArray           Group to impersonate for the operation, this flag can be repeated to specify multiple groups.
      --certificate-authority string   Path to a cert file for the certificate authority
      --client-certificate string      Path to a client certificate file for TLS
      --client-key string              Path to a client key file for TLS
      --cluster string                 The name of the kubeconfig cluster to use
      --config-map string              argocd-notifications-cm.yaml file path
      --context string                 The name of the kubeconfig context to use
      --insecure-skip-tls-verify       If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
      --kubeconfig string              Path to a kube config. Only required if out-of-cluster
  -n, --namespace string               If present, the namespace scope for this CLI request
      --password string                Password for basic authentication to the API server
      --request-timeout string         The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests. (default "0")
      --secret string                  argocd-notifications-secret.yaml file path. Use empty secret if provided value is ':empty'
      --server string                  The address and port of the Kubernetes API server
      --token string                   Bearer token for authentication to the API server
      --user string                    The name of the kubeconfig user to use
      --username string                Username for basic authentication to the API server
```

## tools trigger run

Evaluates specified trigger condition and prints the result

### Synopsis

Evaluates specified trigger condition and prints the result

```
tools trigger run NAME APPLICATION [flags]
```

### Examples

```

# Execute trigger configured in 'argocd-notification-cm' ConfigMap
argocd-notifications tools trigger run on-sync-status-unknown ./sample-app.yaml

# Execute trigger using argocd-notifications-cm.yaml instead of 'argocd-notification-cm' ConfigMap
argocd-notifications tools trigger run on-sync-status-unknown ./sample-app.yaml \
    --config-map ./argocd-notifications-cm.yaml
```

### Options

```
  -h, --help   help for run
```

### Options inherited from parent commands

```
      --argocd-repo-server string      Argo CD repo server address (default "argocd-repo-server:8081")
      --as string                      Username to impersonate for the operation
      --as-group stringArray           Group to impersonate for the operation, this flag can be repeated to specify multiple groups.
      --certificate-authority string   Path to a cert file for the certificate authority
      --client-certificate string      Path to a client certificate file for TLS
      --client-key string              Path to a client key file for TLS
      --cluster string                 The name of the kubeconfig cluster to use
      --config-map string              argocd-notifications-cm.yaml file path
      --context string                 The name of the kubeconfig context to use
      --insecure-skip-tls-verify       If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
      --kubeconfig string              Path to a kube config. Only required if out-of-cluster
  -n, --namespace string               If present, the namespace scope for this CLI request
      --password string                Password for basic authentication to the API server
      --request-timeout string         The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests. (default "0")
      --secret string                  argocd-notifications-secret.yaml file path. Use empty secret if provided value is ':empty'
      --server string                  The address and port of the Kubernetes API server
      --token string                   Bearer token for authentication to the API server
      --user string                    The name of the kubeconfig user to use
      --username string                Username for basic authentication to the API server
```

