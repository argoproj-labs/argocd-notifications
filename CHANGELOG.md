# Changelog

## v0.5.0 (Not released)

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
