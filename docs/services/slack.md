## Configuration

1. Create Slack Application using https://api.slack.com/apps?new_app=1
![1](https://user-images.githubusercontent.com/426437/73604308-4cb0c500-4543-11ea-9092-6ca6bae21cbb.png)
1. Once application is created navigate to `Enter OAuth & Permissions`
![2](https://user-images.githubusercontent.com/426437/73604309-4d495b80-4543-11ea-9908-4dea403d3399.png)
1. Click `Permissions` under `Add features and functionality` section and add `chat:write:bot` scope. To use the optional username and icon overrides in the Slack notification service also add the `chat:write.customize` scope.
![3](https://user-images.githubusercontent.com/426437/73604310-4d495b80-4543-11ea-8576-09cd91aea0e5.png)
1. Scroll back to the top, click 'Install App to Workspace' button and confirm the installation.
![4](https://user-images.githubusercontent.com/426437/73604311-4d495b80-4543-11ea-9155-9d216b20ec86.png)
1. Once installation is completed copy the OAuth token. 
![5](https://user-images.githubusercontent.com/426437/73604312-4d495b80-4543-11ea-832b-a9d9d5e4bc29.png)

1. Finally use the OAuth token to configure the slack integration in the `argocd-notifications-secret` secret: 

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-notifications-cm
data:
  service.slack: |
    apiURL: <url>                 # optional URL, e.g. https://my-mattermost-url.com/api
    token: $slack-token
    username: <override-username> # optional username
    icon: <override-icon> # optional icon for the message (supports both emoij and url notation)
```

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: argocd-notifications-secret
stringData:
  token: <auth-token>
```

## Templates

Notification templates can be customized to leverage slack message blocks and attachments
[feature](https://api.slack.com/messaging/composing/layouts).

![](https://user-images.githubusercontent.com/426437/72776856-6dcef880-3bc8-11ea-8e3b-c72df16ee8e6.png)

The message blocks and attachments can be specified in `blocks` and `attachments` string fields under `slack` field:

```yaml
template.app-sync-status: |
  message: |
    Application {{.app.metadata.name}} sync is {{.app.status.sync.status}}.
    Application details: {{.context.argocdUrl}}/applications/{{.app.metadata.name}}.
  slack:
    attachments: |
      [{
        "title": "{{.app.metadata.name}}",
        "title_link": "{{.context.argocdUrl}}/applications/{{.app.metadata.name}}",
        "color": "#18be52",
        "fields": [{
          "title": "Sync Status",
          "value": "{{.app.status.sync.status}}",
          "short": true
        }, {
          "title": "Repository",
          "value": "{{.app.spec.source.repoURL}}",
          "short": true
        }]
      }]
```