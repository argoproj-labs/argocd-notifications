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

	secret := &v1.Secret{Data: map[string][]byte{

		"notifier.email": []byte(`
  host: smtp.gmail.com
  port: 587
  from: <myemail>@gmail.com
  username: <myemail>@gmail.com
  password: <mypassword>`),

		"notifier.slack": []byte(`
  token: <my-token>
  username: <override-username>`),

		"notifier.opsgenie": []byte(`
  apiUrl: api.opsgenie.com
  apiKeys:
    <team-id>: <my-api-key>`),

		"notifier.grafana": []byte(`
  apiUrl: grafana.com/api
  apiKey: <my-api-key>`),
	}}

	n, err := ParseSecret(secret)
	assert.NoError(t, err)
	assert.Len(t, n, 4)
}

func TestParseConfigMap(t *testing.T) {
	configData := map[string]string{
		"subscriptions": `
- recipients:
    - slack:test
  triggers:
    - on-sync-status-custom`,

		"context": `
argocdUrl: testUrl`,

		"trigger.on-sync-status-custom": `
name: on-sync-status-custom
condition: app.status.operationState.phase in ['Custom']
description: Application custom trigger
template: app-sync-status
enabled: true`,
		"template.app-sync-status": `
name: app-sync-status
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
    }]`,
	}

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
}]`,
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
	merged, err := cfg.Merge(&Config{
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

	if !assert.NoError(t, err) {
		return
	}

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

	merged, err := cfg.Merge(&Config{
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
	if !assert.NoError(t, err) {
		return
	}

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

func TestMergeSubscriptions(t *testing.T) {
	cfg := Config{
		Subscriptions: []Subscription{{
			Recipients: []string{"foo"},
			Triggers:   []string{"far"},
		}},
	}

	selector, err := labels.Parse("foo=true")
	if !assert.NoError(t, err) {
		return
	}

	merged, err := cfg.Merge(&Config{
		Subscriptions: []Subscription{{
			Recipients: []string{"replacedFoo"},
			Triggers:   []string{"replacedFoo"},
			Selector:   selector,
		}},
	})

	if !assert.NoError(t, err) {
		return
	}

	assert.ElementsMatch(t, merged.Subscriptions, []Subscription{{
		Recipients: []string{"replacedFoo"},
		Triggers:   []string{"replacedFoo"},
		Selector:   selector,
	}})
}

func TestMergeWebhooks(t *testing.T) {
	cfg := Config{
		Templates: []triggers.NotificationTemplate{{
			Name: "foo",
			Notification: notifiers.Notification{
				Body: "hello world",
			},
		}},
	}

	merged, err := cfg.Merge(&Config{
		Templates: []triggers.NotificationTemplate{{
			Name: "foo",
			Notification: notifiers.Notification{
				Webhook: map[string]notifiers.WebhookNotification{
					"slack": {
						Method: "slack method",
						Body:   "slack body",
					},
				},
			},
		}},
	})

	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "hello world", merged.Templates[0].Body)
	assert.Equal(t, "slack body", merged.Templates[0].Webhook["slack"].Body)
	assert.Equal(t, "slack method", merged.Templates[0].Webhook["slack"].Method)
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
