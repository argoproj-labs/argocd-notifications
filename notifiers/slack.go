package notifiers

import (
	"gomodules.xyz/notify/slack"
)

type SlackOptions struct {
	Token    string   `json:"token"`
	Channels []string `json:"channels"`
}

type slackNotifier struct {
	opts SlackOptions
}

func NewSlackNotifier(opts SlackOptions) Notifier {
	return &slackNotifier{opts: opts}
}

func (n *slackNotifier) Send(_ string, body string, recipient string) error {
	return slack.New(slack.Options{
		AuthToken: n.opts.Token,
		Channel:   n.opts.Channels,
	}).WithBody(body).To(recipient).Send()
}
