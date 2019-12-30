# Changelog


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
