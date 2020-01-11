package controller

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/golang/mock/gomock"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic/fake"

	"github.com/argoproj-labs/argocd-notifications/notifiers"
	notifiermocks "github.com/argoproj-labs/argocd-notifications/notifiers/mocks"
	. "github.com/argoproj-labs/argocd-notifications/testing"
	"github.com/argoproj-labs/argocd-notifications/triggers"
	triggermocks "github.com/argoproj-labs/argocd-notifications/triggers/mocks"
)

var (
	logEntry = logrus.NewEntry(logrus.New())
)

func newController(t *testing.T, ctx context.Context, objs ...runtime.Object) (*notificationController, *triggermocks.MockTrigger, *notifiermocks.MockNotifier, error) {
	mockCtrl := gomock.NewController(t)
	go func() {
		select {
		case <-ctx.Done():
			mockCtrl.Finish()
		}
	}()
	trigger := triggermocks.NewMockTrigger(mockCtrl)
	notifier := notifiermocks.NewMockNotifier(mockCtrl)
	c, err := NewController(
		fake.NewSimpleDynamicClient(runtime.NewScheme(), objs...),
		TestNamespace,
		map[string]triggers.Trigger{"mock": trigger},
		map[string]notifiers.Notifier{"mock": notifier}, map[string]string{},
		"")
	if err != nil {
		return nil, nil, nil, err
	}
	return c.(*notificationController), trigger, notifier, err
}

func TestSendsNotificationIfTriggered(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	app := NewApp("test", WithAnnotations(map[string]string{
		recipientsAnnotation: "mock:recipient",
	}))
	ctrl, trigger, notifier, err := newController(t, ctx, app)
	assert.NoError(t, err)

	trigger.EXPECT().Triggered(app).Return(true, nil)
	trigger.EXPECT().FormatNotification(app, map[string]string{}).Return("title", "body", nil)
	notifier.EXPECT().Send("title", "body", "recipient").Return(nil)

	err = ctrl.processApp(app, logEntry)

	assert.NoError(t, err)
	assert.NotEmpty(t, app.GetAnnotations()[fmt.Sprintf("mock.%s", annotationPostfix)])
}

func TestDoesNotSendNotificationIfAnnotationPresent(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	app := NewApp("test", WithAnnotations(map[string]string{
		recipientsAnnotation:                      "mock:recipient",
		fmt.Sprintf("mock.%s", annotationPostfix): time.Now().Format(time.RFC3339),
	}))
	ctrl, trigger, _, err := newController(t, ctx, app)
	assert.NoError(t, err)

	trigger.EXPECT().Triggered(app).Return(true, nil)

	err = ctrl.processApp(app, logEntry)

	assert.NoError(t, err)
}

func TestRemovesAnnotationIfNoTrigger(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	app := NewApp("test", WithAnnotations(map[string]string{
		recipientsAnnotation:                      "mock:recipient",
		fmt.Sprintf("mock.%s", annotationPostfix): time.Now().Format(time.RFC3339),
	}))
	ctrl, trigger, _, err := newController(t, ctx, app)
	assert.NoError(t, err)

	trigger.EXPECT().Triggered(app).Return(false, nil)

	err = ctrl.processApp(app, logEntry)

	assert.NoError(t, err)
	assert.Empty(t, app.GetAnnotations()[fmt.Sprintf("mock.%s", annotationPostfix)])
}
