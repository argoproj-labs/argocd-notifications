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

3. Create new Telegram [channel](https://telegram.org/blog/channels) and add your bot as an administrator.
4. Create subscription for your Telegram integration:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  annotations:
    notifications.argoproj.io/subscribe.on-sync-succeeded.telegram: my_channel
```