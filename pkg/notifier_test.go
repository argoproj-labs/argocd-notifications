package pkg

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/argoproj-labs/argocd-notifications/pkg/services"
	"github.com/argoproj-labs/argocd-notifications/pkg/services/mocks"
	"github.com/argoproj-labs/argocd-notifications/pkg/templates"
)

func getConfig(ctrl *gomock.Controller, opts ...func(service *mocks.MockNotificationService)) Config {
	return Config{
		Templates: []templates.NotificationTemplate{{
			Name:         "my-template",
			Notification: services.Notification{Body: "hello {{ .foo }}"},
		}},
		Services: map[string]ServiceFactory{
			"slack": func() (services.NotificationService, error) {
				serviceMock := mocks.NewMockNotificationService(ctrl)
				for i := range opts {
					opts[i](serviceMock)
				}
				return serviceMock, nil
			},
		},
	}
}

func TestNewNotifier(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	notifier, err := NewNotifier(getConfig(ctrl))
	if !assert.NoError(t, err) {
		return
	}

	assert.NotNil(t, notifier.services["slack"])
	assert.NotNil(t, notifier.templates["my-template"])
}

func TestSend(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	notifier, err := NewNotifier(getConfig(ctrl, func(service *mocks.MockNotificationService) {
		service.EXPECT().Send(services.Notification{
			Body:    "hello world",
			Webhook: map[string]services.WebhookNotification{},
		}, services.Destination{
			Service:   "slack",
			Recipient: "my-channel",
		}).Return(nil)
	}))
	if !assert.NoError(t, err) {
		return
	}

	err = notifier.Send(
		map[string]interface{}{"foo": "world"},
		"my-template",
		services.Destination{Service: "slack", Recipient: "my-channel"},
	)
	assert.NoError(t, err)
}

func TestAddService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	notifier, err := NewNotifier(getConfig(ctrl))
	if !assert.NoError(t, err) {
		return
	}

	notifier.AddService("hello", mocks.NewMockNotificationService(ctrl))

	servicesMap := notifier.GetServices()
	assert.NotNil(t, servicesMap["slack"])
	assert.NotNil(t, servicesMap["hello"])
}
