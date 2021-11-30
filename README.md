[![codecov](https://codecov.io/gh/argoproj-labs/argocd-notifications/branch/master/graph/badge.svg)](https://codecov.io/gh/argoproj-labs/argocd-notifications)

# Argo CD Notifications is now part of Argo CD

**This project has moved to the main Argo CD repository**

This repository is no longer active. The Argo CD notifications project is now merged with [Argo CD](https://github.com/argoproj/argo-cd) and released along with it.
Further development will happen there.
See [https://argo-cd.readthedocs.io/en/latest/operator-manual/notifications/](https://argo-cd.readthedocs.io/en/latest/operator-manual/notifications/) for more details.

# Argo CD Notifications - OLD README

Argo CD Notifications continuously monitors Argo CD applications and provides a flexible way to notify
users about important changes in the applications state. The project includes a bundle of useful
built-in triggers and notification templates, integrates with various notification services such as
☑ Slack, ☑ SMTP, ☑ Opsgenie, ☑ Telegram and anything else using custom webhooks.

![demo](./docs/demo.gif)

# Why use Argo CD Notifications?

The Argo CD Notifications is not the only way to monitor Argo CD application. You might leverage Prometheus
metrics and [Grafana Alerts](https://grafana.com/docs/grafana/latest/alerting/rules/) or projects
like [bitnami-labs/kubewatch](https://github.com/bitnami-labs/kubewatch) and
[argo-kube-notifier](https://github.com/argoproj-labs/argo-kube-notifier). The advantage of Argo CD Notifications is that
it is focused on Argo CD use cases and ultimately provides a better user experience.

# Old Documentation

Go to the complete [documentation](https://argoproj-labs.github.io/argocd-notifications/) to learn more.
