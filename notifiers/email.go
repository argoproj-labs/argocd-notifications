package notifiers

import (
	"gomodules.xyz/notify/smtp"
)

type emailNotifier struct {
	opts smtp.Options
}

func NewEmailNotifier(opts smtp.Options) Notifier {
	return &emailNotifier{opts: opts}
}

func (n *emailNotifier) Send(title string, body string, recipient string) error {
	return smtp.New(n.opts).WithSubject(title).WithBody(body).To(recipient).Send()
}
