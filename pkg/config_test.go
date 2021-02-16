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

func TestReplaceServiceConfigSecret(t *testing.T) {
	tests := []struct {
		config map[string]interface{}
		secret map[string][]byte
		want   map[string]interface{}
	}{
		{
			config: map[string]interface{}{
				"url": "$endpoint",
				"headers": []map[string]interface{}{
					{
						"name":  "Authorization",
						"value": "Bearer $secret-value",
					},
				},
			},
			secret: map[string][]byte{
				"endpoint":     []byte("https://example.com"),
				"secret-value": []byte("token"),
			},
			want: map[string]interface{}{
				"url": "https://example.com",
				"headers": []map[string]interface{}{
					{
						"name":  "Authorization",
						"value": "Bearer token",
					},
				},
			},
		},
		{
			config: map[string]interface{}{
				"apiUrl": "$endpoint",
				"apiKeys": map[string]interface{}{
					"first-team":  "$first-team-secret",
					"second-team": "$second-team-secret",
				},
			},
			secret: map[string][]byte{
				"first-team-secret":  []byte("first-token"),
				"second-team-secret": []byte("second-token"),
			},
			want: map[string]interface{}{
				"apiUrl": "$endpoint",
				"apiKeys": map[string]interface{}{
					"first-team":  "first-token",
					"second-team": "second-token",
				},
			},
		},
	}

	for _, tt := range tests {
		result := replaceServiceConfigSecret(tt.config, tt.secret)
		assert.Equal(t, tt.want, result)
	}
}
