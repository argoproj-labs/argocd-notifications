package notifiers

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/nlopes/slack"
)

type SlackOptions struct {
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

func (n *slackNotifier) Send(_ string, body string, recipient string) error {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: n.opts.InsecureSkipVerify,
			},
		},
	}
	s := slack.New(n.opts.Token, slack.OptionHTTPClient(client))
	_, _, err := s.PostMessageContext(
		context.TODO(),
		recipient,
		slack.MsgOptionText(body, false))
	return err
}
