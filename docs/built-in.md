# Built-in Triggers and Templates
## Triggers
|          NAME          |            DESCRIPTION            |                  TEMPLATE                   |
|------------------------|-----------------------------------|---------------------------------------------|
| on-sync-status-unknown | Application status is 'Unknown'   | [app-sync-status](#app-sync-status)         |
| on-sync-failed         | Application syncing has failed    | [app-sync-failed](#app-sync-failed)         |
| on-sync-running        | Application is being synced       | [app-sync-running](#app-sync-running)       |
| on-sync-succeeded      | Application syncing has succeeded | [app-sync-succeeded](#app-sync-succeeded)   |
| on-health-degraded     | Application has degraded          | [app-health-degraded](#app-health-degraded) |

## Templates
### app-sync-status
**title**: `Application {{.app.metadata.name}} sync status is {{.app.status.sync.status}}`

**body**:
```
Application {{.app.metadata.name}} sync is {{.app.status.sync.status}}.
Application details: {{.context.argocdUrl}}/applications/{{.app.metadata.name}}.
```
### app-sync-succeeded
**title**: `Application {{.app.metadata.name}} has been successfully synced.`

**body**:
```
Application {{.app.metadata.name}} has been successfully synced at {{.app.status.operationState.finishedAt}}.
Sync operation details is available at: {{.context.argocdUrl}}/applications/{{.app.metadata.name}}?operation=true .
```
### app-sync-failed
**title**: `Failed to sync application {{.app.metadata.name}}.`

**body**:
```
The sync operation of application {{.app.metadata.name}} has failed at {{.app.status.operationState.finishedAt}} with the following error: {{.app.status.operationState.message}}
Sync operation details is available at: {{.context.argocdUrl}}/applications/{{.app.metadata.name}}?operation=true .
```
### app-sync-running
**title**: `Start syncing application {{.app.metadata.name}}.`

**body**:
```
The sync operation of application {{.app.metadata.name}} has started at {{.app.status.operationState.startedAt}}.
Sync operation details is available at: {{.context.argocdUrl}}/applications/{{.app.metadata.name}}?operation=true .
```
### app-health-degraded
**title**: `Application {{.app.metadata.name}} has degraded.`

**body**:
```
Application {{.app.metadata.name}} has degraded.
Application details: {{.context.argocdUrl}}/applications/{{.app.metadata.name}}.
```
