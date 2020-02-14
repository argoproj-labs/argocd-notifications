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
	Recipient         string
	ListSubscriptions *ListSubscriptions
	Subscribe         *UpdateSubscription
	Unsubscribe       *UpdateSubscription
}

type Adapter interface {
	Parse(r *http.Request) (Command, error)
	SendResponse(content string, w http.ResponseWriter)
}
