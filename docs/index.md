# Overview

Argo CD Notifications continuously monitors Argo CD applications and provides a flexible way to notify
users about important changes in the application state. Using flexible mechanism of
[triggers and templates](./triggers_and_templates/index.md) you can configure when the notification should be sent as well as notification content.
Argo CD Notifications includes the set of useful [built-in](./built-in.md) triggers and templates.
So you can just enable them instead of reinventing new ones.

## Getting Started

* Install Argo CD Notifications

```
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj-labs/argocd-notifications/stable/manifests/install.yaml
```

* Configure integration with your Slack in `argocd-notifications-secret` secret:

```bash
kubectl apply -n argocd -f - << EOF
apiVersion: v1
kind: Secret
metadata:
  name: argocd-notifications-secret
stringData:
  notifiers.yaml: |
    slack:
      token: <my-token>
type: Opaque
EOF
```
* Enable built-in trigger in the `argocd-notifications-cm` config map:

```bash
kubectl apply -n argocd -f - << EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-notifications-cm
data:
  config.yaml: |
    triggers:
      - name: on-sync-succeeded
        enabled: true
EOF
```

* Subscribe to notifications by adding the `recipients.argocd-notifications.argoproj.io` annotation to the Argo CD
application or project:

```bash
kubectl patch app <my-app> -n argocd -p '{"metadata": {"annotations": {"recipients.argocd-notifications.argoproj.io":"slack:<my-channel>"}}}' --type merge
```

Try syncing and application and get the notification once sync is completed.
