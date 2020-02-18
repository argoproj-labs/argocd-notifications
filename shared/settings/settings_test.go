package settings

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/utils/pointer"

	"github.com/argoproj-labs/argocd-notifications/notifiers"
	"github.com/argoproj-labs/argocd-notifications/triggers"
)

func TestParseSecret(t *testing.T) {
	notifiersData := []byte(`
email:
  host: smtp.gmail.com
  port: 587
  from: <myemail>@gmail.com
  username: <myemail>@gmail.com
  password: <mypassword>
slack:
  token: <my-token>
  username: <override-username>
opsgenie:
  apiUrl: api.opsgenie.com
  apiKeys:
    <team-id>: <my-api-key>`)

	expectNotifiersCfg := notifiers.Config{
		Email: &notifiers.EmailOptions{
			Host:               "smtp.gmail.com",
			Port:               587,
			From:               "<myemail>@gmail.com",
			InsecureSkipVerify: false,
			Username:           "<myemail>@gmail.com",
			Password:           "<mypassword>",
		},
		Slack: &notifiers.SlackOptions{
			Username:           "<override-username>",
			Token:              "<my-token>",
			Channels:           nil,
			InsecureSkipVerify: false,
		},
		Opsgenie: &notifiers.OpsgenieOptions{
			ApiUrl:  "api.opsgenie.com",
			ApiKeys: map[string]string{"<team-id>": "<my-api-key>"},
		},
	}
	actualNotifiersCfg, err := ParseSecret(&v1.Secret{Data: map[string][]byte{"notifiers.yaml": notifiersData}})
	assert.NoError(t, err)
	assert.Equal(t, expectNotifiersCfg, actualNotifiersCfg)
}

func TestParseConfigMap(t *testing.T) {
	configData := map[string]string{
		"config.yaml": `
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
				Slack: &notifiers.SlackSpecific{
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

func TestParseConfigMap_EmptyMap(t *testing.T) {
	cfg, err := ParseConfigMap(&v1.ConfigMap{})
	assert.NoError(t, err)
	assert.Empty(t, cfg.Templates)
	assert.Empty(t, cfg.Triggers)
	assert.Empty(t, cfg.Context)
}

func TestMergeConfigTemplate(t *testing.T) {
	cfg := Config{
		Templates: []triggers.NotificationTemplate{{
			Name: "foo",
			Notification: notifiers.Notification{
				Body:  "the body",
				Title: "the title",
				Slack: &notifiers.SlackSpecific{
					Attachments: "attachments",
				},
			},
		}},
	}
	merged := cfg.Merge(&Config{
		Templates: []triggers.NotificationTemplate{{
			Name: "foo",
			Notification: notifiers.Notification{
				Body: "new body",
				Slack: &notifiers.SlackSpecific{
					Blocks: "blocks",
				},
			},
		}, {
			Name: "bar",
			Notification: notifiers.Notification{
				Body:  "the body",
				Title: "the title",
			},
		}},
	})
	assert.ElementsMatch(t, merged.Templates, []triggers.NotificationTemplate{{
		Name: "foo",
		Notification: notifiers.Notification{
			Title: "the title",
			Body:  "new body",
			Slack: &notifiers.SlackSpecific{
				Attachments: "attachments",
				Blocks:      "blocks",
			},
		},
	}, {
		Name: "bar",
		Notification: notifiers.Notification{
			Body:  "the body",
			Title: "the title",
		},
	}})
}

func TestMergeConfigTriggers(t *testing.T) {
	cfg := Config{
		Triggers: []triggers.NotificationTrigger{{
			Name:      "foo",
			Template:  "template name",
			Condition: "the condition",
			Enabled:   pointer.BoolPtr(false),
		}},
	}

	merged := cfg.Merge(&Config{
		Triggers: []triggers.NotificationTrigger{{
			Name:      "foo",
			Condition: "new condition",
			Enabled:   pointer.BoolPtr(true),
		}, {
			Name:      "bar",
			Condition: "the condition",
			Template:  "the template",
			Enabled:   pointer.BoolPtr(true),
		}},
	})

	assert.ElementsMatch(t, merged.Triggers, []triggers.NotificationTrigger{{
		Name:      "foo",
		Template:  "template name",
		Condition: "new condition",
		Enabled:   pointer.BoolPtr(true),
	}, {
		Name:      "bar",
		Condition: "the condition",
		Template:  "the template",
		Enabled:   pointer.BoolPtr(true),
	}})
}
