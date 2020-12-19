package recipients

import (
	"fmt"

	"github.com/argoproj-labs/argocd-notifications/pkg/services"
	"github.com/argoproj-labs/argocd-notifications/pkg/util/text"
	"github.com/argoproj-labs/argocd-notifications/shared/settings"
	"github.com/argoproj-labs/argocd-notifications/triggers"
	"k8s.io/apimachinery/pkg/fields"
)

func GetGlobalRecipients(
	labels map[string]string,
	subscriptions settings.DefaultSubscriptions,
	triggersByName map[string]triggers.Trigger,
	defaultTriggers []string,
) (Recipients, error) {

	res := map[triggerTemplate][]services.Destination{}
	for _, s := range subscriptions {
		triggerNames := s.Triggers
		if len(triggerNames) == 0 {
			triggerNames = defaultTriggers
		}
		for _, trigger := range triggerNames {
			t, ok := triggersByName[trigger]
			if !ok {
				return nil, fmt.Errorf("trigger '%s' is not configured", trigger)
			}
			if s.MatchesTrigger(trigger) && s.Selector.Matches(fields.Set(labels)) {
				for _, recipient := range s.Recipients {
					dst, templ, err := ParseDestinationAndTemplate(recipient)
					if err != nil {
						return nil, err
					}
					tt := triggerTemplate{trigger: trigger, template: text.Coalesce(templ, t.GetTemplate())}
					res[tt] = append(res[tt], dst)
				}
			}
		}
	}
	return res, nil
}
