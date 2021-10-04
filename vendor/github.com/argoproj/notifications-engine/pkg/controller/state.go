package controller

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/argoproj/notifications-engine/pkg/services"
	"github.com/argoproj/notifications-engine/pkg/subscriptions"
	"github.com/argoproj/notifications-engine/pkg/triggers"
)

const (
	notifiedHistoryMaxSize = 100
	NotifiedAnnotationKey  = "notified." + subscriptions.AnnotationPrefix
)

func StateItemKey(trigger string, conditionResult triggers.ConditionResult, dest services.Destination) string {
	key := fmt.Sprintf("%s:%s:%s:%s", trigger, conditionResult.Key, dest.Service, dest.Recipient)
	if conditionResult.OncePer != "" {
		key = conditionResult.OncePer + ":" + key
	}
	return key
}

// NotificationsState track notification triggers state (already notified/not notified)
type NotificationsState map[string]int64

// truncate ensures that state has no more than specified number of items and
// removes unnecessary items starting from oldest
func (s NotificationsState) truncate(maxSize int) {
	if cnt := len(s) - maxSize; cnt > 0 {
		var keys []string
		for k := range s {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool {
			return s[keys[i]] < s[keys[j]]
		})

		for i := 0; i < cnt; i++ {
			delete(s, keys[i])
		}
	}
}

// SetAlreadyNotified set the state of given trigger/destination and return if state has been changed
func (s NotificationsState) SetAlreadyNotified(trigger string, result triggers.ConditionResult, dest services.Destination, isNotified bool) bool {
	key := StateItemKey(trigger, result, dest)
	if _, alreadyNotified := s[key]; alreadyNotified == isNotified {
		return false
	}
	if isNotified {
		s[key] = time.Now().Unix()
	} else {
		if result.OncePer != "" {
			return false
		}
		delete(s, key)
	}
	return true
}

func (s NotificationsState) Persist(res metav1.Object) (map[string]string, error) {
	s.truncate(notifiedHistoryMaxSize)

	annotations := map[string]string{}

	if res.GetAnnotations() != nil {
		for k, v := range res.GetAnnotations() {
			annotations[k] = v
		}
	}

	if len(s) == 0 {
		delete(annotations, NotifiedAnnotationKey)
	} else {
		stateJson, err := json.Marshal(s)
		if err != nil {
			return nil, err
		}
		annotations[NotifiedAnnotationKey] = string(stateJson)
	}

	return annotations, nil
}

func NewState(val string) NotificationsState {
	if val == "" {
		return NotificationsState{}
	}
	res := NotificationsState{}
	if err := json.Unmarshal([]byte(val), &res); err != nil {
		return NotificationsState{}
	}
	return res
}

func NewStateFromRes(res metav1.Object) NotificationsState {
	if annotations := res.GetAnnotations(); annotations != nil {
		return NewState(annotations[NotifiedAnnotationKey])
	}
	return NotificationsState{}
}
