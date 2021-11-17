# Changelog

## v1.2.0 (Unreleased)

### Features

* feat: Subscribe to all triggers at once (#202)
* feat: Add .strings.ReplaceAll expression (#332)
* feat: Add .sync.GetInfoItem expression that simplifies retrieving operation info items by name
* feat: Plaintext connection to repo-server with ability disable TLS (#281)
* feat: Dynamic ConfigMap name and Secret name (#77)
* feat: Configurable path for slack bot (#94)
* feat: Support rocketchat 
* feat: Support google chat
* feat: Support alertmanager
* feat: Support pushover
* feat: Support email SendHtml
* feat: Add message aggregation feature by slack threads API
* feat: Add summary field into teams message
* feat: Support Markdown parse mode in telegram

### Bug Fixes

* fix: syntax error in Teams notifications (#271)
* fix: service account rbac issue , add namespace support for informer (#322)
* fix: add annotations nil check
* fix: add expr error log

### Other

* Move notification providers to notifications-engine library

## v1.1.0 (2021-04-17)

### Features

* feat: ArgoCD Notifications for Created, Deleted status (#231)
* feat: improve oncePer evaluate (#228)
* feat: support change timezone (#226)
* feat: support mattermost integration (#212)
* feat: support telegram private channel (#207)
* feat: GitHub App integration (#180)
* feat: MS Teams integration (#181)

### Bug Fixes

* fix: merging secrets into service config not working (fixes #208)
* fix: update cached informer object instead of reloading app to avoid duplicated notifications (#204)
* fix: static configmap and secret binding (#136)

## v1.0.2 (2020-02-17)

* fix: revision changes only if someone run sync operation or changes are detected (#157)
* fix: if app has no subscriptions, then nothing to process (#174)
* fix:improve annotation iterate (#159)

## v1.0.1 (2020-01-20)

* fix: the on-deployed trigger sends multiple notifications (#154)

## v1.0.0 (2020-01-19)

### Features

* feat: triggers with multiple conditions and multiple templates per condition
* feat: support `oncePer` trigger property that allows sending notification "once per" app field value (#60)
* feat: add support for proxy settings (#42)
* feat: support self-signed certificates in all HTTP based integrations (#61)
* feat: subscription support specifying message template
* feat: support Telegram notifications (#49)

### Bug Fixes

* Failed notifications affect multiple subscribers (#79)

### Refactor

* Built-in triggers/templates replaced with triggers/templates "catalog" (#56)
* `config.yaml` and `notifiers.yaml` configs split into multiple ConfigMap keys (#76)
* `trigger.enabled` field is replaced with `defaultTriggers` setting
* Replace `template.body`, `template.title` fields with `template.message` and `template.email.subject` fields

## v0.7.0 (2020-05-10)

### Features

* feat: support default subscriptions
* feat: support loading commit metadata (#87)
* feat: add controller prometheus metrics (#86)
* feat: log http request/response in every notifier (#83)
* feat: add CLI debugging commands (#81)

### Bug Fixes

* fix: don't append slash to webhook url (#70) (#85)
* fix: improve settings parsing errors (#84)
* fix: use strategic merge patch to merge built-in and user provided config (#74)
* fix: ensure slack attachment properly formatted json object (#73)

## v0.6.1 (2020-03-20)

### Features

* feat: support default subscriptions
* fix: ensure slack attachment properly formatted json object

## v0.6.0 (2020-03-13)

### Features

* feat: support sending the generic webhook request
* feat: Grafana annotation notifier ( thanks to [nhuray](https://github.com/nhuray) )

###  Bug Fixes

* fix: wait for next reconciliation after sync ( thanks to [sboschman](https://github.com/sboschman) )

## v0.5.0 (2020-03-01)

### Features
* feat: support managing subscriptions using Slack bot
* feat: support `time.Now()` and `time.Parse(...)` in trigger condition ( thanks to [@HatsuneMiku3939](https://github.com/HatsuneMiku3939) )
* feat: Add icon emoij and icon url support for Slack messages ( thanks to [sboschman](https://github.com/sboschman) )
* feat: Introduce sprig functions to templates( thanks to [imranismail](https://github.com/imranismail) )

###  Bug Fixes
* fix: fix null pointer dereference error while config parsing

## v0.4.2 (2020-02-03)

###  Bug Fixes
* fix: fix null pointer dereference error while config parsing

## v0.4.1 (2020-01-26)

###  Bug Fixes
* fix: notification config parse (#19) ( thanks to [@yutachaos](https://github.com/yutachaos) ! )

## v0.4.0 (2020-01-24)

* Opsgenie support (thanks to [Dominik Münch](https://github.com/muenchdo))
* Slack message blocks and attachments support

## v0.3.0 (2020-01-13)

### Features
* Trigger specific subscriptions

### Other
* Move repo and docker image to https://github.com/argoproj-labs/argocd-notifications

## v0.2.1 (2019-12-29)

### Bug Fixes
* built-in triggers are disabled by default

## v0.2.0 (2019-12-26)

### Features
* support setting hot reload
* embed built-in triggers/templates into binary instead of default config map
* support enabling/disabling triggers
* support customizing built-in triggers/templates
* add on-sync-running/on-sync-succeeded triggers and templates

### Bug Fixes
* fix sending same notification twice

### Other
* use `scratch` as a base image 

## v0.1.0 (2019-12-14)

First MVP:
- email, slack notifications
- subscribe at application/project level
