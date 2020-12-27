package slack

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

var noopVerifier = func(data []byte, header http.Header) (string, error) {
	return "slack", nil
}

func TestParse_ListSubscriptionsCommand(t *testing.T) {
	s := NewSlackAdapter(noopVerifier)

	cmd, err := s.Parse(httptest.NewRequest("GET", "http://localhost/slack",
		bytes.NewBufferString("text=list-subscriptions&channel_name=test")))
	assert.NoError(t, err)

	assert.NotNil(t, cmd.ListSubscriptions)
	assert.Equal(t, cmd.Recipient, "test")
}

func TestParse_SubscribeAppTrigger(t *testing.T) {
	s := NewSlackAdapter(noopVerifier)

	cmd, err := s.Parse(httptest.NewRequest("GET", "http://localhost/slack",
		bytes.NewBufferString("text=subscribe%20foo%20on-sync-failed&channel_name=test")))
	assert.NoError(t, err)

	assert.NotNil(t, cmd.Subscribe)
	assert.Equal(t, cmd.Subscribe.Trigger, "on-sync-failed")
	assert.Equal(t, cmd.Subscribe.App, "foo")
	assert.Equal(t, cmd.Subscribe.Project, "")
	assert.Equal(t, cmd.Recipient, "test")
}

func TestParse_SubscribeProject(t *testing.T) {
	s := NewSlackAdapter(noopVerifier)

	cmd, err := s.Parse(httptest.NewRequest("GET", "http://localhost/slack",
		bytes.NewBufferString("text=subscribe%20proj%3Afoo&channel_name=test")))
	assert.NoError(t, err)

	assert.NotNil(t, cmd.Subscribe)
	assert.Equal(t, cmd.Subscribe.Trigger, "")
	assert.Equal(t, cmd.Subscribe.App, "")
	assert.Equal(t, cmd.Subscribe.Project, "foo")
	assert.Equal(t, cmd.Recipient, "test")
}

func TestParse_UnsubscribeApp(t *testing.T) {
	s := NewSlackAdapter(noopVerifier)

	cmd, err := s.Parse(httptest.NewRequest("GET", "http://localhost/slack",
		bytes.NewBufferString("text=unsubscribe%20app%3Afoo&channel_name=test")))
	assert.NoError(t, err)

	assert.NotNil(t, cmd.Unsubscribe)
	assert.Equal(t, cmd.Unsubscribe.Trigger, "")
	assert.Equal(t, cmd.Unsubscribe.App, "foo")
	assert.Equal(t, cmd.Unsubscribe.Project, "")
	assert.Equal(t, cmd.Recipient, "test")
}

func TestParse_WrongCommandHelpResponse(t *testing.T) {
	s := NewSlackAdapter(noopVerifier)

	_, err := s.Parse(httptest.NewRequest("GET", "http://localhost/slack",
		bytes.NewBufferString("text=wrong&channel_name=test")))
	assert.Error(t, err)

	assert.Contains(t, err.Error(), "Need some help")
}

func TestParse_NoCommandHelpResponse(t *testing.T) {
	s := NewSlackAdapter(noopVerifier)

	_, err := s.Parse(httptest.NewRequest("GET", "http://localhost/slack",
		bytes.NewBufferString("channel_name=test")))
	assert.Error(t, err)

	assert.Contains(t, err.Error(), "Need some help")
}

func TestParse_NoAppArgument(t *testing.T) {
	s := NewSlackAdapter(noopVerifier)

	_, err := s.Parse(httptest.NewRequest("GET", "http://localhost/slack",
		bytes.NewBufferString("text=unsubscribe&channel_name=test")))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one argument expected")
}

func TestSendResponse(t *testing.T) {
	s := NewSlackAdapter(noopVerifier)
	w := httptest.NewRecorder()

	s.SendResponse("test", w)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, `{"blocks":[{"type":"section","text":{"type":"mrkdwn","text":"test"}}]}`, string(body))
}
