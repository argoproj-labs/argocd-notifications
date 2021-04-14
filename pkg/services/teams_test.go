package services

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

func TestGetTemplater_Teams(t *testing.T) {
	notificationTemplate := Notification{
		Teams: &TeamsNotification{
			Template:        "template {{.value}}",
			Title:           "title {{.value}}",
			Text:            "text {{.value}}",
			Facts:           "facts {{.value}}",
			Sections:        "sections {{.value}}",
			PotentialAction: "actions {{.value}}",
			ThemeColor:      "theme color {{.value}}",
		},
	}

	templater, err := notificationTemplate.GetTemplater("test", template.FuncMap{})

	if err != nil {
		t.Error(err)
		return
	}

	notification := Notification{}

	err = templater(&notification, map[string]interface{}{
		"value": "value",
	})

	if err != nil {
		t.Error(err)
		return
	}

	assert.Equal(t, notification.Teams.Template, "template value")
	assert.Equal(t, notification.Teams.Title, "title value")
	assert.Equal(t, notification.Teams.Text, "text value")
	assert.Equal(t, notification.Teams.Sections, "sections value")
	assert.Equal(t, notification.Teams.Facts, "facts value")
	assert.Equal(t, notification.Teams.PotentialAction, "actions value")
	assert.Equal(t, notification.Teams.ThemeColor, "theme color value")
}

func TestTeams_DefaultMessage(t *testing.T) {
	var receivedBody teamsMessage
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		data, err := ioutil.ReadAll(request.Body)
		assert.NoError(t, err)

		err = json.Unmarshal(data, &receivedBody)
		assert.NoError(t, err)

		_, err = writer.Write([]byte("1"))
		assert.NoError(t, err)
	}))
	defer server.Close()

	service := NewTeamsService(TeamsOptions{
		RecipientUrls: map[string]string{
			"test": server.URL,
		},
	})

	notification := Notification{
		Message: "simple message",
	}

	err := service.Send(notification,
		Destination{
			Recipient: "test",
			Service:   "test",
		},
	)

	assert.NoError(t, err)

	assert.Equal(t, receivedBody.Text, notification.Message)
}

func TestTeams_TemplateMessage(t *testing.T) {
	var receivedBody string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		data, err := ioutil.ReadAll(request.Body)
		assert.NoError(t, err)

		receivedBody = string(data)

		_, err = writer.Write([]byte("1"))
		assert.NoError(t, err)
	}))
	defer server.Close()

	service := NewTeamsService(TeamsOptions{
		RecipientUrls: map[string]string{
			"test": server.URL,
		},
	})

	notification := Notification{
		Teams: &TeamsNotification{
			Template: "template body",
		},
	}

	err := service.Send(notification,
		Destination{
			Recipient: "test",
			Service:   "test",
		},
	)

	assert.NoError(t, err)

	assert.Equal(t, receivedBody, notification.Teams.Template)
}

func TestTeams_MessageFields(t *testing.T) {
	var receivedBody teamsMessage
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		data, err := ioutil.ReadAll(request.Body)
		assert.NoError(t, err)

		err = json.Unmarshal(data, &receivedBody)
		assert.NoError(t, err)

		_, err = writer.Write([]byte("1"))
		assert.NoError(t, err)
	}))
	defer server.Close()

	service := NewTeamsService(TeamsOptions{
		RecipientUrls: map[string]string{
			"test": server.URL,
		},
	})

	notification := Notification{
		Message: "welcome message",
		Teams: &TeamsNotification{
			Text:            "Text",
			Facts:           "[{\"facts\": true}]",
			Sections:        "[{\"sections\": true}]",
			PotentialAction: "[{\"actions\": true}]",
			Title:           "Title",
			ThemeColor:      "#000080",
		},
	}

	err := service.Send(notification,
		Destination{
			Recipient: "test",
			Service:   "test",
		},
	)

	assert.NoError(t, err)

	assert.Contains(t, receivedBody.Text, notification.Teams.Text)
	assert.Contains(t, receivedBody.Title, notification.Teams.Title)
	assert.Contains(t, receivedBody.ThemeColor, notification.Teams.ThemeColor)
	assert.Contains(t, receivedBody.PotentialAction, teamsAction{"actions": true})
	assert.Contains(t, receivedBody.Sections, teamsSection{"sections": true})
	assert.EqualValues(t, receivedBody.Sections[len(receivedBody.Sections)-1]["facts"],
		[]interface{}{
			map[string]interface{}{
				"facts": true,
			},
		})
}
