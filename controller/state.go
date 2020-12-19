package controller

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/argoproj-labs/argocd-notifications/pkg/services"
)

func stateItemKey(trigger string, oncePer string, dest services.Destination) string {
	key := fmt.Sprintf("%s:%s:%s", trigger, dest.Service, dest.Recipient)
	if oncePer != "" {
		key = oncePer + ":" + key
	}
	return key
}

// notificationState track notification triggers state (already notified/not notified)
type notificationState map[string]int64

// truncate ensures that state has no more than specified number of items and
// removes unnecessary items starting from oldest
func (s notificationState) truncate(maxSize int) {
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

// setAlreadyNotified set the state of given trigger/destination and return if state has been changed
func (s notificationState) setAlreadyNotified(trigger string, oncePer string, dest services.Destination, isNotified bool) bool {
	key := stateItemKey(trigger, oncePer, dest)
	if _, alreadyNotified := s[key]; alreadyNotified == isNotified {
		return false
	}
	if isNotified {
		s[key] = time.Now().Unix()
	} else {
		if oncePer != "" {
			return false
		}
		delete(s, key)
	}
	return true
}

func newState(val string) notificationState {
	if val == "" {
		return notificationState{}
	}
	res := notificationState{}
	if err := json.Unmarshal([]byte(val), &res); err != nil {
		return notificationState{}
	}
	return res
}
