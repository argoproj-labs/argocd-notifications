package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	kubetesting "k8s.io/client-go/testing"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic/fake"

	"k8s.io/client-go/dynamic"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/golang/mock/gomock"

	"github.com/argoproj-labs/argocd-notifications/notifiers"
	notifiermocks "github.com/argoproj-labs/argocd-notifications/notifiers/mocks"
	. "github.com/argoproj-labs/argocd-notifications/testing"
	"github.com/argoproj-labs/argocd-notifications/triggers"
	triggermocks "github.com/argoproj-labs/argocd-notifications/triggers/mocks"
)

var (
	logEntry = logrus.NewEntry(logrus.New())
)

func newController(t *testing.T, ctx context.Context, client dynamic.Interface) (*notificationController, *triggermocks.MockTrigger, *notifiermocks.MockNotifier, error) {
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
		client,
		TestNamespace,
		map[string]triggers.Trigger{"mock": trigger},
		map[string]notifiers.Notifier{"mock": notifier}, map[string]string{},
		"")
	if err != nil {
		return nil, nil, nil, err
	}
	err = c.Init(ctx)
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
	ctrl, trigger, notifier, err := newController(t, ctx, fake.NewSimpleDynamicClient(runtime.NewScheme(), app))
	assert.NoError(t, err)

	trigger.EXPECT().Triggered(app).Return(true, nil)
	trigger.EXPECT().FormatNotification(app, map[string]string{"notificationType": "mock"}).Return(
		&notifiers.Notification{Title: "title", Body: "body"}, nil)
	notifier.EXPECT().Send(notifiers.Notification{Title: "title", Body: "body"}, "recipient").Return(nil)

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
	ctrl, trigger, _, err := newController(t, ctx, fake.NewSimpleDynamicClient(runtime.NewScheme(), app))
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
	ctrl, trigger, _, err := newController(t, ctx, fake.NewSimpleDynamicClient(runtime.NewScheme(), app))
	assert.NoError(t, err)

	trigger.EXPECT().Triggered(app).Return(false, nil)

	err = ctrl.processApp(app, logEntry)

	assert.NoError(t, err)
	assert.Empty(t, app.GetAnnotations()[fmt.Sprintf("mock.%s", annotationPostfix)])
}

func TestGetRecipients(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	app := NewApp("test", WithProject("default"), WithAnnotations(map[string]string{
		recipientsAnnotation: "slack:test1",
	}))
	appProj := NewProject("default", WithAnnotations(map[string]string{
		recipientsAnnotation: "slack:test2",
		fmt.Sprintf("on-app-sync-unknown.%s", recipientsAnnotation): "slack:test3",
	}))
	ctrl, _, _, err := newController(t, ctx, fake.NewSimpleDynamicClient(runtime.NewScheme(), app, appProj))
	assert.NoError(t, err)

	recipients := ctrl.getRecipients(app, "on-app-health-degraded")
	assert.Equal(t, map[string]bool{"slack:test1": true, "slack:test2": true}, recipients)
}

func TestGetRecipientsFromAnnotations_NoTriggerNameInAnnotation(t *testing.T) {
	recipients := getRecipientsFromAnnotations(
		map[string]string{recipientsAnnotation: "slack:test"}, "on-app-sync-unknown")
	assert.ElementsMatch(t, recipients, []string{"slack:test"})
}

func TestGetRecipientsFromAnnotations_HasTriggerNameInAnnotation(t *testing.T) {
	recipients := getRecipientsFromAnnotations(map[string]string{
		recipientsAnnotation: "slack:test",
		fmt.Sprintf("on-app-sync-unknown.%s", recipientsAnnotation):    "slack:test1",
		fmt.Sprintf("on-app-health-degraded.%s", recipientsAnnotation): "slack:test2",
	}, "on-app-sync-unknown")
	assert.ElementsMatch(t, recipients, []string{"slack:test", "slack:test1"})
}

func TestUpdatedAnnotationsSavedAsPatch(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	app := NewApp("test", WithAnnotations(map[string]string{
		recipientsAnnotation:                      "mock:recipient",
		fmt.Sprintf("mock.%s", annotationPostfix): time.Now().Format(time.RFC3339),
	}))

	patchCh := make(chan []byte)

	client := fake.NewSimpleDynamicClient(runtime.NewScheme(), app)
	client.PrependReactor("patch", "*", func(action kubetesting.Action) (handled bool, ret runtime.Object, err error) {
		patchCh <- action.(kubetesting.PatchAction).GetPatch()
		return true, nil, nil
	})
	ctrl, trigger, _, err := newController(t, ctx, client)
	assert.NoError(t, err)

	trigger.EXPECT().Triggered(gomock.Any()).Return(false, nil).AnyTimes()

	go ctrl.Run(ctx, 1)

	select {
	case <-time.After(time.Second * 10000):
		t.Error("application was not patched")
	case patchData := <-patchCh:
		patch := map[string]interface{}{}
		err = json.Unmarshal(patchData, &patch)
		assert.NoError(t, err)
		val, ok, err := unstructured.NestedFieldNoCopy(patch, "metadata", "annotations", fmt.Sprintf("mock.%s", annotationPostfix))
		assert.NoError(t, err)
		assert.True(t, ok)
		assert.Nil(t, val)
	}
}
