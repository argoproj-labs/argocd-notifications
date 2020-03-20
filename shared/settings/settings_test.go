package settings

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
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
    <team-id>: <my-api-key>
grafana:
  apiUrl: grafana.com/api
  apiKey: <my-api-key>`)

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
		Grafana: &notifiers.GrafanaOptions{
			ApiUrl: "grafana.com/api",
			ApiKey: "<my-api-key>",
		},
	}
	actualNotifiersCfg, err := ParseSecret(&v1.Secret{Data: map[string][]byte{"notifiers.yaml": notifiersData}})
	assert.NoError(t, err)
	assert.Equal(t, expectNotifiersCfg, actualNotifiersCfg)
}

func TestParseConfigMap(t *testing.T) {
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
				Slack: &notifiers.SlackNotification{
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
				Slack: &notifiers.SlackNotification{
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
			Slack: &notifiers.SlackNotification{
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

func TestDefaultSubscriptions_GetRecipients(t *testing.T) {
	selector, err := labels.Parse("test=true")
	assert.NoError(t, err)

	subscriptions := DefaultSubscriptions([]Subscription{{
		Recipients: []string{"slack:test1", "slack:test2"},
		Selector:   labels.NewSelector(),
	}, {
		Recipients: []string{"slack:test3"},
		Triggers:   []string{"trigger2"},
		Selector:   labels.NewSelector(),
	}, {
		Recipients: []string{"slack:test4"},
		Selector:   selector,
	}})

	assert.ElementsMatch(t, []string{"slack:test1", "slack:test2"}, subscriptions.GetRecipients("trigger1", map[string]string{}))
	assert.ElementsMatch(t, []string{"slack:test1", "slack:test2", "slack:test3"}, subscriptions.GetRecipients("trigger2", map[string]string{}))
	assert.ElementsMatch(t, []string{"slack:test1", "slack:test2", "slack:test4"}, subscriptions.GetRecipients("trigger3", map[string]string{"test": "true"}))
}
