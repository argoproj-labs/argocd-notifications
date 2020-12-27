package triggers

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/argoproj-labs/argocd-notifications/pkg/services"
)

func StateItemKey(trigger string, conditionResult ConditionResult, dest services.Destination) string {
	key := fmt.Sprintf("%s:%s:%s:%s", trigger, conditionResult.Key, dest.Service, dest.Recipient)
	if conditionResult.OncePer != "" {
		key = conditionResult.OncePer + ":" + key
	}
	return key
}

// State track notification triggers state (already notified/not notified)
type State map[string]int64

// Truncate ensures that state has no more than specified number of items and
// removes unnecessary items starting from oldest
func (s State) Truncate(maxSize int) {
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
func (s State) SetAlreadyNotified(trigger string, result ConditionResult, dest services.Destination, isNotified bool) bool {
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

func NewState(val string) State {
	if val == "" {
		return State{}
	}
	res := State{}
	if err := json.Unmarshal([]byte(val), &res); err != nil {
		return State{}
	}
	return res
}
