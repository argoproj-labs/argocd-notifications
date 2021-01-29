The notification services represent integration with services such as slack, email or custom webhook. Services are configured in `argocd-notifications-cm` ConfigMap
using `service.<type>.(<custom-name>)` keys and might referense sensitive data from `argocd-notifications-secret` Secret. Following example demonstrates slack
service configuration:

```yaml
  service.slack: |
    token: $slack-token
```


The `slack` indicates that service sends slack notification; name is missing and defaults to `slack`.

## Sensitive Data

Sensitive data like authentication tokens should be stored in `argocd-notifications-secret` Secret and can be referenced in
service configuration using `$<secret-key>` format. For example `$slack-token` referencing value of key `slack-token` in
`argocd-notifications-secret` Secret.

## Custom Names

Service custom names allow configuring two instances of the same service type. For example, in addition to slack, you might register slack compatible service
that leverages [Mattermost](https://mattermost.com/):

```yaml
  #  Slack based notifier with name mattermost
  service.slack.mattermost: |
    apiURL: https://my-mattermost-url.com/api
    token: $slack-token
```

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  annotations:
    notifications.argoproj.io/subscribe.on-sync-succeeded.mattermost: my-channel
```

## Service Types

* [Email](./email.md)
* [Slack](./slack.md)
* [Opsgenie](./opsgenie.md)
* [Grafana](./grafana.md)
* [Webhook](./webhook.md)
* [Telegram](./telegram.md)
