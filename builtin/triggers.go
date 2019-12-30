package builtin

import (
	"github.com/argoproj-labs/argocd-notifications/triggers"
	"k8s.io/utils/pointer"
)

var (
	Triggers = []triggers.NotificationTrigger{{
		Name:        "on-sync-status-unknown",
		Condition:   "app.status.sync.status == 'Unknown'",
		Description: "Application status is 'Unknown'",
		Template:    "app-sync-status",
		Enabled:     pointer.BoolPtr(false),
	}, {
		Name:        "on-sync-failed",
		Condition:   "app.status.operationState.phase in ['Error', 'Failed']",
		Description: "Application syncing has failed",
		Template:    "app-sync-failed",
		Enabled:     pointer.BoolPtr(false),
	}, {
		Name:        "on-sync-running",
		Condition:   "app.status.operationState.phase in ['Running']",
		Description: "Application is being synced",
		Template:    "app-sync-running",
		Enabled:     pointer.BoolPtr(false),
	}, {
		Name:        "on-sync-succeeded",
		Condition:   "app.status.operationState.phase in ['Succeeded']",
		Description: "Application syncing has succeeded",
		Template:    "app-sync-succeeded",
		Enabled:     pointer.BoolPtr(false),
	}, {
		Name:        "on-health-degraded",
		Condition:   "app.status.health.status == 'Degraded'",
		Description: "Application has degraded",
		Template:    "app-health-degraded",
		Enabled:     pointer.BoolPtr(false),
	}}
)
