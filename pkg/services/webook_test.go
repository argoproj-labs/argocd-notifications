package services

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWebhook_SuccessfullySendsNotification(t *testing.T) {
	var receivedHeaders http.Header
	var receivedBody string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		receivedHeaders = request.Header
		data, err := ioutil.ReadAll(request.Body)
		assert.NoError(t, err)
		receivedBody = string(data)
	}))
	defer server.Close()

	service := NewWebhookService(WebhookOptions{{
		Name:      "test",
		BasicAuth: &BasicAuth{Username: "testUsername", Password: "testPassword"},
		URL:       server.URL,
		Headers:   []Header{{Name: "testHeader", Value: "testHeaderValue"}},
	}})
	err := service.Send(
		Notification{
			Webhook: map[string]WebhookNotification{
				"test": {Body: "hello world", Method: http.MethodPost},
			},
		}, "test")
	assert.NoError(t, err)

	assert.Equal(t, "hello world", receivedBody)
	assert.Equal(t, receivedHeaders.Get("testHeader"), "testHeaderValue")
	assert.Contains(t, receivedHeaders.Get("Authorization"), "Basic")
}

func TestWebhook_FailedToSendNotConfigured(t *testing.T) {
	service := NewWebhookService(WebhookOptions{})
	err := service.Send(Notification{}, "test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not configured")
}

func TestWebhook_SubPath(t *testing.T) {
	var receivedPath string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		receivedPath = request.URL.Path
	}))
	defer server.Close()

	service := NewWebhookService(WebhookOptions{{
		Name: "test",
		URL:  fmt.Sprintf("%s/subpath1", server.URL),
	}})

	err := service.Send(Notification{
		Webhook: map[string]WebhookNotification{
			"test": {Body: "hello world", Method: http.MethodPost},
		},
	}, "test")
	assert.NoError(t, err)
	assert.Equal(t, "/subpath1", receivedPath)

	err = service.Send(Notification{
		Webhook: map[string]WebhookNotification{
			"test": {Body: "hello world", Method: http.MethodPost, Path: "/subpath2"},
		},
	}, "test")
	assert.NoError(t, err)
	assert.Equal(t, "/subpath1/subpath2", receivedPath)
}
