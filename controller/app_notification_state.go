package controller

import (
	"context"
	"encoding/json"

	"github.com/argoproj-labs/argocd-notifications/pkg/services"
	"github.com/argoproj-labs/argocd-notifications/pkg/subscriptions"
	"github.com/argoproj-labs/argocd-notifications/pkg/triggers"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
)

const (
	notifiedHistoryMaxSize = 100
	notifiedAnnotationKey  = "notified." + subscriptions.AnnotationPrefix
)

type AppNotificationState struct {
	app       *unstructured.Unstructured
	state     triggers.State
	refreshed bool

	appClient dynamic.ResourceInterface
}

func NewAppState(app *unstructured.Unstructured, appClient dynamic.ResourceInterface) *AppNotificationState {
	ensureAnnotations(app)

	return &AppNotificationState{
		app:       app,
		state:     triggers.NewState(app.GetAnnotations()[notifiedAnnotationKey]),
		refreshed: false,
		appClient: appClient,
	}
}

func (s *AppNotificationState) ClearAlreadyNotified(trigger string, cr triggers.ConditionResult, to services.Destination) (bool, error) {
	return s.setAlreadyNotified(trigger, cr, to, false)
}

// Clear only the in-memory state of the notification, without attempting to optionally
// refresh the cache.
func (s *AppNotificationState) ClearAlreadyNotifiedCache(trigger string, cr triggers.ConditionResult, to services.Destination) bool {
	return s.state.SetAlreadyNotified(trigger, cr, to, false)
}

func (s *AppNotificationState) SetAlreadyNotified(trigger string, cr triggers.ConditionResult, to services.Destination) (bool, error) {
	return s.setAlreadyNotified(trigger, cr, to, true)
}

func (s *AppNotificationState) Persist() error {
	s.state.Truncate(notifiedHistoryMaxSize)

	annotations := s.app.GetAnnotations()

	if len(s.state) == 0 {
		delete(annotations, notifiedAnnotationKey)
	} else {
		stateJson, err := json.Marshal(s.state)
		if err != nil {
			return err
		}
		annotations[notifiedAnnotationKey] = string(stateJson)
	}

	s.app.SetAnnotations(annotations)

	return nil
}

func (s *AppNotificationState) setAlreadyNotified(trigger string, result triggers.ConditionResult, dest services.Destination, isNotified bool) (bool, error) {
	// changes state of specified trigger/destination and returns if state has changed or not
	changed := s.state.SetAlreadyNotified(trigger, result, dest, isNotified)
	// if state changes reload application
	if changed && !s.refreshed {
		refreshedApp, err := s.appClient.Get(context.Background(), s.app.GetName(), v1.GetOptions{})
		if err != nil {
			return false, err
		}
		ensureAnnotations(refreshedApp)
		s.app.GetAnnotations()[notifiedAnnotationKey] = refreshedApp.GetAnnotations()[notifiedAnnotationKey]

		s.state = triggers.NewState(s.app.GetAnnotations()[notifiedAnnotationKey])
		s.refreshed = true
		return s.state.SetAlreadyNotified(trigger, result, dest, isNotified), nil
	}
	return changed, nil
}
