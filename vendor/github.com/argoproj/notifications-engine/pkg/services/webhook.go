package services

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	texttemplate "text/template"

	log "github.com/sirupsen/logrus"

	httputil "github.com/argoproj/notifications-engine/pkg/util/http"
	"github.com/argoproj/notifications-engine/pkg/util/text"
)

type WebhookNotification struct {
	Method string `json:"method"`
	Body   string `json:"body"`
	Path   string `json:"path"`
}

type WebhookNotifications map[string]WebhookNotification

type compiledWebhookTemplate struct {
	body   *texttemplate.Template
	path   *texttemplate.Template
	method string
}

func (n WebhookNotifications) GetTemplater(name string, f texttemplate.FuncMap) (Templater, error) {
	webhooks := map[string]compiledWebhookTemplate{}
	for k, v := range n {
		body, err := texttemplate.New(name + k).Funcs(f).Parse(v.Body)
		if err != nil {
			return nil, err
		}
		path, err := texttemplate.New(name + k).Funcs(f).Parse(v.Path)
		if err != nil {
			return nil, err
		}
		webhooks[k] = compiledWebhookTemplate{body: body, method: v.Method, path: path}
	}
	return func(notification *Notification, vars map[string]interface{}) error {
		for k, v := range webhooks {
			if notification.Webhook == nil {
				notification.Webhook = map[string]WebhookNotification{}
			}
			var body bytes.Buffer
			err := webhooks[k].body.Execute(&body, vars)
			if err != nil {
				return err
			}
			var path bytes.Buffer
			err = webhooks[k].path.Execute(&path, vars)
			if err != nil {
				return err
			}
			notification.Webhook[k] = WebhookNotification{
				Method: v.method,
				Body:   body.String(),
				Path:   path.String(),
			}
		}
		return nil
	}, nil
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
	request := request{
		body:        notification.Message,
		method:      http.MethodGet,
		url:         s.opts.URL,
		destService: dest.Service,
	}

	if webhookNotification, ok := notification.Webhook[dest.Service]; ok {
		request.applyOverridesFrom(webhookNotification)
	}

	resp, err := request.execute(&s)
	if err != nil {
		return err
	}

	if !(resp.StatusCode >= 200 && resp.StatusCode <= 299) {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			data = []byte(fmt.Sprintf("unable to read response data: %v", err))
		}
		return fmt.Errorf("request to %s has failed with error code %d : %s", request, resp.StatusCode, string(data))
	}
	return nil
}

type request struct {
	body        string
	method      string
	url         string
	destService string
}

func (r *request) applyOverridesFrom(notification WebhookNotification) {
	r.body = notification.Body
	r.method = text.Coalesce(notification.Method, r.method)
	if notification.Path != "" {
		r.url = strings.TrimRight(r.url, "/") + "/" + strings.TrimLeft(notification.Path, "/")
	}
}

func (r *request) intoHttpRequest(service *webhookService) (*http.Request, error) {
	req, err := http.NewRequest(r.method, r.url, bytes.NewBufferString(r.body))
	if err != nil {
		return nil, err
	}
	for _, header := range service.opts.Headers {
		req.Header.Set(header.Name, header.Value)
	}
	if service.opts.BasicAuth != nil {
		req.SetBasicAuth(service.opts.BasicAuth.Username, service.opts.BasicAuth.Password)
	}
	return req, nil
}

func (r *request) execute(service *webhookService) (*http.Response, error) {
	req, err := r.intoHttpRequest(service)
	if err != nil {
		return nil, err
	}

	client := http.Client{
		Transport: httputil.NewLoggingRoundTripper(
			httputil.NewTransport(r.url, false),
			log.WithField("service", r.destService)),
	}

	return client.Do(req)
}
