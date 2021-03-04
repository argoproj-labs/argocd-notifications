package services

import (
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

func TestGetTemplater_GitHub(t *testing.T) {
	n := Notification{
		GitHub: &GitHubNotification{
			Status: &GitHubStatus{
				State:     "{{.context.state}}",
				Label:     "continuous-delivery/{{.app.metadata.name}}",
				TargetURL: "{{.context.argocdUrl}}/applications/{{.app.metadata.name}}",
			},
		},
	}
	templater, err := n.GetTemplater("", template.FuncMap{})

	if !assert.NoError(t, err) {
		return
	}

	var notification Notification
	err = templater(&notification, map[string]interface{}{
		"context": map[string]interface{}{
			"argocdUrl": "https://example.com",
			"state":     "success",
		},
		"app": map[string]interface{}{
			"metadata": map[string]interface{}{
				"name": "argocd-notifications",
			},
			"spec": map[string]interface{}{
				"source": map[string]interface{}{
					"repoURL": "https://github.com/argoproj-labs/argocd-notifications.git",
				},
			},
			"status": map[string]interface{}{
				"sync": map[string]interface{}{
					"revision": "0123456789",
				},
			},
		},
	})

	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "https://github.com/argoproj-labs/argocd-notifications.git", notification.GitHub.repoURL)
	assert.Equal(t, "0123456789", notification.GitHub.revision)
	assert.Equal(t, "success", notification.GitHub.Status.State)
	assert.Equal(t, "continuous-delivery/argocd-notifications", notification.GitHub.Status.Label)
	assert.Equal(t, "https://example.com/applications/argocd-notifications", notification.GitHub.Status.TargetURL)
}
