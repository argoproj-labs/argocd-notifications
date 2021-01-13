package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	texttemplate "text/template"

	httputil "github.com/argoproj-labs/argocd-notifications/pkg/util/http"

	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
)

type SlackNotification struct {
	Attachments string `json:"attachments,omitempty"`
	Blocks      string `json:"blocks,omitempty"`
}

func (n *SlackNotification) GetTemplater(name string, f texttemplate.FuncMap) (Templater, error) {
	slackAttachments, err := texttemplate.New(name).Funcs(f).Parse(n.Attachments)
	if err != nil {
		return nil, err
	}
	slackBlocks, err := texttemplate.New(name).Funcs(f).Parse(n.Blocks)
	if err != nil {
		return nil, err
	}
	return func(notification *Notification, vars map[string]interface{}) error {
		if notification.Slack == nil {
			notification.Slack = &SlackNotification{}
		}
		var slackAttachmentsData bytes.Buffer
		if err := slackAttachments.Execute(&slackAttachmentsData, vars); err != nil {
			return err
		}

		notification.Slack.Attachments = slackAttachmentsData.String()
		var slackBlocksData bytes.Buffer
		if err := slackBlocks.Execute(&slackBlocksData, vars); err != nil {
			return err
		}
		notification.Slack.Blocks = slackBlocksData.String()
		return nil
	}, nil
}

type SlackOptions struct {
	Username           string   `json:"username"`
	Icon               string   `json:"icon"`
	Token              string   `json:"token"`
	SigningSecret      string   `json:"signingSecret"`
	Channels           []string `json:"channels"`
	InsecureSkipVerify bool     `json:"insecureSkipVerify"`
	ApiURL             string   `json:"apiURL"`
}

type slackService struct {
	opts SlackOptions
}

var validIconEmoij = regexp.MustCompile(`^:.+:$`)

func NewSlackService(opts SlackOptions) NotificationService {
	return &slackService{opts: opts}
}

func (s *slackService) Send(notification Notification, dest Destination) error {
	apiURL := slack.APIURL
	if s.opts.ApiURL != "" {
		apiURL = s.opts.ApiURL
	}
	transport := httputil.NewTransport(apiURL, s.opts.InsecureSkipVerify)
	client := &http.Client{
		Transport: httputil.NewLoggingRoundTripper(transport, log.WithField("service", "slack")),
	}
	sl := slack.New(s.opts.Token, slack.OptionHTTPClient(client), slack.OptionAPIURL(apiURL))
	msgOptions := []slack.MsgOption{slack.MsgOptionText(notification.Message, false)}
	if s.opts.Username != "" {
		msgOptions = append(msgOptions, slack.MsgOptionUsername(s.opts.Username))
	}
	if s.opts.Icon != "" {
		if validIconEmoij.MatchString(s.opts.Icon) {
			msgOptions = append(msgOptions, slack.MsgOptionIconEmoji(s.opts.Icon))
		} else if isValidIconURL(s.opts.Icon) {
			msgOptions = append(msgOptions, slack.MsgOptionIconURL(s.opts.Icon))
		} else {
			log.Warnf("Icon reference '%v' is not a valid emoij or url", s.opts.Icon)
		}
	}

	if notification.Slack != nil {
		attachments := make([]slack.Attachment, 0)
		if notification.Slack.Attachments != "" {
			if err := json.Unmarshal([]byte(notification.Slack.Attachments), &attachments); err != nil {
				return fmt.Errorf("failed to unmarshal attachments '%s' : %v", notification.Slack.Attachments, err)
			}
		}

		blocks := slack.Blocks{}
		if notification.Slack.Blocks != "" {
			if err := json.Unmarshal([]byte(notification.Slack.Blocks), &blocks); err != nil {
				return fmt.Errorf("failed to unmarshal blocks '%s' : %v", notification.Slack.Blocks, err)
			}
		}
		msgOptions = append(msgOptions, slack.MsgOptionAttachments(attachments...), slack.MsgOptionBlocks(blocks.BlockSet...))
	}

	_, _, err := sl.PostMessageContext(context.TODO(), dest.Recipient, msgOptions...)
	return err
}

// GetSigningSecret exposes signing secret for slack bot
func (s *slackService) GetSigningSecret() string {
	return s.opts.SigningSecret
}

func isValidIconURL(iconURL string) bool {
	_, err := url.ParseRequestURI(iconURL)
	if err != nil {
		return false
	}

	u, err := url.Parse(iconURL)
	if err != nil || (u.Scheme == "" || !(u.Scheme == "http" || u.Scheme == "https")) || u.Host == "" {
		return false
	}

	return true
}
