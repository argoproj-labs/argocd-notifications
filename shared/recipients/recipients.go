package recipients

import (
	"fmt"
	"strings"

	"github.com/argoproj-labs/argocd-notifications/pkg"
	"github.com/argoproj-labs/argocd-notifications/pkg/services"
)

type triggerTemplate struct {
	trigger  string
	template string
}

type Recipients map[triggerTemplate][]services.Destination

func (r Recipients) Merge(other Recipients) Recipients {
	res := map[triggerTemplate][]services.Destination{}
	for k := range r {
		res[k] = r[k]
	}
	for k := range other {
		res[k] = append(res[k], other[k]...)
	}
	return res
}

func (r Recipients) GetNotificationSubscriptions() []pkg.NotificationSubscription {
	var subscriptions []pkg.NotificationSubscription
	for tt, destinations := range r {
		subscriptions = append(subscriptions, pkg.NotificationSubscription{
			When: tt.trigger,
			Send: tt.template,
			To:   destinations,
		})
	}
	return subscriptions
}

func ParseDestinationAndTemplate(recipient string) (services.Destination, string, error) {
	parts := strings.Split(recipient, ":")
	if len(parts) < 2 {
		return services.Destination{}, "", fmt.Errorf("%s is not valid recipient. Expected recipient format is <serviceType>:<name>(:template)", recipient)
	}
	dest := services.Destination{Service: parts[0], Recipient: parts[1]}
	templ := ""
	if len(parts) > 2 {
		templ = parts[2]
	}
	return dest, templ, nil
}
