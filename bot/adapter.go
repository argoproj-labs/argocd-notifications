package bot

import "net/http"

type ListSubscriptions struct {
}

type Command struct {
	Subscriber string
	*ListSubscriptions
}

type Adapter interface {
	Parse(r *http.Request) (Command, error)
	SendResponse(content string, w http.ResponseWriter)
}
