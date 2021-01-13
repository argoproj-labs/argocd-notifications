package services

import (
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

func TestGetTemplater(t *testing.T) {
	n := Notification{Message: "{{.foo}}"}

	templater, err := n.GetTemplater("", template.FuncMap{})
	if !assert.NoError(t, err) {
		return
	}

	var notification Notification

	err = templater(&notification, map[string]interface{}{
		"foo": "hello",
	})

	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "hello", notification.Message)
}
