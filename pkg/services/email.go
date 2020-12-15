package services

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

type emailService struct {
	opts EmailOptions
}

func NewEmailService(opts EmailOptions) NotificationService {
	return &emailService{opts: opts}
}

func (s *emailService) Send(notification Notification, recipient string) error {
	return smtp.New(smtp.Options{
		From:               s.opts.From,
		Host:               s.opts.Host,
		Port:               s.opts.Port,
		InsecureSkipVerify: s.opts.InsecureSkipVerify,
		Password:           s.opts.Password,
		Username:           s.opts.Username,
	}).WithSubject(notification.Title).WithBody(notification.Body).To(recipient).Send()
}
