package recipients

import (
	"fmt"
	"strings"

	"github.com/argoproj-labs/argocd-notifications/pkg/services"
	"github.com/argoproj-labs/argocd-notifications/pkg/util/text"
	"github.com/argoproj-labs/argocd-notifications/triggers"
)

const (
	AnnotationPostfix = "argocd-notifications.argoproj.io"
)

var (
	AnnotationKey = "recipients." + AnnotationPostfix
)

func GetRecipientsFromAnnotations(annotations map[string]string, triggersByName map[string]triggers.Trigger, defaultTriggers []string) (Recipients, error) {
	destByTriggerTemplate := map[triggerTemplate][]services.Destination{}
	for k, v := range annotations {
		if !strings.HasSuffix(k, AnnotationKey) {
			continue
		}

		var triggerNames []string
		triggerName := strings.TrimRight(k[0:len(k)-len(AnnotationKey)], ".")
		if triggerName == "" {
			triggerNames = defaultTriggers
		} else {
			triggerNames = []string{triggerName}
		}

		for _, recipient := range text.SplitRemoveEmpty(v, ",") {
			if recipient = strings.TrimSpace(recipient); recipient != "" {

				dest, templ, err := ParseDestinationAndTemplate(recipient)
				if err != nil {
					return nil, err
				}

				for _, name := range triggerNames {
					trigger, ok := triggersByName[name]
					if !ok {
						return nil, fmt.Errorf("trigger '%s' is not configured", name)
					}
					templateName := text.Coalesce(templ, trigger.GetTemplate())
					if templateName == "" {
						return nil, fmt.Errorf("recipient '%s' should include template since trigger '%s' does not have default template", recipient, name)
					}
					tt := triggerTemplate{trigger: name, template: templateName}
					destByTriggerTemplate[tt] = append(destByTriggerTemplate[tt], dest)
				}
			}
		}
	}
	return destByTriggerTemplate, nil
}

func GetAnnotationKeys(annotations map[string]string, trigger string) []string {
	keys := make([]string, 0)
	for k := range annotations {
		if !strings.HasSuffix(k, AnnotationKey) {
			continue
		}
		if name := strings.TrimRight(k[0:len(k)-len(AnnotationKey)], "."); name != "" && name != trigger {
			continue
		}
		keys = append(keys, k)
	}
	return keys
}

func GetDestinations(annotations map[string]string) (map[services.Destination]bool, error) {
	res := map[services.Destination]bool{}
	for k, v := range annotations {
		if !strings.HasSuffix(k, AnnotationKey) {
			continue
		}
		for _, recipient := range text.SplitRemoveEmpty(v, ",") {
			if recipient = strings.TrimSpace(recipient); recipient != "" {

				dest, _, err := ParseDestinationAndTemplate(recipient)
				if err != nil {
					return nil, err
				}
				res[dest] = true
			}
		}
	}
	return res, nil
}

func ParseRecipients(annotation string) []string {
	recipients := make([]string, 0)
	for _, recipient := range text.SplitRemoveEmpty(annotation, ",") {
		if recipient = strings.TrimSpace(recipient); recipient != "" {
			recipients = append(recipients, recipient)
		}
	}
	return recipients
}
