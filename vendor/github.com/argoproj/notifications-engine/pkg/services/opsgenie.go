package services

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	texttemplate "text/template"

	"github.com/opsgenie/opsgenie-go-sdk-v2/alert"
	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
	log "github.com/sirupsen/logrus"

	httputil "github.com/argoproj/notifications-engine/pkg/util/http"
)

type OpsgenieOptions struct {
	ApiUrl  string            `json:"apiUrl"`
	ApiKeys map[string]string `json:"apiKeys"`
}

type OpsgenieNotification struct {
	Description string `json:"description"`
}

func (n *OpsgenieNotification) GetTemplater(name string, f texttemplate.FuncMap) (Templater, error) {
	desc, err := texttemplate.New(name).Funcs(f).Parse(n.Description)
	if err != nil {
		return nil, err
	}
	return func(notification *Notification, vars map[string]interface{}) error {
		if notification.Opsgenie == nil {
			notification.Opsgenie = &OpsgenieNotification{}
		}
		var descData bytes.Buffer
		if err := desc.Execute(&descData, vars); err != nil {
			return err
		}
		notification.Opsgenie.Description = descData.String()
		return nil
	}, nil
}

type opsgenieService struct {
	opts OpsgenieOptions
}

func NewOpsgenieService(opts OpsgenieOptions) NotificationService {
	return &opsgenieService{opts: opts}
}

func (s *opsgenieService) Send(notification Notification, dest Destination) error {
	apiKey, ok := s.opts.ApiKeys[dest.Recipient]
	if !ok {
		return fmt.Errorf("no API key configured for recipient %s", dest.Recipient)
	}
	alertClient, _ := alert.NewClient(&client.Config{
		ApiKey:         apiKey,
		OpsGenieAPIURL: client.ApiUrl(s.opts.ApiUrl),
		HttpClient: &http.Client{
			Transport: httputil.NewLoggingRoundTripper(
				httputil.NewTransport(s.opts.ApiUrl, false), log.WithField("service", "opsgenie")),
		},
	})
	description := ""
	if notification.Opsgenie != nil {
		description = notification.Opsgenie.Description
	}

	_, err := alertClient.Create(context.TODO(), &alert.CreateAlertRequest{
		Message:     notification.Message,
		Description: description,
		Responders: []alert.Responder{
			{
				Type: "team",
				Id:   dest.Recipient,
			},
		},
		Source: "Argo CD",
	})
	return err
}
