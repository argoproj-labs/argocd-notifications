package settings

import (
	"testing"

	"github.com/argoproj-labs/argocd-notifications/notifiers"
	"github.com/argoproj-labs/argocd-notifications/triggers"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/utils/pointer"
)

func TestParseLegacyConfigMap(t *testing.T) {
	configData := map[string]string{
		"config.yaml": `
subscriptions:
- recipients:
    - slack:test
  triggers:
    - on-sync-status-custom
triggers:
  - name: on-sync-status-custom
    condition: app.status.operationState.phase in ['Custom']
    description: Application custom trigger
    template: app-sync-status
    enabled: true
templates:
  - name: app-sync-status
    title: Application {{.app.metadata.name}} sync status is {{.app.status.sync.status}}
    body: |
      Application {{.app.metadata.name}} sync is {{.app.status.sync.status}}.
      Application details: {{.context.argocdUrl}}/applications/{{.app.metadata.name}}.
    slack:
      attachments: |
        [{
          "title": "{{.app.metadata.name}}",
          "title_link": "{{.context.argocdUrl}}/applications/{{.app.metadata.name}}",
          "color": "#18be52",
          "fields": [{
            "title": "Sync Status",
            "value": "{{.app.status.sync.status}}",
            "short": true
          }, {
            "title": "Repository",
            "value": "{{.app.spec.source.repoURL}}",
            "short": true
          }]
        }]
context:
    argocdUrl: testUrl`}

	expectCfg := &Config{
		Subscriptions: []Subscription{{
			Recipients: []string{"slack:test"},
			Triggers:   []string{"on-sync-status-custom"},
			Selector:   labels.NewSelector(),
		}},
		Triggers: []triggers.NotificationTrigger{
			{
				Name:        "on-sync-status-custom",
				Condition:   "app.status.operationState.phase in ['Custom']",
				Description: "Application custom trigger",
				Template:    "app-sync-status",
				Enabled:     pointer.BoolPtr(true),
			},
		},
		Templates: []triggers.NotificationTemplate{{
			Name: "app-sync-status",
			Notification: notifiers.Notification{
				Title: "Application {{.app.metadata.name}} sync status is {{.app.status.sync.status}}",
				Body: `Application {{.app.metadata.name}} sync is {{.app.status.sync.status}}.
Application details: {{.context.argocdUrl}}/applications/{{.app.metadata.name}}.
`,
				Slack: &notifiers.SlackNotification{
					Attachments: `[{
  "title": "{{.app.metadata.name}}",
  "title_link": "{{.context.argocdUrl}}/applications/{{.app.metadata.name}}",
  "color": "#18be52",
  "fields": [{
    "title": "Sync Status",
    "value": "{{.app.status.sync.status}}",
    "short": true
  }, {
    "title": "Repository",
    "value": "{{.app.spec.source.repoURL}}",
    "short": true
  }]
}]
`,
					Blocks: "",
				}},
		}},
		Context: map[string]string{"argocdUrl": "testUrl"},
	}
	actualCfg, err := ParseConfigMap(&v1.ConfigMap{Data: configData})
	assert.NoError(t, err)
	assert.Equal(t, expectCfg, actualCfg)
}
