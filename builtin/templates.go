package builtin

import (
	"fmt"
	"github.com/argoproj-labs/argocd-notifications/notifiers"
	"github.com/argoproj-labs/argocd-notifications/triggers"
)

const (
	slackAttachmentTemplate = `[{
	"title": "{{.app.metadata.name}}",
	"title_link": "{{.context.argocdUrl}}/applications/{{.app.metadata.name}}",
	"color": "%s",
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
	{
		"title": "{{$c.type}}",
		"value": "{{$c.message}}",
		"short": true
	}
	{{end}}
	]
}]`
)

var (
	slackAttachmentSuccess     = fmt.Sprintf(slackAttachmentTemplate, "#18be52")
	slackAttachmentWarning     = fmt.Sprintf(slackAttachmentTemplate, "#f4c030")
	slackAttachmentProgressing = fmt.Sprintf(slackAttachmentTemplate, "#0DADEA")
	slackAttachmentFailed      = fmt.Sprintf(slackAttachmentTemplate, "#E96D76")

	Templates = []triggers.NotificationTemplate{{
		Name: "app-sync-status-unknown",
		Notification: notifiers.Notification{
			Slack: &notifiers.SlackSpecific{Attachments: slackAttachmentFailed},
			Title: "Application {{.app.metadata.name}} sync status is 'Unknown'",
			Body: `{{if eq .context.notificationType "slack"}}:exclamation:{{end}} Application {{.app.metadata.name}} sync is 'Unknown'.
Application details: {{.context.argocdUrl}}/applications/{{.app.metadata.name}}.
{{if ne .context.notificationType "slack"}}
{{range $c := .app.status.conditions}}
     * {{$c.message}}
{{end}}
{{end}}`,
		}}, {
		Name: "app-sync-succeeded",
		Notification: notifiers.Notification{
			Slack: &notifiers.SlackSpecific{Attachments: slackAttachmentSuccess},
			Title: "Application {{.app.metadata.name}} has been successfully synced.",
			Body: `{{if eq .context.notificationType "slack"}}:white_check_mark:{{end}} Application {{.app.metadata.name}} has been successfully synced at {{.app.status.operationState.finishedAt}}.
Sync operation details are available at: {{.context.argocdUrl}}/applications/{{.app.metadata.name}}?operation=true .`,
		}}, {
		Name: "app-sync-failed",
		Notification: notifiers.Notification{
			Slack: &notifiers.SlackSpecific{Attachments: slackAttachmentFailed},
			Title: "Failed to sync application {{.app.metadata.name}}.",
			Body: `{{if eq .context.notificationType "slack"}}:exclamation:{{end}}  The sync operation of application {{.app.metadata.name}} has failed at {{.app.status.operationState.finishedAt}} with the following error: {{.app.status.operationState.message}}
Sync operation details are available at: {{.context.argocdUrl}}/applications/{{.app.metadata.name}}?operation=true .`,
		}}, {
		Name: "app-sync-running",
		Notification: notifiers.Notification{
			Slack: &notifiers.SlackSpecific{Attachments: slackAttachmentProgressing},
			Title: "Start syncing application {{.app.metadata.name}}.",
			Body: `The sync operation of application {{.app.metadata.name}} has started at {{.app.status.operationState.startedAt}}.
Sync operation details are available at: {{.context.argocdUrl}}/applications/{{.app.metadata.name}}?operation=true .`,
		}}, {
		Name: "app-health-degraded",
		Notification: notifiers.Notification{
			Slack: &notifiers.SlackSpecific{Attachments: slackAttachmentFailed},
			Title: "Application {{.app.metadata.name}} has degraded.",
			Body: `{{if eq .context.notificationType "slack"}}:exclamation:{{end}} Application {{.app.metadata.name}} has degraded.
Application details: {{.context.argocdUrl}}/applications/{{.app.metadata.name}}.`,
		}}}
)
