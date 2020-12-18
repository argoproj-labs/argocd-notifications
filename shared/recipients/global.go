package recipients

import (
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
) (Recipients, error) {

	res := map[triggerTemplate][]services.Destination{}
	for trigger, t := range triggersByName {
		for _, s := range subscriptions {
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
