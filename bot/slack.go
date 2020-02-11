package bot

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type slack struct {
}

func (s *slack) Parse(r *http.Request) (Command, error) {
	cmd := Command{}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return cmd, err
	}
	query, err := url.ParseQuery(string(data))
	if err != nil {
		return cmd, err
	}
	subscriber := query.Get("user_name")
	if subscriber == "" {
		return cmd, errors.New("request does not have user info")
	}
	channel := query.Get("channel_name")
	if channel == "" {
		return cmd, errors.New("request does not have channel")
	}
	parts := strings.Fields(query.Get("text"))
	if len(channel) < 1 {
		return cmd, errors.New("request does not have command")
	}
	command := parts[0]

	cmd.Subscriber = fmt.Sprintf("slack:%s", subscriber)

	switch command {
	case "list-subscriptions":
		cmd.ListSubscriptions = &ListSubscriptions{Channel: fmt.Sprintf("slack:%s", channel)}
	default:
		return cmd, fmt.Errorf("command %s is not supported", command)
	}
	return cmd, nil
}

func (s *slack) SendResponse(content string, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{
    "blocks": [
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": "%s"
            }
        }
    ]
}`, content)
}
