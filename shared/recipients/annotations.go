package recipients

import "strings"

const (
	AnnotationPostfix = "argocd-notifications.argoproj.io"
)

var (
	RecipientsAnnotation = "recipients." + AnnotationPostfix
)

func GetRecipientsFromAnnotations(annotations map[string]string, trigger string) []string {
	recipients := make([]string, 0)
	for k, annotation := range annotations {
		if !strings.HasSuffix(k, RecipientsAnnotation) {
			continue
		}
		if name := strings.TrimRight(k[0:len(k)-len(RecipientsAnnotation)], "."); name != "" && name != trigger {
			continue
		}

		for _, recipient := range strings.Split(annotation, ",") {
			if recipient = strings.TrimSpace(recipient); recipient != "" {
				recipients = append(recipients, recipient)
			}
		}
	}

	return recipients
}
