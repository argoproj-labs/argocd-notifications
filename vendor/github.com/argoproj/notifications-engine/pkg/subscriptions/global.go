package subscriptions

import (
	"encoding/json"

	"k8s.io/apimachinery/pkg/labels"
)

type rawSubscription struct {
	Recipients []string `json:"recipients"`
	Triggers   []string `json:"triggers"`
	Selector   string   `json:"selector"`
}

// DefaultSubscription holds recipients that receives notification by default.
type DefaultSubscription struct {
	// Recipients comma separated list of recipients
	Recipients []string
	// Optional trigger name
	Triggers []string
	// Options label selector that limits applied applications
	Selector labels.Selector
}

func (s *DefaultSubscription) MatchesTrigger(trigger string) bool {
	if len(s.Triggers) == 0 {
		return true
	}
	for i := range s.Triggers {
		if s.Triggers[i] == trigger {
			return true
		}
	}
	return false
}

func (s *DefaultSubscription) UnmarshalJSON(data []byte) error {
	raw := rawSubscription{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	s.Triggers = raw.Triggers
	s.Recipients = raw.Recipients
	selector, err := labels.Parse(raw.Selector)
	if err != nil {
		return err
	}
	s.Selector = selector
	return nil
}

func (s *DefaultSubscription) MarshalJSON() ([]byte, error) {
	raw := rawSubscription{
		Triggers:   s.Triggers,
		Recipients: s.Recipients,
	}
	if s.Selector != nil {
		raw.Selector = s.Selector.String()
	}
	return json.Marshal(raw)
}

type DefaultSubscriptions []DefaultSubscription
