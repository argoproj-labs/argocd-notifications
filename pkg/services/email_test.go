package services

import (
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

func TestGetTemplater_Email(t *testing.T) {
	n := Notification{
		Email: &EmailNotification{
			Subject: "{{.foo}}", Body: "{{.bar}}",
		},
	}

	templater, err := n.GetTemplater("", template.FuncMap{})
	if !assert.NoError(t, err) {
		return
	}

	var notification Notification

	err = templater(&notification, map[string]interface{}{
		"foo": "hello",
		"bar": "world",
	})

	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "hello", notification.Email.Subject)
	assert.Equal(t, "world", notification.Email.Body)
}
