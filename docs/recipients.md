# Recipients

The list of recipients is not stored in a centralized configuration file. Instead, recipients might be configured using
`Application` or `AppProject` CRD annotations. The example below demonstrates how to subscribe to the email 
notifications triggered for a specific application:

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

Each recipient is prefixed with the [notification service type](./services/overview.md) such as `slack` or `email`.

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

