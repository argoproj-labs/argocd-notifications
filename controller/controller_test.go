package controller

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	kubetesting "k8s.io/client-go/testing"

	"github.com/argoproj-labs/argocd-notifications/pkg"
	"github.com/argoproj-labs/argocd-notifications/pkg/controller"
	"github.com/argoproj-labs/argocd-notifications/pkg/mocks"
	"github.com/argoproj-labs/argocd-notifications/pkg/services"
	"github.com/argoproj-labs/argocd-notifications/pkg/triggers"
	"github.com/argoproj-labs/argocd-notifications/shared/legacy"
	"github.com/argoproj-labs/argocd-notifications/shared/settings"
	. "github.com/argoproj-labs/argocd-notifications/testing"
)

var (
	logEntry = logrus.NewEntry(logrus.New())
)

func mustToJson(val interface{}) string {
	res, err := json.Marshal(val)
	if err != nil {
		panic(err)
	}
	return string(res)
}

func newController(t *testing.T, ctx context.Context, client dynamic.Interface) (*notificationController, *mocks.MockAPI, error) {
	mockCtrl := gomock.NewController(t)
	go func() {
		<-ctx.Done()
		mockCtrl.Finish()
	}()
	api := mocks.NewMockAPI(mockCtrl)
	cfg := settings.Config{Config: pkg.Config{}, API: api}
	c, err := NewController(client, TestNamespace, cfg, "", NewMetricsRegistry())
	if err != nil {
		return nil, nil, err
	}
	err = c.Init(ctx)
	if err != nil {
		return nil, nil, err
	}
	return c.(*notificationController), api, err
}

func TestSendsNotificationIfTriggered(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	app := NewApp("test", WithAnnotations(map[string]string{
		controller.SubscribeAnnotationKey("my-trigger", "mock"): "recipient",
	}))

	ctrl, api, err := newController(t, ctx, NewFakeClient(app))
	assert.NoError(t, err)

	receivedVars := map[string]interface{}{}
	api.EXPECT().RunTrigger("my-trigger", gomock.Any()).Return([]triggers.ConditionResult{{Triggered: true, Templates: []string{"test"}}}, nil)
	api.EXPECT().Send(mock.MatchedBy(func(vars map[string]interface{}) bool {
		receivedVars = vars
		return true
	}), []string{"test"}, services.Destination{Service: "mock", Recipient: "recipient"}).Return(nil)

	err = ctrl.processApp(app, logEntry)

	assert.NoError(t, err)

	state := controller.NewState(app.GetAnnotations()[controller.NotifiedAnnotationKey])
	assert.NotNil(t, state[controller.StateItemKey("mock", triggers.ConditionResult{}, services.Destination{Service: "mock", Recipient: "recipient"})])
	assert.Equal(t, app.Object, receivedVars["app"])
	assert.Equal(t, legacy.InjectLegacyVar(ctrl.cfg.Context, "mock"), receivedVars["context"])
}

func TestSendsNotificationIfProjectTriggered(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	appProj := NewProject("default", WithAnnotations(map[string]string{
		controller.SubscribeAnnotationKey("my-trigger", "mock"): "recipient",
	}))
	app := NewApp("test", WithProject("default"))

	ctrl, api, err := newController(t, ctx, NewFakeClient(app, appProj))
	assert.NoError(t, err)

	receivedVars := map[string]interface{}{}
	api.EXPECT().RunTrigger("my-trigger", gomock.Any()).Return([]triggers.ConditionResult{{Triggered: true, Templates: []string{"test"}}}, nil)
	api.EXPECT().Send(mock.MatchedBy(func(vars map[string]interface{}) bool {
		receivedVars = vars
		return true
	}), []string{"test"}, services.Destination{Service: "mock", Recipient: "recipient"}).Return(nil)

	err = ctrl.processApp(app, logEntry)

	assert.NoError(t, err)

	state := controller.NewState(app.GetAnnotations()[controller.NotifiedAnnotationKey])
	assert.NotNil(t, state[controller.StateItemKey("mock", triggers.ConditionResult{}, services.Destination{Service: "mock", Recipient: "recipient"})])
	assert.Equal(t, app.Object, receivedVars["app"])
	assert.Equal(t, legacy.InjectLegacyVar(ctrl.cfg.Context, "mock"), receivedVars["context"])
}

func TestDoesNotSendNotificationIfAnnotationPresent(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	state := controller.NotificationsState{}
	_ = state.SetAlreadyNotified("my-trigger", triggers.ConditionResult{}, services.Destination{Service: "mock", Recipient: "recipient"}, true)
	app := NewApp("test", WithAnnotations(map[string]string{
		controller.SubscribeAnnotationKey("my-trigger", "mock"): "recipient",
		controller.NotifiedAnnotationKey:                        mustToJson(state),
	}))
	ctrl, api, err := newController(t, ctx, NewFakeClient(app))
	assert.NoError(t, err)

	api.EXPECT().RunTrigger("my-trigger", gomock.Any()).Return([]triggers.ConditionResult{{Triggered: true, Templates: []string{"test"}}}, nil)

	err = ctrl.processApp(app, logEntry)

	assert.NoError(t, err)
}

func TestRemovesAnnotationIfNoTrigger(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	state := controller.NotificationsState{}
	_ = state.SetAlreadyNotified("my-trigger", triggers.ConditionResult{}, services.Destination{Service: "mock", Recipient: "recipient"}, true)
	app := NewApp("test", WithAnnotations(map[string]string{
		controller.SubscribeAnnotationKey("my-trigger", "mock"): "recipient",
		controller.NotifiedAnnotationKey:                        mustToJson(state),
	}))
	ctrl, api, err := newController(t, ctx, NewFakeClient(app))
	assert.NoError(t, err)

	api.EXPECT().RunTrigger("my-trigger", gomock.Any()).Return([]triggers.ConditionResult{{Triggered: false}}, nil)

	err = ctrl.processApp(app, logEntry)

	assert.NoError(t, err)
	state = controller.NewState(app.GetAnnotations()[controller.NotifiedAnnotationKey])
	assert.Empty(t, state)
}

func TestUpdatedAnnotationsSavedAsPatch(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	state := controller.NotificationsState{}
	_ = state.SetAlreadyNotified("my-trigger", triggers.ConditionResult{}, services.Destination{Service: "mock", Recipient: "recipient"}, true)

	app := NewApp("test", WithAnnotations(map[string]string{
		controller.SubscribeAnnotationKey("my-trigger", "mock"): "recipient",
		controller.NotifiedAnnotationKey:                        mustToJson(state),
	}))

	patchCh := make(chan []byte)

	client := NewFakeClient(app)
	client.PrependReactor("patch", "*", func(action kubetesting.Action) (handled bool, ret runtime.Object, err error) {
		patchCh <- action.(kubetesting.PatchAction).GetPatch()
		return true, nil, nil
	})
	ctrl, api, err := newController(t, ctx, client)
	assert.NoError(t, err)
	api.EXPECT().RunTrigger("my-trigger", gomock.Any()).Return([]triggers.ConditionResult{{Triggered: false}}, nil)

	go ctrl.Run(ctx, 1)

	select {
	case <-time.After(time.Second * 5):
		t.Error("application was not patched")
	case patchData := <-patchCh:
		patch := map[string]interface{}{}
		err = json.Unmarshal(patchData, &patch)
		assert.NoError(t, err)
		val, ok, err := unstructured.NestedFieldNoCopy(patch, "metadata", "annotations", controller.NotifiedAnnotationKey)
		assert.NoError(t, err)
		assert.True(t, ok)
		assert.Nil(t, val)
	}
}

func TestAppSyncStatusRefreshed(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	ctrl, _, err := newController(t, ctx, NewFakeClient())
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

func TestAnnotationIsTheSame(t *testing.T) {
	t.Run("same", func(t *testing.T) {
		app1 := NewApp("test", WithAnnotations(map[string]string{
			controller.SubscribeAnnotationKey("my-trigger", "mock"): "recipient",
		}))
		app2 := NewApp("test", WithAnnotations(map[string]string{
			controller.SubscribeAnnotationKey("my-trigger", "mock"): "recipient",
		}))
		assert.True(t, mapsEqual(app1.GetAnnotations(), app2.GetAnnotations()))
	})

	t.Run("same-nil-nil", func(t *testing.T) {
		app1 := NewApp("test", WithAnnotations(nil))
		app2 := NewApp("test", WithAnnotations(nil))
		assert.True(t, mapsEqual(app1.GetAnnotations(), app2.GetAnnotations()))
	})

	t.Run("same-nil-emptyMap", func(t *testing.T) {
		app1 := NewApp("test", WithAnnotations(nil))
		app2 := NewApp("test", WithAnnotations(map[string]string{}))
		assert.True(t, mapsEqual(app1.GetAnnotations(), app2.GetAnnotations()))
	})

	t.Run("same-emptyMap-nil", func(t *testing.T) {
		app1 := NewApp("test", WithAnnotations(map[string]string{}))
		app2 := NewApp("test", WithAnnotations(nil))
		assert.True(t, mapsEqual(app1.GetAnnotations(), app2.GetAnnotations()))
	})

	t.Run("same-emptyMap-emptyMap", func(t *testing.T) {
		app1 := NewApp("test", WithAnnotations(map[string]string{}))
		app2 := NewApp("test", WithAnnotations(map[string]string{}))
		assert.True(t, mapsEqual(app1.GetAnnotations(), app2.GetAnnotations()))
	})

	t.Run("notSame-nil-map", func(t *testing.T) {
		app1 := NewApp("test", WithAnnotations(nil))
		app2 := NewApp("test", WithAnnotations(map[string]string{
			controller.SubscribeAnnotationKey("my-trigger", "mock"): "recipient",
		}))
		assert.False(t, mapsEqual(app1.GetAnnotations(), app2.GetAnnotations()))
	})

	t.Run("notSame-map-nil", func(t *testing.T) {
		app1 := NewApp("test", WithAnnotations(map[string]string{
			controller.SubscribeAnnotationKey("my-trigger", "mock"): "recipient",
		}))
		app2 := NewApp("test", WithAnnotations(nil))
		assert.False(t, mapsEqual(app1.GetAnnotations(), app2.GetAnnotations()))
	})

	t.Run("notSame-map-map", func(t *testing.T) {
		app1 := NewApp("test", WithAnnotations(map[string]string{
			controller.SubscribeAnnotationKey("my-trigger", "mock"): "recipient",
		}))
		app2 := NewApp("test", WithAnnotations(map[string]string{
			controller.SubscribeAnnotationKey("my-trigger", "mock"): "recipient2",
		}))
		assert.False(t, mapsEqual(app1.GetAnnotations(), app2.GetAnnotations()))
	})
}
