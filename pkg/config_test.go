package pkg

import (
	"testing"

	"github.com/argoproj-labs/argocd-notifications/pkg/services"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

var (
	emptySecret = &v1.Secret{Data: map[string][]byte{}}
)

func TestParseConfig_Services(t *testing.T) {
	cfg, err := ParseConfig(&v1.ConfigMap{Data: map[string]string{
		"service.slack": `
token: my-token
`}}, emptySecret)

	if !assert.NoError(t, err) {
		return
	}

	assert.NotNil(t, cfg.Services["slack"])
}

func TestParseConfig_Templates(t *testing.T) {
	cfg, err := ParseConfig(&v1.ConfigMap{Data: map[string]string{
		"template.my-template": `
message: hello world
`}}, emptySecret)

	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, map[string]services.Notification{
		"my-template": {Message: "hello world"},
	}, cfg.Templates)
}

func TestReplaceStringSecret_KeyPresent(t *testing.T) {
	val := replaceStringSecret("hello $secret-value", map[string][]byte{
		"secret-value": []byte("world"),
	})

	assert.Equal(t, "hello world", val)
}

func TestReplaceStringSecret_KeyMissing(t *testing.T) {
	val := replaceStringSecret("hello $secret-value", map[string][]byte{
		"another-secret-value": []byte("world"),
	})

	assert.Equal(t, "hello $secret-value", val)
}
