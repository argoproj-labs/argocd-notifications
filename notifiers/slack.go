package notifiers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"

	httputil "github.com/argoproj-labs/argocd-notifications/shared/http"

	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
)

type SlackNotification struct {
	Attachments string `json:"attachments,omitempty"`
	Blocks      string `json:"blocks,omitempty"`
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

type slackNotifier struct {
	opts SlackOptions
}

var validIconEmoij = regexp.MustCompile(`^:.+:$`)

func NewSlackNotifier(opts SlackOptions) Notifier {
	return &slackNotifier{opts: opts}
}

func (n *slackNotifier) Send(notification Notification, recipient string) error {
	apiURL := slack.APIURL
	if n.opts.ApiURL != "" {
		apiURL = n.opts.ApiURL
	}
	transport := httputil.NewTransport(apiURL, n.opts.InsecureSkipVerify)
	client := &http.Client{
		Transport: httputil.NewLoggingRoundTripper(transport, log.WithField("notifier", "slack")),
	}
	s := slack.New(n.opts.Token, slack.OptionHTTPClient(client), slack.OptionAPIURL(apiURL))
	msgOptions := []slack.MsgOption{slack.MsgOptionText(notification.Body, false)}
	if n.opts.Username != "" {
		msgOptions = append(msgOptions, slack.MsgOptionUsername(n.opts.Username))
	}
	if n.opts.Icon != "" {
		if validIconEmoij.MatchString(n.opts.Icon) {
			msgOptions = append(msgOptions, slack.MsgOptionIconEmoji(n.opts.Icon))
		} else if isValidIconURL(n.opts.Icon) {
			msgOptions = append(msgOptions, slack.MsgOptionIconURL(n.opts.Icon))
		} else {
			log.Warnf("Icon reference '%v' is not a valid emoij or url", n.opts.Icon)
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

	_, _, err := s.PostMessageContext(context.TODO(), recipient, msgOptions...)
	return err
}

// GetSigningSecret exposes signing secret for slack bot
func (n *slackNotifier) GetSigningSecret() string {
	return n.opts.SigningSecret
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
