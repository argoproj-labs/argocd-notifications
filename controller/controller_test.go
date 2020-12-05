package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/fake"
	kubetesting "k8s.io/client-go/testing"

	"github.com/argoproj-labs/argocd-notifications/notifiers"
	notifiermocks "github.com/argoproj-labs/argocd-notifications/notifiers/mocks"
	"github.com/argoproj-labs/argocd-notifications/shared/recipients"
	"github.com/argoproj-labs/argocd-notifications/shared/settings"
	. "github.com/argoproj-labs/argocd-notifications/testing"
	"github.com/argoproj-labs/argocd-notifications/triggers"
	triggermocks "github.com/argoproj-labs/argocd-notifications/triggers/mocks"
)

var (
	logEntry = logrus.NewEntry(logrus.New())
)

func newController(t *testing.T, ctx context.Context, client dynamic.Interface, subscriptions ...settings.Subscription) (*notificationController, *triggermocks.MockTrigger, *notifiermocks.MockNotifier, error) {
	mockCtrl := gomock.NewController(t)
	go func() {
		<-ctx.Done()
		mockCtrl.Finish()
	}()
	trigger := triggermocks.NewMockTrigger(mockCtrl)
	notifier := notifiermocks.NewMockNotifier(mockCtrl)
	c, err := NewController(
		client,
		TestNamespace,
		map[string]triggers.Trigger{"mock": trigger},
		map[string]notifiers.Notifier{"mock": notifier},
		map[string]string{},
		subscriptions,
		"",
		NewMetricsRegistry())
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
		recipients.RecipientsAnnotation: "mock:recipient",
	}))
	ctrl, trigger, notifier, err := newController(t, ctx, fake.NewSimpleDynamicClient(runtime.NewScheme(), app))
	assert.NoError(t, err)

	trigger.EXPECT().GetTemplateName().Return("test")
	trigger.EXPECT().Triggered(app).Return(true, nil)
	trigger.EXPECT().FormatNotification(app, map[string]string{"notificationType": "mock"}).Return(
		&notifiers.Notification{Title: "title", Body: "body"}, nil)
	notifier.EXPECT().Send(notifiers.Notification{Title: "title", Body: "body"}, "recipient").Return(nil)

	err = ctrl.processApp(app, logEntry)

	assert.NoError(t, err)

	annotation := app.GetAnnotations()[fmt.Sprintf("mock.%s", recipients.AnnotationPostfix)]
	assert.NotEmpty(t, annotation)
	assert.Contains(t, annotation, "mock:recipient")
}

func TestDoesNotSendNotificationIfAnnotationPresent(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	app := NewApp("test", WithAnnotations(map[string]string{
		recipients.RecipientsAnnotation:                      "mock:recipient",
		fmt.Sprintf("mock.%s", recipients.AnnotationPostfix): fmt.Sprintf("mock:recipient=%s\n", time.Now().Format(time.RFC3339)),
	}))
	ctrl, trigger, _, err := newController(t, ctx, fake.NewSimpleDynamicClient(runtime.NewScheme(), app))
	assert.NoError(t, err)

	trigger.EXPECT().Triggered(app).Return(true, nil)

	err = ctrl.processApp(app, logEntry)

	assert.NoError(t, err)
}

func TestSendsNotificationIfAnnotationPresentInStaleCache(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	staleApp := NewApp("test", WithAnnotations(map[string]string{
		recipients.RecipientsAnnotation:                      "mock:recipient",
		fmt.Sprintf("mock.%s", recipients.AnnotationPostfix): fmt.Sprintf("mock:recipient=%s\n", time.Now().Format(time.RFC3339)),
	}))
	refreshedApp := NewApp("test", WithAnnotations(map[string]string{
		recipients.RecipientsAnnotation: "mock:recipient",
	}))
	ctrl, trigger, notifier, err := newController(t, ctx, fake.NewSimpleDynamicClient(runtime.NewScheme(), refreshedApp))
	assert.NoError(t, err)

	trigger.EXPECT().GetTemplateName().Return("test")
	trigger.EXPECT().Triggered(staleApp).Return(true, nil)
	trigger.EXPECT().FormatNotification(staleApp, map[string]string{"notificationType": "mock"}).Return(
		&notifiers.Notification{Title: "title", Body: "body"}, nil)
	notifier.EXPECT().Send(notifiers.Notification{Title: "title", Body: "body"}, "recipient").Return(nil)

	err = ctrl.processApp(staleApp, logEntry)

	assert.NoError(t, err)
}

func TestRemovesAnnotationIfNoTrigger(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	app := NewApp("test", WithAnnotations(map[string]string{
		recipients.RecipientsAnnotation:                      "mock:recipient",
		fmt.Sprintf("mock.%s", recipients.AnnotationPostfix): fmt.Sprintf("mock:recipient=%s\n", time.Now().Format(time.RFC3339)),
	}))
	ctrl, trigger, _, err := newController(t, ctx, fake.NewSimpleDynamicClient(runtime.NewScheme(), app))
	assert.NoError(t, err)

	trigger.EXPECT().Triggered(app).Return(false, nil)

	err = ctrl.processApp(app, logEntry)

	assert.NoError(t, err)
	assert.Empty(t, app.GetAnnotations()[fmt.Sprintf("mock.%s", recipients.AnnotationPostfix)])
}

func TestGetRecipients(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	app := NewApp("test", WithProject("default"), WithAnnotations(map[string]string{
		recipients.RecipientsAnnotation: "slack:test1",
	}))
	appProj := NewProject("default", WithAnnotations(map[string]string{
		recipients.RecipientsAnnotation:                                        "slack:test2",
		fmt.Sprintf("on-app-sync-unknown.%s", recipients.RecipientsAnnotation): "slack:test3",
	}))
	ctrl, _, _, err := newController(t, ctx, fake.NewSimpleDynamicClient(runtime.NewScheme(), app, appProj))
	assert.NoError(t, err)

	recipients := ctrl.getRecipients(app, "on-app-health-degraded")
	assert.Equal(t, map[string]bool{"slack:test1": true, "slack:test2": true}, recipients)
}

func TestGetRecipients_HasDefaultSubscriptions(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	app := NewApp("test", WithAnnotations(map[string]string{
		recipients.RecipientsAnnotation: "slack:test1",
	}))
	ctrl, _, _, err := newController(t, ctx, fake.NewSimpleDynamicClient(runtime.NewScheme(), app), settings.Subscription{
		Recipients: []string{"slack:test2"}, Selector: labels.NewSelector()})
	assert.NoError(t, err)

	recipients := ctrl.getRecipients(app, "on-app-health-degraded")
	assert.Equal(t, map[string]bool{"slack:test1": true, "slack:test2": true}, recipients)
}

func TestUpdatedAnnotationsSavedAsPatch(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	app := NewApp("test", WithAnnotations(map[string]string{
		recipients.RecipientsAnnotation:                      "mock:recipient",
		fmt.Sprintf("mock.%s", recipients.AnnotationPostfix): fmt.Sprintf("mock:recipient=%s\n", time.Now().Format(time.RFC3339)),
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
	case <-time.After(time.Second * 60):
		t.Error("application was not patched")
	case patchData := <-patchCh:
		patch := map[string]interface{}{}
		err = json.Unmarshal(patchData, &patch)
		assert.NoError(t, err)
		val, ok, err := unstructured.NestedFieldNoCopy(patch, "metadata", "annotations", fmt.Sprintf("mock.%s", recipients.AnnotationPostfix))
		assert.NoError(t, err)
		assert.True(t, ok)
		assert.Nil(t, val)
	}
}

func TestAppSyncStatusRefreshed(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	ctrl, _, _, err := newController(t, ctx, fake.NewSimpleDynamicClient(runtime.NewScheme()))
	assert.NoError(t, err)

	for name, tc := range testsAppSyncStatusRefreshed {
		t.Run(name, func(t *testing.T) {
			if tc.result {
				assert.True(t, ctrl.isAppSyncStatusRefreshed(&unstructured.Unstructured{Object: tc.app}, logEntry))
			} else {
				assert.False(t, ctrl.isAppSyncStatusRefreshed(&unstructured.Unstructured{Object: tc.app}, logEntry))
			}
		})
	}
}

var testsAppSyncStatusRefreshed = map[string]struct {
	app    map[string]interface{}
	result bool
}{
	"MissingOperationState": {app: map[string]interface{}{"status": map[string]interface{}{}}, result: true},
	"MissingOperationStatePhase": {app: map[string]interface{}{
		"status": map[string]interface{}{
			"operationState": map[string]interface{}{},
		},
	}, result: true},
	"RunningOperation": {app: map[string]interface{}{
		"status": map[string]interface{}{
			"operationState": map[string]interface{}{
				"phase": "Running",
			},
		},
	}, result: true},
	"MissingFinishedAt": {app: map[string]interface{}{
		"status": map[string]interface{}{
			"operationState": map[string]interface{}{
				"phase": "Succeeded",
			},
		},
	}, result: false},
	"Reconciled": {app: map[string]interface{}{
		"status": map[string]interface{}{
			"reconciledAt": "2020-03-01T13:37:00Z",
			"observedAt":   "2020-03-01T13:37:00Z",
			"operationState": map[string]interface{}{
				"phase":      "Succeeded",
				"finishedAt": "2020-03-01T13:37:00Z",
			},
		},
	}, result: true},
	"NotYetReconciled": {app: map[string]interface{}{
		"status": map[string]interface{}{
			"reconciledAt": "2020-03-01T00:13:37Z",
			"observedAt":   "2020-03-01T00:13:37Z",
			"operationState": map[string]interface{}{
				"phase":      "Succeeded",
				"finishedAt": "2020-03-01T13:37:00Z",
			},
		},
	}, result: false},
}
