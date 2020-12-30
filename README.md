[![codecov](https://codecov.io/gh/argoproj-labs/argocd-notifications/branch/master/graph/badge.svg)](https://codecov.io/gh/argoproj-labs/argocd-notifications)

:bangbang: | v1.0 release is work in progress. See what is comming in [release-1.0](https://github.com/argoproj-labs/argocd-notifications/tree/release-1.0) branch
:---: | :---

# Argo CD Notifications

Argo CD Notifications continuously monitors Argo CD applications and provides a flexible way to notify
users about important changes in the applications state. The project includes a bundle of useful
built-in triggers and notification templates, integrates with various notification services such as
☑ Slack, ☑ SMTP and plans to support Telegram, Discord, etc.

![demo](./docs/demo.gif)

# Why use Argo CD Notifications?

The Argo CD Notifications is not the only way to monitor Argo CD application. You might leverage Prometheus
metrics and [Grafana Alerts](https://grafana.com/docs/grafana/latest/alerting/rules/) or projects
like [bitnami-labs/kubewatch](https://github.com/bitnami-labs/kubewatch) and
[argo-kube-notifier](https://github.com/argoproj-labs/argo-kube-notifier). The advantage of Argo CD Notifications is that
it is focused on Argo CD use cases and ultimately provides a better user experience.

# Documentation

Go to the complete [documentation](https://argoproj-labs.github.io/argocd-notifications/) to learn more.
