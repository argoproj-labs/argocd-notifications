## Troubleshooting
(v0.7.0)

The `argocd-notifications` binary includes a set of CLI commands that helps to configure the controller
settings and troubleshooting issues. All CLI commands are available as `argocd-notifications tools` sub-commands: 

```bash
argocd-notifications tools <sub-command-name>
```

## Global flags
Following global flags are available for all sub-commands:
* `config-map` - path to the file containing `argocd-notifications-cm` ConfigMap. If not specified
then the command loads `argocd-notification-cm` ConfigMap using the local Kubernetes config file.
* `secret` - path to the file containing `argocd-notifications-secret` ConfigMap. If not
specified then the command loads `argocd-notification-secret` Secret using the local Kubernetes config file.
Additionally, you can specify `:empty` value to use empty secret with no notification service settings. 

**Examples:**

* Get list of triggers configured in the local config map:

```
argocd-notifications tools trigger get \
  --config-map ./argocd-notifications-cm.yaml --secret :empty
```

* Trigger notification using in-cluster config map and secret:

```
argocd-notifications tools template notify \
  app-sync-succeeded guestbook --recipient slack:argocd-notifications
```

## How to use it

### On your laptop

The binary is available in `argoprojlabs/argocd-notifications` image. Use the `docker run` and volume mount
to execute binary on any platform. 

**Example:**

```bash
docker run --rm -it -w /src -v $(pwd):/src \
  argoprojlabs/argocd-notifications:<version> \
  /app/argocd-notifications tools trigger get \
  --config-map ./argocd-notifications-cm.yaml --secret :empty
```

### In your cluster

SSH into the running `argocd-notifications-controller` pod and use `kubectl exec` command to validate in-cluster
configuration.

**Example**
```bash
kubectl exec -it argocd-notifications-controller-<pod-hash> \
  /app/argocd-notifications tools trigger get
```

## Commands

{!troubleshooting-commands.md!}

