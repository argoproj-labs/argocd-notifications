# Triggers and Templates Catalog
## Triggers
|          NAME          |                          DESCRIPTION                          |                      TEMPLATE                       |
|------------------------|---------------------------------------------------------------|-----------------------------------------------------|
| on-deployed            | Application is synced and healthy. Triggered once per commit. | [app-deployed](#app-deployed)                       |
| on-health-degraded     | Application has degraded                                      | [app-health-degraded](#app-health-degraded)         |
| on-sync-failed         | Application syncing has failed                                | [app-sync-failed](#app-sync-failed)                 |
| on-sync-running        | Application is being synced                                   | [app-sync-running](#app-sync-running)               |
| on-sync-status-unknown | Application status is 'Unknown'                               | [app-sync-status-unknown](#app-sync-status-unknown) |
| on-sync-succeeded      | Application syncing has succeeded                             | [app-sync-succeeded](#app-sync-succeeded)           |

## Templates
### app-deployed
**definition**:
```yaml
email:
  subject: New version of an application {{.app.metadata.name}} is up and running.
message: |
  {{if eq .serviceType "slack"}}:white_check_mark:{{end}} Application {{.app.metadata.name}} is now running new version of deployments manifests.
slack:
  attachments: |
    [{
      "title": "{{ .app.metadata.name}}",
      "title_link":"{{.context.argocdUrl}}/applications/{{.app.metadata.name}}",
      "color": "#18be52",
      "fields": [
      {
        "title": "Sync Status",
        "value": "{{.app.status.sync.status}}",
        "short": true
      },
      {
        "title": "Repository",
        "value": "{{.app.spec.source.repoURL}}",
        "short": true
      },
      {
        "title": "Revision",
        "value": "{{.app.status.sync.revision}}",
        "short": true
      }
      {{range $index, $c := .app.status.conditions}}
      {{if not $index}},{{end}}
      {{if $index}},{{end}}
      {
        "title": "{{$c.type}}",
        "value": "{{$c.message}}",
        "short": true
      }
      {{end}}
      ]
    }]

```
### app-health-degraded
**definition**:
```yaml
email:
  subject: Application {{.app.metadata.name}} has degraded.
message: |
  {{if eq .serviceType "slack"}}:exclamation:{{end}} Application {{.app.metadata.name}} has degraded.
  Application details: {{.context.argocdUrl}}/applications/{{.app.metadata.name}}.
slack:
  attachments: |-
    [{
      "title": "{{ .app.metadata.name}}",
      "title_link": "{{.context.argocdUrl}}/applications/{{.app.metadata.name}}",
      "color": "#f4c030",
      "fields": [
      {
        "title": "Sync Status",
        "value": "{{.app.status.sync.status}}",
        "short": true
      },
      {
        "title": "Repository",
        "value": "{{.app.spec.source.repoURL}}",
        "short": true
      }
      {{range $index, $c := .app.status.conditions}}
      {{if not $index}},{{end}}
      {{if $index}},{{end}}
      {
        "title": "{{$c.type}}",
        "value": "{{$c.message}}",
        "short": true
      }
      {{end}}
      ]
    }]

```
### app-sync-failed
**definition**:
```yaml
email:
  subject: Failed to sync application {{.app.metadata.name}}.
message: |
  {{if eq .serviceType "slack"}}:exclamation:{{end}}  The sync operation of application {{.app.metadata.name}} has failed at {{.app.status.operationState.finishedAt}} with the following error: {{.app.status.operationState.message}}
  Sync operation details are available at: {{.context.argocdUrl}}/applications/{{.app.metadata.name}}?operation=true .
slack:
  attachments: |-
    [{
      "title": "{{ .app.metadata.name}}",
      "title_link":"{{.context.argocdUrl}}/applications/{{.app.metadata.name}}",
      "color": "#E96D76",
      "fields": [
      {
        "title": "Sync Status",
        "value": "{{.app.status.sync.status}}",
        "short": true
      },
      {
        "title": "Repository",
        "value": "{{.app.spec.source.repoURL}}",
        "short": true
      }
      {{range $index, $c := .app.status.conditions}}
      {{if not $index}},{{end}}
      {{if $index}},{{end}}
      {
        "title": "{{$c.type}}",
        "value": "{{$c.message}}",
        "short": true
      }
      {{end}}
      ]
    }]

```
### app-sync-running
**definition**:
```yaml
email:
  subject: Start syncing application {{.app.metadata.name}}.
message: |
  The sync operation of application {{.app.metadata.name}} has started at {{.app.status.operationState.startedAt}}.
  Sync operation details are available at: {{.context.argocdUrl}}/applications/{{.app.metadata.name}}?operation=true .
slack:
  attachments: |-
    [{
      "title": "{{ .app.metadata.name}}",
      "title_link":"{{.context.argocdUrl}}/applications/{{.app.metadata.name}}",
      "color": "#0DADEA",
      "fields": [
      {
        "title": "Sync Status",
        "value": "{{.app.status.sync.status}}",
        "short": true
      },
      {
        "title": "Repository",
        "value": "{{.app.spec.source.repoURL}}",
        "short": true
      }
      {{range $index, $c := .app.status.conditions}}
      {{if not $index}},{{end}}
      {{if $index}},{{end}}
      {
        "title": "{{$c.type}}",
        "value": "{{$c.message}}",
        "short": true
      }
      {{end}}
      ]
    }]

```
### app-sync-status-unknown
**definition**:
```yaml
email:
  subject: Application {{.app.metadata.name}} sync status is 'Unknown'
message: |
  {{if eq .serviceType "slack"}}:exclamation:{{end}} Application {{.app.metadata.name}} sync is 'Unknown'.
  Application details: {{.context.argocdUrl}}/applications/{{.app.metadata.name}}.
  {{if ne .serviceType "slack"}}
  {{range $c := .app.status.conditions}}
      * {{$c.message}}
  {{end}}
  {{end}}
slack:
  attachments: |-
    [{
      "title": "{{ .app.metadata.name}}",
      "title_link":"{{.context.argocdUrl}}/applications/{{.app.metadata.name}}",
      "color": "#E96D76",
      "fields": [
      {
        "title": "Sync Status",
        "value": "{{.app.status.sync.status}}",
        "short": true
      },
      {
        "title": "Repository",
        "value": "{{.app.spec.source.repoURL}}",
        "short": true
      }
      {{range $index, $c := .app.status.conditions}}
      {{if not $index}},{{end}}
      {{if $index}},{{end}}
      {
        "title": "{{$c.type}}",
        "value": "{{$c.message}}",
        "short": true
      }
      {{end}}
      ]
    }]

```
### app-sync-succeeded
**definition**:
```yaml
email:
  subject: Application {{.app.metadata.name}} has been successfully synced.
message: |
  {{if eq .serviceType "slack"}}:white_check_mark:{{end}} Application {{.app.metadata.name}} has been successfully synced at {{.app.status.operationState.finishedAt}}.
  Sync operation details are available at: {{.context.argocdUrl}}/applications/{{.app.metadata.name}}?operation=true .
slack:
  attachments: "[{\n  \"title\": \"{{ .app.metadata.name}}\",\n  \"title_link\":\"{{.context.argocdUrl}}/applications/{{.app.metadata.name}}\",\n  \"color\": \"#18be52\",\n  \"fields\": [\n  {\n    \"title\": \"Sync Status\",\n    \"value\": \"{{.app.status.sync.status}}\",\n    \"short\": true\n  },\n  {\n    \"title\": \"Repository\",\n    \"value\": \"{{.app.spec.source.repoURL}}\",\n    \"short\": true\n  }\n  {{range $index, $c := .app.status.conditions}}\n  {{if not $index}},{{end}}\n  {{if $index}},{{end}}\n  {\n    \"title\": \"{{$c.type}}\",\n    \"value\": \"{{$c.message}}\",\n    \"short\": true\n  }\n  {{end}}\n  ]\n}]    "

```
