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
	channel := query.Get("channel_name")
	if channel == "" {
		return cmd, errors.New("request does not have channel")
	}
	parts := strings.Fields(query.Get("text"))
	if len(parts) < 1 {
		return cmd, errors.New("request does not have command")
	}
	command := parts[0]

	cmd.Recipient = fmt.Sprintf("slack:%s", channel)

	switch command {
	case "list-subscriptions":
		cmd.ListSubscriptions = &ListSubscriptions{}
	case "subscribe", "unsubscribe":
		if len(parts) < 2 {
			return cmd, fmt.Errorf("command %s expects at least one argument", command)
		}
		update := &UpdateSubscription{}
		nameParts := strings.Split(parts[1], ":")
		if len(nameParts) == 1 {
			nameParts = append([]string{"app"}, nameParts...)
		}
		switch nameParts[0] {
		case "app":
			update.App = nameParts[1]
		case "proj":
			update.Project = nameParts[1]
		default:
			return cmd, fmt.Errorf("incorrect name argument: %s", parts[1])
		}
		if len(parts) > 2 {
			update.Trigger = parts[2]
		}
		if command == "subscribe" {
			cmd.Subscribe = update
		} else {
			cmd.Unsubscribe = update
		}
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
