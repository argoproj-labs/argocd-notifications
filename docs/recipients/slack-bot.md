# Slack bot (v0.5)

The Slack bot leverages [slash commands](https://api.slack.com/interactivity/slash-commands). The bot allows slack users
to view existing channel subscriptions and subscribe or unsubscribe channels.

1. Make sure bot component is [installed](./bot.md).
1. Configure slack [integration](../services/slack.md).
1. In the slack application settings page navigate to the 'Slash Commands' section and click 'Create New Command' button.
1. Fill in new slack command details
![image](https://user-images.githubusercontent.com/426437/75645798-2e022480-5bfc-11ea-8682-5ce362bdcc9a.png)
1. In the slack application settings page navigate to the 'Basic Information' section and copy 'Signing Secret' from the 'App Credentials' section.
1. Add `signingSecret` to the slack configuration in the `notifiers.yaml` field of the `argocd-notification-secret`:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: argocd-notifications-secret
stringData:
  notifiers.yaml: |
    slack:
      token: <my-token>
      signingSecret: <my-secret>
```

## Commands

The bot supports following commands:

* `list-subscriptions` - list channel subscriptions
* `subscribe <my-app> <optional-trigger>` - subscribes channel to the app notifications
* `subscribe proj:<my-app> <optional-trigger>` - subscribes channel to the app project notifications
* `unsubscribe <my-app> <optional-trigger>` - unsubscribes channel frm the app notifications
* `unsubscribe proj:<my-app> <optional-trigger>` - unsubscribes channel frm the app project notifications