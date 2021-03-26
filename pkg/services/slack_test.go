package services

import (
	"os"
	"testing"
	"text/template"
	"time"

	"github.com/stretchr/testify/assert"

	et "github.com/argoproj-labs/argocd-notifications/expr/time"
)

func TestValidIconEmoij(t *testing.T) {
	assert.Equal(t, true, validIconEmoij.MatchString(":slack:"))
	assert.Equal(t, true, validIconEmoij.MatchString(":chart_with_upwards_trend:"))
	assert.Equal(t, false, validIconEmoij.MatchString("http://lorempixel.com/48/48"))
}

func TestValidIconURL(t *testing.T) {
	assert.Equal(t, true, isValidIconURL("http://lorempixel.com/48/48"))
	assert.Equal(t, true, isValidIconURL("https://lorempixel.com/48/48"))
	assert.Equal(t, false, isValidIconURL("favicon.ico"))
	assert.Equal(t, false, isValidIconURL("ftp://favicon.ico"))
	assert.Equal(t, false, isValidIconURL("ftp://lorempixel.com/favicon.ico"))
}

func TestGetTemplater_Slack(t *testing.T) {
	n := Notification{
		Slack: &SlackNotification{
			Attachments: "{{.foo}}",
			Blocks:      "{{.bar}}",
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

	assert.Equal(t, "hello", notification.Slack.Attachments)
	assert.Equal(t, "world", notification.Slack.Blocks)
}

func TestChangeTimezone_Slack(t *testing.T) {
	tz := os.Getenv("TZ")
	if err := os.Setenv("TZ", "Asia/Tokyo"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Setenv("TZ", tz); err != nil {
			t.Fatal(err)
		}
	})

	n := Notification{
		Slack: &SlackNotification{
			Attachments: `{{ (call .time.Parse .date).Local.Format "2006-01-02T15:04:05Z07:00" }}`,
		},
	}
	templater, err := n.GetTemplater("", template.FuncMap{})

	if !assert.NoError(t, err) {
		return
	}

	currentTime, err := time.Parse(time.RFC3339, "2021-03-26T23:38:01Z")
	if !assert.NoError(t, err) {
		return
	}

	var notification Notification
	err = templater(&notification, map[string]interface{}{
		"date": currentTime.Format(time.RFC3339),
		"time": et.NewExprs(),
	})

	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "2021-03-27T08:38:01+09:00", notification.Slack.Attachments)
}
