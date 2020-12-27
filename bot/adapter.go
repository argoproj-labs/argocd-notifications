package bot

import "net/http"

type ListSubscriptions struct {
}

type UpdateSubscription struct {
	App     string
	Project string
	Trigger string
}

type Command struct {
	Service           string
	Recipient         string
	ListSubscriptions *ListSubscriptions
	Subscribe         *UpdateSubscription
	Unsubscribe       *UpdateSubscription
}

// Adapter encapsulates integration with the notification service
type Adapter interface {
	// Parses requested command
	Parse(r *http.Request) (Command, error)
	// Sends formatted response
	SendResponse(content string, w http.ResponseWriter)
}
