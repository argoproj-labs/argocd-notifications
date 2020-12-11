package notifiers

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/argoproj-labs/argocd-notifications/shared/text"
	log "github.com/sirupsen/logrus"

	httputil "github.com/argoproj-labs/argocd-notifications/shared/http"
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
	url := webhookSettings.URL
	if urlPath != "" {
		url = strings.TrimRight(webhookSettings.URL, "/") + "/" + strings.TrimLeft(urlPath, "/")
	}
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

	client := http.Client{
		Transport: httputil.NewLoggingRoundTripper(
			httputil.NewTransport(url, false), log.WithField("notifier", fmt.Sprintf("webhook:%s", webhookSettings.Name))),
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if !(resp.StatusCode >= 200 && resp.StatusCode <= 299) {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			data = []byte(fmt.Sprintf("unable to read response data: %v", err))
		}
		return fmt.Errorf("request to %s has failed with error code %d : %s", url, resp.StatusCode, string(data))
	}
	return nil
}
