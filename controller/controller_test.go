package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/argoproj-labs/argocd-notifications/pkg/services"

	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/fake"
	kubetesting "k8s.io/client-go/testing"

	"github.com/argoproj-labs/argocd-notifications/pkg/mocks"
	"github.com/argoproj-labs/argocd-notifications/shared/recipients"
	"github.com/argoproj-labs/argocd-notifications/shared/settings"
	. "github.com/argoproj-labs/argocd-notifications/testing"
	"github.com/argoproj-labs/argocd-notifications/triggers"
	triggermocks "github.com/argoproj-labs/argocd-notifications/triggers/mocks"
)

var (
	logEntry = logrus.NewEntry(logrus.New())
)

func newController(t *testing.T, ctx context.Context, client dynamic.Interface) (*notificationController, *triggermocks.MockTrigger, *mocks.MockNotifier, error) {
	mockCtrl := gomock.NewController(t)
	go func() {
		<-ctx.Done()
		mockCtrl.Finish()
	}()
	trigger := triggermocks.NewMockTrigger(mockCtrl)
	notifier := mocks.NewMockNotifier(mockCtrl)
	c, err := NewController(
		client,
		TestNamespace,
		settings.Config{
			Notifier:        notifier,
			Triggers:        map[string]triggers.Trigger{"mock": trigger},
			DefaultTriggers: []string{"mock"},
		},
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
		recipients.AnnotationKey: "mock:recipient",
	}))
	ctrl, trigger, notifier, err := newController(t, ctx, fake.NewSimpleDynamicClient(runtime.NewScheme(), app))
	assert.NoError(t, err)

	trigger.EXPECT().Triggered(app).Return(true, nil)
	trigger.EXPECT().GetTemplate().Return("test")

	receivedVars := map[string]interface{}{}
	notifier.EXPECT().Send(mock.MatchedBy(func(vars map[string]interface{}) bool {
		receivedVars = vars
		return true
	}), "test", services.Destination{Service: "mock", Recipient: "recipient"}).Return(nil)

	err = ctrl.processApp(app, logEntry)

	assert.NoError(t, err)

	annotation := app.GetAnnotations()[fmt.Sprintf("mock.%s", recipients.AnnotationPostfix)]
	assert.NotEmpty(t, annotation)
	assert.Contains(t, annotation, "mock:recipient")
	assert.Equal(t, app.Object, receivedVars["app"])
	assert.Equal(t, ctrl.cfg.Context, receivedVars["context"])
}

func TestDoesNotSendNotificationIfAnnotationPresent(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	app := NewApp("test", WithAnnotations(map[string]string{
		recipients.AnnotationKey:                             "mock:recipient",
		fmt.Sprintf("mock.%s", recipients.AnnotationPostfix): fmt.Sprintf("mock:recipient=%s\n", time.Now().Format(time.RFC3339)),
	}))
	ctrl, trigger, _, err := newController(t, ctx, fake.NewSimpleDynamicClient(runtime.NewScheme(), app))
	assert.NoError(t, err)

	trigger.EXPECT().Triggered(app).Return(true, nil)
	trigger.EXPECT().GetTemplate().Return("test")

	err = ctrl.processApp(app, logEntry)

	assert.NoError(t, err)
}

func TestSendsNotificationIfAnnotationPresentInStaleCache(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	staleApp := NewApp("test", WithAnnotations(map[string]string{
		recipients.AnnotationKey:                             "mock:recipient",
		fmt.Sprintf("mock.%s", recipients.AnnotationPostfix): fmt.Sprintf("mock:recipient=%s\n", time.Now().Format(time.RFC3339)),
	}))
	refreshedApp := NewApp("test", WithAnnotations(map[string]string{
		recipients.AnnotationKey: "mock:recipient",
	}))
	ctrl, trigger, notifier, err := newController(t, ctx, fake.NewSimpleDynamicClient(runtime.NewScheme(), refreshedApp))
	assert.NoError(t, err)

	receivedVars := map[string]interface{}{}

	trigger.EXPECT().Triggered(staleApp).Return(true, nil)
	trigger.EXPECT().GetTemplate().Return("test")
	notifier.EXPECT().Send(mock.MatchedBy(func(vars map[string]interface{}) bool {
		receivedVars = vars
		return true
	}), "test", services.Destination{Service: "mock", Recipient: "recipient"}).Return(nil)

	err = ctrl.processApp(staleApp, logEntry)

	assert.NoError(t, err)
	assert.Equal(t, staleApp.Object, receivedVars["app"])
	assert.Equal(t, ctrl.cfg.Context, receivedVars["context"])
}

func TestRemovesAnnotationIfNoTrigger(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	app := NewApp("test", WithAnnotations(map[string]string{
		recipients.AnnotationKey:                             "mock:recipient",
		fmt.Sprintf("mock.%s", recipients.AnnotationPostfix): fmt.Sprintf("mock:recipient=%s\n", time.Now().Format(time.RFC3339)),
	}))
	ctrl, trigger, _, err := newController(t, ctx, fake.NewSimpleDynamicClient(runtime.NewScheme(), app))
	assert.NoError(t, err)

	trigger.EXPECT().Triggered(app).Return(false, nil)
	trigger.EXPECT().GetTemplate().Return("test")

	err = ctrl.processApp(app, logEntry)

	assert.NoError(t, err)
	assert.Empty(t, app.GetAnnotations()[fmt.Sprintf("mock.%s", recipients.AnnotationPostfix)])
}

func TestUpdatedAnnotationsSavedAsPatch(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	app := NewApp("test", WithAnnotations(map[string]string{
		recipients.AnnotationKey:                             "mock:recipient",
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
	trigger.EXPECT().GetTemplate().Return("test")

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
