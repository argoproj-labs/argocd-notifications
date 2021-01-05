package services

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/argoproj-labs/argocd-notifications/pkg/util/text"
	log "github.com/sirupsen/logrus"

	httputil "github.com/argoproj-labs/argocd-notifications/pkg/util/http"
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

type WebhookOptions struct {
	URL       string     `json:"url"`
	Headers   []Header   `json:"headers"`
	BasicAuth *BasicAuth `json:"basicAuth"`
}

func NewWebhookService(opts WebhookOptions) NotificationService {
	return &webhookService{opts: opts}
}

type webhookService struct {
	opts WebhookOptions
}

func (s webhookService) Send(notification Notification, dest Destination) error {
	body := notification.Body
	method := http.MethodGet
	urlPath := ""
	if notification.Webhook != nil {
		if webhookNotification, ok := notification.Webhook[dest.Service]; ok {
			body = webhookNotification.Body
			method = text.Coalesce(webhookNotification.Method, method)
			if webhookNotification.Path != "" {
				urlPath = webhookNotification.Path
			}
		}
	}
	url := s.opts.URL
	if urlPath != "" {
		url = strings.TrimRight(s.opts.URL, "/") + "/" + strings.TrimLeft(urlPath, "/")
	}
	req, err := http.NewRequest(method, url, bytes.NewBufferString(body))
	if err != nil {
		return err
	}
	for _, h := range s.opts.Headers {
		req.Header.Set(h.Name, h.Value)
	}
	if s.opts.BasicAuth != nil {
		req.SetBasicAuth(s.opts.BasicAuth.Username, s.opts.BasicAuth.Password)
	}

	client := http.Client{
		Transport: httputil.NewLoggingRoundTripper(
			httputil.NewTransport(url, false), log.WithField("service", dest.Service)),
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
