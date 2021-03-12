# Telegram

1. Get an API token using [@Botfather](https://t.me/Botfather).
2. Store token in `argocd_notifications-secret` Secret and configure telegram integration
in `argocd-notifications-cm` ConfigMap:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-notifications-cm
data:
  service.telegram: |
    token: $telegram-token
```

3. Create new Telegram [channel](https://telegram.org/blog/channels), this channel should be [public to have a username](https://telegram.org/faq_channels?ln=f#q-how-are-public-and-private-channels-different).
4. Add your bot as an administrator.
5. Use this channel `username` in the subscription for your Telegram integration:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  annotations:
    notifications.argoproj.io/subscribe.on-sync-succeeded.telegram: username
```
