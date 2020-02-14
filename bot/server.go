package bot

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/argoproj-labs/argocd-notifications/shared/clients"
	"github.com/argoproj-labs/argocd-notifications/shared/recipients"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

type Server interface {
	Serve(port int) error
}

func NewServer(dynamicClient dynamic.Interface, namespace string) *server {
	return &server{
		appClient:     clients.NewAppClient(dynamicClient, namespace),
		appProjClient: clients.NewAppProjClient(dynamicClient, namespace),
	}
}

type server struct {
	appClient     dynamic.ResourceInterface
	appProjClient dynamic.ResourceInterface
}

func (s *server) handler(adapter Adapter) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		cmd, err := adapter.Parse(r)
		if err != nil {
			adapter.SendResponse(err.Error(), w)
			return
		}
		if res, err := s.execute(cmd); err != nil {
			adapter.SendResponse(fmt.Sprintf("cannot execute command: %v", err), w)
		} else {
			adapter.SendResponse(res, w)
		}
	}
}

func (s *server) execute(cmd Command) (string, error) {
	switch {
	case cmd.ListSubscriptions != nil:
		return s.listSubscriptions(cmd.Recipient)
	case cmd.Subscribe != nil:
		return s.updateSubscription(cmd.Recipient, true, *cmd.Subscribe)
	case cmd.Unsubscribe != nil:
		return s.updateSubscription(cmd.Recipient, false, *cmd.Unsubscribe)
	default:
		return "", errors.New("unknown command")
	}
}

func findStringIndex(items []string, item string) int {
	for i := range items {
		if items[i] == item {
			return i
		}
	}
	return -1
}

func addSubscription(recipient string, trigger string, annotations map[string]string) map[string]string {
	annotations = recipients.CopyStringMap(annotations)
	annotationKey := recipients.RecipientsAnnotation
	if trigger != "" {
		annotationKey = fmt.Sprintf("%s.%s", trigger, annotationKey)
	}
	existingRecipients := recipients.ParseRecipients(annotations[annotationKey])
	if index := findStringIndex(existingRecipients, recipient); index < 0 {
		annotations[annotationKey] = strings.Join(append(existingRecipients, recipient), ",")
	}
	return annotations
}

func removeSubscription(recipient string, trigger string, annotations map[string]string) map[string]string {
	annotations = recipients.CopyStringMap(annotations)
	for _, k := range recipients.GetAnnotationKeys(annotations, trigger) {
		existingRecipients := recipients.ParseRecipients(annotations[k])
		if index := findStringIndex(existingRecipients, recipient); index > -1 {
			newRecipients := append(existingRecipients[:index], existingRecipients[index+1:]...)
			if len(newRecipients) > 0 {
				annotations[k] = strings.Join(newRecipients, ",")
			} else {
				delete(annotations, k)
			}
		}
	}
	return annotations
}

func (s *server) updateSubscription(recipient string, subscribe bool, opts UpdateSubscription) (string, error) {
	var name string
	var client dynamic.ResourceInterface
	switch {
	case opts.App != "":
		name = opts.App
		client = s.appClient
	case opts.Project != "":
		name = opts.Project
		client = s.appProjClient
	default:
		return "", errors.New("either application or project name must be specified")
	}
	obj, err := client.Get(name, v1.GetOptions{})
	if err != nil {
		return "", err
	}
	oldAnnotations := recipients.CopyStringMap(obj.GetAnnotations())
	var newAnnotations map[string]string
	if subscribe {
		newAnnotations = addSubscription(recipient, opts.Trigger, obj.GetAnnotations())
	} else {
		newAnnotations = removeSubscription(recipient, opts.Trigger, obj.GetAnnotations())
	}
	annotationsPatch := recipients.AnnotationsPatch(oldAnnotations, newAnnotations)
	if len(annotationsPatch) > 0 {
		patch := map[string]map[string]interface{}{
			"metadata": {
				"annotations": annotationsPatch,
			},
		}
		patchData, err := json.Marshal(patch)
		if err != nil {
			return "", err
		}
		_, err = client.Patch(name, types.MergePatchType, patchData, v1.PatchOptions{})
		if err != nil {
			return "", err
		}
	}

	return "subscription updated", nil
}

func (s *server) listSubscriptions(recipient string) (string, error) {
	appList, err := s.appClient.List(v1.ListOptions{})
	if err != nil {
		return "", err
	}
	var apps []string
	for _, app := range appList.Items {
		if findStringIndex(recipients.GetRecipientsFromAnnotations(app.GetAnnotations(), ""), recipient) > -1 {
			apps = append(apps, fmt.Sprintf("%s/%s", app.GetNamespace(), app.GetName()))
		}
	}
	appProjList, err := s.appProjClient.List(v1.ListOptions{})
	if err != nil {
		return "", err
	}
	var appProjs []string
	for _, appProj := range appProjList.Items {
		if findStringIndex(recipients.GetRecipientsFromAnnotations(appProj.GetAnnotations(), ""), recipient) > -1 {
			appProjs = append(appProjs, fmt.Sprintf("%s/%s", appProj.GetNamespace(), appProj.GetName()))
		}
	}
	response := fmt.Sprintf("The %s has no subscriptions.", recipient)
	if len(apps) > 0 || len(appProjs) > 0 {
		response = fmt.Sprintf("The %s is subscribed to %d applications and %d projects.",
			recipient, len(apps), len(appProjs))
		if len(apps) > 0 {
			response = fmt.Sprintf("%s\nApplications: %s.", response, strings.Join(apps, ", "))
		}
		if len(appProjs) > 0 {
			response = fmt.Sprintf("%s\nProjects: %s.", response, strings.Join(appProjs, ", "))
		}
	}
	return response, nil
}

func (s *server) Serve(port int) error {
	http.HandleFunc("/slack", s.handler(&slack{}))
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
