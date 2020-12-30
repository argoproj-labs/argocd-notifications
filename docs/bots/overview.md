# Bot (v0.5)

The optional bot component simplifies managing subscriptions. The end users can use bot commands to manage subscriptions
even if they don't have access to the Kubernetes API and cannot modify annotations. 

The bot is not installed by default. Use the `install-bot.yaml` to intall it:

```
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj-labs/argocd-notifications/stable/manifests/install-bot.yaml
```

* [Slack bot](./slack-bot.md)
* [Opsgenie bot](./opsgenie-bot.md)
* [Telegram bot](./telegram-bot.md)