# Overview

The recipients are configured using `Application` or `AppProject` CRD annotations. 

Each recipient is prefixed with the [notification service type](../services/overview.md) such as `slack` or `email`. Multiple recipients are separated with a comma, e.g.

```yaml
recipients.argocd-notifications.argoproj.io: email:<sample-email>, slack:<sample-channel-name>
```

The example below demonstrates how to subscribe to the email notifications triggered for a specific application:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  annotations:
    recipients.argocd-notifications.argoproj.io: email:<sample-email>
```

The example below demonstrates how to get to the Slack message on a notification of the any project application:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: AppProject
metadata:
  annotations:
    recipients.argocd-notifications.argoproj.io: slack:<sample-channel-name>
```

The example below demonstrates how to create a Grafana annotation for a specific application:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  annotations:
    recipients.argocd-notifications.argoproj.io: grafana:tag1|tag2
```

## Trigger Specific Subscription (v0.3)

It is possible to subscribe recipient to a specific trigger instead of all triggers. The annotation key should be
prefixed with `<trigger-name>.`. The example below demonstrates how to receive only `on-sync-failed` trigger
notifications:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  annotations:
    on-sync-failed.recipients.argocd-notifications.argoproj.io: email:<sample-email>
```

## Default Subscriptions (v0.6.1)

The recipients might be configured globally in the `argocd-notifications-cm` ConfigMap. The default subscriptions
are applied to all applications and triggers by default. The trigger and applications might be configured using the
`trigger` and `selector` fields:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-notifications-cm
data:
  config.yaml: |
    subscriptions:
    # global subscription for all type of notifications
    - recipients:
      - slack:test1
      - webhook:github
    # subscription for on-sync-status-unknown trigger notifications
    - recipients:
      - slack:test2
      - email:test@gmail.com
      triggers:
      - on-sync-status-unknown
    # global subscription restricted to applications with matching labels only
    - recipients: slack:test3
      selector: test=true
```
 
## Manage subscriptions using bots

The [bot](./bot.md) component simplifies managing subscriptions.