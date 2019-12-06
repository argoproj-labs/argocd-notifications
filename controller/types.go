package controller

import "context"

type Config struct {
	Triggers  []NotificationTrigger  `json:"triggers"`
	Templates []NotificationTemplate `json:"templates"`
}

type NotificationTrigger struct {
	Name      string `json:"name"`
	Condition string `json:"condition"`
	Template  string `json:"template"`
}

type NotificationRecipients struct {
	Slack []string `json:"slack"`
}

type NotificationTemplate struct {
	Name  string `json:"name"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

type NotificationController interface {
	Run(ctx context.Context, processors int) error
}
