package notifiers

import (
	"gomodules.xyz/notify/smtp"
)

type EmailOptions struct {
	Host               string `json:"host"`
	Port               int    `json:"port"`
	InsecureSkipVerify bool   `json:"insecure_skip_verify"`
	Username           string `json:"username"`
	Password           string `json:"password"`
	From               string `json:"from"`
}

type emailNotifier struct {
	opts EmailOptions
}

func NewEmailNotifier(opts EmailOptions) Notifier {
	return &emailNotifier{opts: opts}
}

func (n *emailNotifier) Send(title string, body string, recipient string) error {
	return smtp.New(smtp.Options{
		From:               n.opts.From,
		Host:               n.opts.Host,
		Port:               n.opts.Port,
		InsecureSkipVerify: n.opts.InsecureSkipVerify,
		Password:           n.opts.Password,
		Username:           n.opts.Username,
	}).WithSubject(title).WithBody(body).To(recipient).Send()
}
