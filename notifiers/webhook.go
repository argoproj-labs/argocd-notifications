package notifiers

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"

	"github.com/argoproj-labs/argocd-notifications/shared/text"
)

type WebhookNotification struct {
	Method string `json:"method"`
	Body   string `json:"body"`
	Path   string `json:"path"`
}

type Header struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type BasicAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type WebhookSettings struct {
	Name      string     `json:"name"`
	URL       string     `json:"url"`
	Headers   []Header   `json:"headers"`
	BasicAuth *BasicAuth `json:"basicAuth"`
}

// WebhookOptions holds list of configured webhooks settings
type WebhookOptions []WebhookSettings

func NewWebhookNotifier(opts WebhookOptions) Notifier {
	return &webhookNotifier{opts: opts}
}

type webhookNotifier struct {
	opts WebhookOptions
}

func findWebhookSettingsByName(settings []WebhookSettings, name string) (*WebhookSettings, error) {
	for _, item := range settings {
		if item.Name == name {
			return &item, nil
		}
	}
	return nil, fmt.Errorf("webhook with name '%s' is not configured", name)
}

func (w webhookNotifier) Send(notification Notification, recipient string) error {
	webhookSettings, err := findWebhookSettingsByName(w.opts, recipient)
	if err != nil {
		return err
	}
	body := notification.Body
	method := http.MethodGet
	urlPath := ""
	if webhookNotification, ok := notification.Webhook[webhookSettings.Name]; ok {
		body = webhookNotification.Body
		method = text.Coalesce(webhookNotification.Method, method)
		if webhookNotification.Path != "" {
			urlPath = webhookNotification.Path
		}
	}
	url := strings.TrimRight(webhookSettings.URL, "/") + "/" + strings.TrimLeft(urlPath, "/")
	req, err := http.NewRequest(method, url, bytes.NewBufferString(body))
	if err != nil {
		return err
	}
	for _, h := range webhookSettings.Headers {
		req.Header.Set(h.Name, h.Value)
	}
	if webhookSettings.BasicAuth != nil {
		req.SetBasicAuth(webhookSettings.BasicAuth.Username, webhookSettings.BasicAuth.Password)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if !(resp.StatusCode >= 200 && resp.StatusCode <= 299) {
		return fmt.Errorf("request to %s has faild with error code %d", webhookSettings.URL, resp.StatusCode)
	}
	return nil
}
