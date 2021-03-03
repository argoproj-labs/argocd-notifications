# Teams

1. Open `Teams` and goto `Apps`
2. Find `Incoming Webhook` microsoft app and click on it
3. Press `Add to a team` -> select team and channel -> press `Set up a connector`
4. Enter webhook name and upload image (optional)
5. Press `Create` then copy webhook url and store it in `argocd_notifications-secret`
6. in `argocd-notifications-cm` ConfigMap:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-notifications-cm
data:
  service.teams: |
    recipientUrls: 
      channelName: $channel-teams-url
```

7. Create subscription for your Teams integration:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  annotations:
    notifications.argoproj.io/subscribe.on-sync-succeeded.teams: channelName
```