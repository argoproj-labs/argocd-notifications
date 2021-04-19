package legacy

import (
	"strings"

	"github.com/argoproj-labs/argocd-notifications/pkg"

	"github.com/argoproj-labs/argocd-notifications/pkg/services"
	"github.com/argoproj-labs/argocd-notifications/pkg/util/text"
)

const (
	annotationKey = "recipients.argocd-notifications.argoproj.io"
)

func GetSubscriptions(annotations map[string]string, defaultTriggers []string, serviceDefaultTriggers map[string][]string) pkg.Subscriptions {
	subscriptions := pkg.Subscriptions{}
	for k, v := range annotations {
		if !strings.HasSuffix(k, annotationKey) {
			continue
		}

		var triggerNames []string
		triggerName := strings.TrimRight(k[0:len(k)-len(annotationKey)], ".")
		if triggerName == "" {
			triggerNames = defaultTriggers
		} else {
			triggerNames = []string{triggerName}
		}

		for _, recipient := range text.SplitRemoveEmpty(v, ",") {
			if recipient = strings.TrimSpace(recipient); recipient != "" {
				parts := strings.Split(recipient, ":")
				dest := services.Destination{Service: parts[0]}
				if len(parts) > 1 {
					dest.Recipient = parts[1]
				}

				t := triggerNames
				if v, ok := serviceDefaultTriggers[dest.Service]; ok {
					t = v
				}
				for _, name := range t {
					subscriptions[name] = append(subscriptions[name], dest)
				}
			}
		}
	}
	return subscriptions
}
