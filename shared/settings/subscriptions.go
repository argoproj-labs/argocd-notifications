package settings

import (
	"encoding/json"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
)

type rawSubscription struct {
	Recipients []string `json:"recipients"`
	Triggers   []string `json:"triggers"`
	Selector   string   `json:"selector"`
}

// DefaultSubscription holds recipients that receives notification by default.
type Subscription struct {
	// Recipients comma separated list of recipients
	Recipients []string
	// Optional trigger name
	Triggers []string
	// Options label selector that limits applied applications
	Selector labels.Selector
}

func (s *Subscription) MatchesTrigger(trigger string) bool {
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

func (s *Subscription) UnmarshalJSON(data []byte) error {
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

func (s *Subscription) MarshalJSON() ([]byte, error) {
	raw := rawSubscription{
		Triggers:   s.Triggers,
		Recipients: s.Recipients,
	}
	if s.Selector != nil {
		raw.Selector = s.Selector.String()
	}
	return json.Marshal(raw)
}

type DefaultSubscriptions []Subscription

// Returns list of recipients for the specified trigger
func (subscriptions DefaultSubscriptions) GetRecipients(trigger string, labels map[string]string) []string {
	var result []string
	for _, s := range subscriptions {
		if s.MatchesTrigger(trigger) && s.Selector.Matches(fields.Set(labels)) {
			result = append(result, s.Recipients...)
		}
	}
	return result
}
