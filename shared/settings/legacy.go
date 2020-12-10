package settings

import "github.com/argoproj-labs/argocd-notifications/notifiers"

type legacyConfig struct {
	Email    *notifiers.EmailOptions    `json:"email"`
	Slack    *notifiers.SlackOptions    `json:"slack"`
	Opsgenie *notifiers.OpsgenieOptions `json:"opsgenie"`
	Grafana  *notifiers.GrafanaOptions  `json:"grafana"`
	Webhook  *notifiers.WebhookOptions  `json:"webhook"`
}

func (legacyConf *legacyConfig) addNotifiers(res map[string]notifiers.Notifier) {
	if legacyConf.Email != nil {
		res["email"] = notifiers.NewEmailNotifier(*legacyConf.Email)
	}
	if legacyConf.Slack != nil {
		res["slack"] = notifiers.NewSlackNotifier(*legacyConf.Slack)
	}
	if legacyConf.Grafana != nil {
		res["grafana"] = notifiers.NewGrafanaNotifier(*legacyConf.Grafana)
	}
	if legacyConf.Opsgenie != nil {
		res["opsgenie"] = notifiers.NewOpsgenieNotifier(*legacyConf.Opsgenie)
	}
	if legacyConf.Webhook != nil {
		res["webhook"] = notifiers.NewWebhookNotifier(*legacyConf.Webhook)
	}
}
