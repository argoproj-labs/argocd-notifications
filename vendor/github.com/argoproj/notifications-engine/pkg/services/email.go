package services

import (
	"bytes"
	texttemplate "text/template"

	"gomodules.xyz/notify/smtp"

	"github.com/argoproj/notifications-engine/pkg/util/text"
)

type EmailNotification struct {
	Subject string `json:"subject,omitempty"`
	Body    string `json:"body,omitempty"`
}

func (n *EmailNotification) GetTemplater(name string, f texttemplate.FuncMap) (Templater, error) {
	subject, err := texttemplate.New(name).Funcs(f).Parse(n.Subject)
	if err != nil {
		return nil, err
	}
	body, err := texttemplate.New(name).Funcs(f).Parse(n.Body)
	if err != nil {
		return nil, err
	}

	return func(notification *Notification, vars map[string]interface{}) error {
		if notification.Email == nil {
			notification.Email = &EmailNotification{}
		}
		var emailSubjectData bytes.Buffer
		if err := subject.Execute(&emailSubjectData, vars); err != nil {
			return err
		}

		if val := emailSubjectData.String(); val != "" {
			notification.Email.Subject = val
		}

		var emailBodyData bytes.Buffer
		if err := body.Execute(&emailBodyData, vars); err != nil {
			return err
		}
		if val := emailBodyData.String(); val != "" {
			notification.Email.Body = val
		}

		return nil
	}, nil
}

type EmailOptions struct {
	Host               string `json:"host"`
	Port               int    `json:"port"`
	InsecureSkipVerify bool   `json:"insecure_skip_verify"`
	Username           string `json:"username"`
	Password           string `json:"password"`
	From               string `json:"from"`
	Html               bool   `json:"html"`
}

type emailService struct {
	opts EmailOptions
}

func NewEmailService(opts EmailOptions) NotificationService {
	return &emailService{opts: opts}
}

func (s *emailService) Send(notification Notification, dest Destination) error {
	subject := ""
	body := notification.Message
	if notification.Email != nil {
		subject = notification.Email.Subject
		body = text.Coalesce(notification.Email.Body, body)
	}
	email := smtp.New(smtp.Options{
		From:               s.opts.From,
		Host:               s.opts.Host,
		Port:               s.opts.Port,
		InsecureSkipVerify: s.opts.InsecureSkipVerify,
		Password:           s.opts.Password,
		Username:           s.opts.Username,
	}).WithSubject(subject).WithBody(body).To(dest.Recipient)

	if s.opts.Html {
		return email.SendHtml()
	} else {
		return email.Send()
	}
}
