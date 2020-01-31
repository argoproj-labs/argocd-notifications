package notifiers

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"

	"github.com/nlopes/slack"
)

type SlackOptions struct {
	Username           string   `json:"username"`
	Token              string   `json:"token"`
	Channels           []string `json:"channels"`
	InsecureSkipVerify bool     `json:"insecureSkipVerify"`
}

type slackNotifier struct {
	opts SlackOptions
}

func NewSlackNotifier(opts SlackOptions) Notifier {
	return &slackNotifier{opts: opts}
}

func (n *slackNotifier) Send(notification Notification, recipient string) error {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: n.opts.InsecureSkipVerify,
			},
		},
	}
	s := slack.New(n.opts.Token, slack.OptionHTTPClient(client))
	msgOptions := []slack.MsgOption{slack.MsgOptionText(notification.Body, false), slack.MsgOptionUsername(n.opts.Username)}

	if notification.Slack != nil {
		attachments := make([]slack.Attachment, 0)
		if notification.Slack.Attachments != "" {
			if err := json.Unmarshal([]byte(notification.Slack.Attachments), &attachments); err != nil {
				return err
			}
		}

		blocks := slack.Blocks{}
		if notification.Slack.Blocks != "" {
			if err := json.Unmarshal([]byte(notification.Slack.Blocks), &blocks); err != nil {
				return err
			}
		}
		msgOptions = append(msgOptions, slack.MsgOptionAttachments(attachments...), slack.MsgOptionBlocks(blocks.BlockSet...))
	}

	_, _, err := s.PostMessageContext(context.TODO(), recipient, msgOptions...)
	return err
}
