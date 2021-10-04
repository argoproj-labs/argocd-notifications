package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	texttemplate "text/template"

	log "github.com/sirupsen/logrus"

	httputil "github.com/argoproj/notifications-engine/pkg/util/http"
)

type MattermostNotification struct {
	Attachments string `json:"attachments,omitempty"`
}

func (n *MattermostNotification) GetTemplater(name string, f texttemplate.FuncMap) (Templater, error) {
	mattermostAttachments, err := texttemplate.New(name).Funcs(f).Parse(n.Attachments)
	if err != nil {
		return nil, err
	}
	return func(notification *Notification, vars map[string]interface{}) error {
		if notification.Mattermost == nil {
			notification.Mattermost = &MattermostNotification{}
		}
		var mattermostAttachmentsData bytes.Buffer
		if err := mattermostAttachments.Execute(&mattermostAttachmentsData, vars); err != nil {
			return err
		}

		notification.Mattermost.Attachments = mattermostAttachmentsData.String()
		return nil
	}, nil
}

type MattermostOptions struct {
	ApiURL             string `json:"apiURL"`
	Token              string `json:"token"`
	InsecureSkipVerify bool   `json:"insecureSkipVerify"`
}

type mattermostService struct {
	opts MattermostOptions
}

func NewMattermostService(opts MattermostOptions) NotificationService {
	return &mattermostService{opts: opts}
}

func (m *mattermostService) Send(notification Notification, dest Destination) error {
	transport := httputil.NewTransport(m.opts.ApiURL, m.opts.InsecureSkipVerify)
	client := &http.Client{
		Transport: httputil.NewLoggingRoundTripper(transport, log.WithField("service", "mattermost")),
	}

	attachments := []interface{}{}
	if notification.Mattermost != nil {
		if notification.Mattermost.Attachments != "" {
			if err := json.Unmarshal([]byte(notification.Mattermost.Attachments), &attachments); err != nil {
				return fmt.Errorf("failed to unmarshal attachments '%s' : %v", notification.Mattermost.Attachments, err)
			}
		}
	}

	body := map[string]interface{}{
		"channel_id": dest.Recipient,
		"message":    notification.Message,
		"props": map[string]interface{}{
			"attachments": attachments,
		},
	}
	b, _ := json.Marshal(body)

	req, err := http.NewRequest(http.MethodPost, m.opts.ApiURL+"/api/v4/posts", bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.opts.Token))

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to request: %v", err)
	}
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed to read body: %v", err)
	}

	if res.StatusCode/100 != 2 {
		return fmt.Errorf("request to %s has failed with error code %d : %s", body, res.StatusCode, string(data))
	}

	return nil
}
