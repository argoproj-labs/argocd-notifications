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

func TestParseConfig_DefaultTriggers(t *testing.T) {
	cfg, err := ParseConfig(&v1.ConfigMap{Data: map[string]string{
		"defaultTriggers.slack": `
- trigger-a
- trigger-b
`}}, emptySecret)

	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, map[string][]string{
		"slack": {
			"trigger-a",
			"trigger-b",
		},
	}, cfg.ServiceDefaultTriggers)
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

func TestReplaceServiceConfigSecrets_WithBasicWebhook_ReplacesSecrets(t *testing.T) {
	input := `url: $endpoint
headers:
  - name: Authorization
    value: Bearer $secret-value
`

	secrets := v1.Secret{
		Data: map[string][]byte{
			"endpoint":     []byte("https://example.com"),
			"secret-value": []byte("token"),
		},
	}

	expected := `url: https://example.com
headers:
  - name: Authorization
    value: Bearer token
`

	result, err := replaceServiceConfigSecrets(input, &secrets)

	assert.NoError(t, err)
	assert.Equal(t, expected, string(result))
}

func TestReplaceServiceConfigSecrets_WithMapOfSecrets_ReplacesSecrets(t *testing.T) {
	input := `apiUrl: $api-url
apiKeys:
  first-team: $first-team-secret
  second-team: $second-team-secret
`

	secrets := v1.Secret{
		Data: map[string][]byte{
			"first-team-secret":  []byte("first-token"),
			"second-team-secret": []byte("second-token"),
		},
	}

	expected := `apiUrl: $api-url
apiKeys:
    first-team: first-token
    second-team: second-token
`

	result, err := replaceServiceConfigSecrets(input, &secrets)

	assert.NoError(t, err)
	assert.Equal(t, expected, string(result))
}

func TestReplaceServiceConfigSecrets_WithMultilineSecret_ReplacesSecrets(t *testing.T) {
	input := `appID: 12345
privateKey: $github-privateKey
installationID: 67890
`

	secrets := v1.Secret{
		Data: map[string][]byte{
			"github-privateKey": []byte("A\nValue\nOn\nMultiple\nLines"),
		},
	}

	expected := `appID: 12345
privateKey: |-
    A
    Value
    On
    Multiple
    Lines
installationID: 67890
`

	result, err := replaceServiceConfigSecrets(input, &secrets)

	assert.NoError(t, err)
	assert.Equal(t, expected, string(result))
}
