# Built-in Triggers and Templates
## Triggers
|          NAME          |            DESCRIPTION            |                      TEMPLATE                       |
|------------------------|-----------------------------------|-----------------------------------------------------|
| on-health-degraded     | Application has degraded          | [app-health-degraded](#app-health-degraded)         |
| on-sync-failed         | Application syncing has failed    | [app-sync-failed](#app-sync-failed)                 |
| on-sync-running        | Application is being synced       | [app-sync-running](#app-sync-running)               |
| on-sync-status-unknown | Application status is 'Unknown'   | [app-sync-status-unknown](#app-sync-status-unknown) |
| on-sync-succeeded      | Application syncing has succeeded | [app-sync-succeeded](#app-sync-succeeded)           |

## Templates
### app-health-degraded
**title**: `Application {{.app.metadata.name}} has degraded.`

**body**:
```
{{if eq .serviceType "slack"}}:exclamation:{{end}} Application {{.app.metadata.name}} has degraded.
Application details: {{.context.argocdUrl}}/applications/{{.app.metadata.name}}.

```
### app-sync-failed
**title**: `Failed to sync application {{.app.metadata.name}}.`

**body**:
```
{{if eq .serviceType "slack"}}:exclamation:{{end}}  The sync operation of application {{.app.metadata.name}} has failed at {{.app.status.operationState.finishedAt}} with the following error: {{.app.status.operationState.message}}
Sync operation details are available at: {{.context.argocdUrl}}/applications/{{.app.metadata.name}}?operation=true .

```
### app-sync-running
**title**: `Start syncing application {{.app.metadata.name}}.`

**body**:
```
The sync operation of application {{.app.metadata.name}} has started at {{.app.status.operationState.startedAt}}.
Sync operation details are available at: {{.context.argocdUrl}}/applications/{{.app.metadata.name}}?operation=true .

```
### app-sync-status-unknown
**title**: `Application {{.app.metadata.name}} sync status is 'Unknown'`

**body**:
```
{{if eq .serviceType "slack"}}:exclamation:{{end}} Application {{.app.metadata.name}} sync is 'Unknown'.
Application details: {{.context.argocdUrl}}/applications/{{.app.metadata.name}}.
{{if ne .serviceType "slack"}}
{{range $c := .app.status.conditions}}
    * {{$c.message}}
{{end}}
{{end}}

```
### app-sync-succeeded
**title**: `Application {{.app.metadata.name}} has been successfully synced.`

**body**:
```
{{if eq .serviceType "slack"}}:white_check_mark:{{end}} Application {{.app.metadata.name}} has been successfully synced at {{.app.status.operationState.finishedAt}}.
Sync operation details are available at: {{.context.argocdUrl}}/applications/{{.app.metadata.name}}?operation=true .

```
