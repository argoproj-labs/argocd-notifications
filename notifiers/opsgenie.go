package notifiers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/opsgenie/opsgenie-go-sdk-v2/alert"
	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
	log "github.com/sirupsen/logrus"

	httputil "github.com/argoproj-labs/argocd-notifications/shared/http"
)

type OpsgenieOptions struct {
	ApiUrl  string            `json:"apiUrl"`
	ApiKeys map[string]string `json:"apiKeys"`
}

type opsgenieNotifier struct {
	opts OpsgenieOptions
}

func NewOpsgenieNotifier(opts OpsgenieOptions) Notifier {
	return &opsgenieNotifier{opts: opts}
}

func (n *opsgenieNotifier) Send(notification Notification, recipient string) error {
	apiKey, ok := n.opts.ApiKeys[recipient]
	if !ok {
		return fmt.Errorf("no API key configured for recipient %s", recipient)
	}
	alertClient, _ := alert.NewClient(&client.Config{
		ApiKey:         apiKey,
		OpsGenieAPIURL: client.ApiUrl(n.opts.ApiUrl),
		HttpClient: &http.Client{
			Transport: httputil.NewLoggingRoundTripper(
				httputil.NewTransport(n.opts.ApiUrl, false), log.WithField("notifier", "opsgenie")),
		},
	})
	_, err := alertClient.Create(context.TODO(), &alert.CreateAlertRequest{
		Message:     notification.Title,
		Description: notification.Body,
		Responders: []alert.Responder{
			{
				Type: "team",
				Id:   recipient,
			},
		},
		Source: "Argo CD",
	})
	return err
}
