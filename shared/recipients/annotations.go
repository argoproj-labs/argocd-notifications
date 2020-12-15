package recipients

import (
	"strings"

	"github.com/argoproj-labs/argocd-notifications/pkg/shared/text"
)

const (
	AnnotationPostfix = "argocd-notifications.argoproj.io"
)

var (
	RecipientsAnnotation = "recipients." + AnnotationPostfix
)

func GetRecipientsFromAnnotations(annotations map[string]string, trigger string) []string {
	recipients := make([]string, 0)
	for _, k := range GetAnnotationKeys(annotations, trigger) {
		annotation := annotations[k]
		recipients = append(recipients, ParseRecipients(annotation)...)
	}

	return recipients
}

func GetAnnotationKeys(annotations map[string]string, trigger string) []string {
	keys := make([]string, 0)
	for k := range annotations {
		if !strings.HasSuffix(k, RecipientsAnnotation) {
			continue
		}
		if name := strings.TrimRight(k[0:len(k)-len(RecipientsAnnotation)], "."); name != "" && name != trigger {
			continue
		}
		keys = append(keys, k)
	}
	return keys
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

func CopyStringMap(in map[string]string) map[string]string {
	out := make(map[string]string)
	for k, v := range in {
		out[k] = v
	}
	return out
}

func AnnotationsPatch(old map[string]string, new map[string]string) map[string]*string {
	patch := map[string]*string{}
	for k := range new {
		val := new[k]
		if val != old[k] {
			patch[k] = &val
		}
	}
	for k := range old {
		if _, ok := new[k]; !ok {
			patch[k] = nil
		}
	}
	return patch
}
