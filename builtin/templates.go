package builtin

import "github.com/argoproj-labs/argocd-notifications/triggers"

var (
	Templates = []triggers.NotificationTemplate{{
		Name:  "app-sync-status",
		Title: "Application {{.app.metadata.name}} sync status is {{.app.status.sync.status}}",
		Body: `Application {{.app.metadata.name}} sync is {{.app.status.sync.status}}.
Application details: {{.context.argocdUrl}}/applications/{{.app.metadata.name}}.
{{range $c := .app.status.conditions}}
     * {{$c.message}}
{{end}}`,
	}, {
		Name:  "app-sync-succeeded",
		Title: "Application {{.app.metadata.name}} has been successfully synced.",
		Body: `Application {{.app.metadata.name}} has been successfully synced at {{.app.status.operationState.finishedAt}}.
Sync operation details is available at: {{.context.argocdUrl}}/applications/{{.app.metadata.name}}?operation=true .`,
	}, {
		Name:  "app-sync-failed",
		Title: "Failed to sync application {{.app.metadata.name}}.",
		Body: `The sync operation of application {{.app.metadata.name}} has failed at {{.app.status.operationState.finishedAt}} with the following error: {{.app.status.operationState.message}}
Sync operation details is available at: {{.context.argocdUrl}}/applications/{{.app.metadata.name}}?operation=true .`,
	}, {
		Name:  "app-sync-running",
		Title: "Start syncing application {{.app.metadata.name}}.",
		Body: `The sync operation of application {{.app.metadata.name}} has started at {{.app.status.operationState.startedAt}}.
Sync operation details is available at: {{.context.argocdUrl}}/applications/{{.app.metadata.name}}?operation=true .`,
	}, {
		Name:  "app-health-degraded",
		Title: "Application {{.app.metadata.name}} has degraded.",
		Body: `Application {{.app.metadata.name}} has degraded.
Application details: {{.context.argocdUrl}}/applications/{{.app.metadata.name}}.`,
	}}
)
