package bot

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
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
	cmd.Subscriber = fmt.Sprintf("slack:%s", subscriber)
	cmd.ListSubscriptions = &ListSubscriptions{}
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
