package pkg

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/argoproj-labs/argocd-notifications/pkg/services"
	"github.com/argoproj-labs/argocd-notifications/pkg/services/mocks"
)

func getConfig(ctrl *gomock.Controller, opts ...func(service *mocks.MockNotificationService)) Config {
	return Config{
		Templates: map[string]services.Notification{
			"my-template": {
				Message: "hello {{ .foo }} {{ .serviceType }}:{{ .recipient }}",
			},
		},
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
func TestSend(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, err := NewAPI(getConfig(ctrl, func(service *mocks.MockNotificationService) {
		service.EXPECT().Send(services.Notification{
			Message: "hello world slack:my-channel",
		}, services.Destination{
			Service:   "slack",
			Recipient: "my-channel",
		}).Return(nil)
	}))
	if !assert.NoError(t, err) {
		return
	}

	err = api.Send(
		map[string]interface{}{"foo": "world"},
		[]string{"my-template"},
		services.Destination{Service: "slack", Recipient: "my-channel"},
	)
	assert.NoError(t, err)
}

func TestAddService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, err := NewAPI(getConfig(ctrl))
	if !assert.NoError(t, err) {
		return
	}

	api.AddNotificationService("hello", mocks.NewMockNotificationService(ctrl))

	servicesMap := api.GetNotificationServices()
	assert.NotNil(t, servicesMap["slack"])
	assert.NotNil(t, servicesMap["hello"])
}
