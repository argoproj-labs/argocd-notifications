package bot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/argoproj-labs/argocd-notifications/pkg/controller"

	"github.com/argoproj-labs/argocd-notifications/shared/k8s"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

type Server interface {
	Serve(port int) error
	AddAdapter(path string, adapter Adapter)
}

func NewServer(dynamicClient dynamic.Interface, namespace string) *server {
	return &server{
		mux:           http.NewServeMux(),
		appClient:     k8s.NewAppClient(dynamicClient, namespace),
		appProjClient: k8s.NewAppProjClient(dynamicClient, namespace),
	}
}

type server struct {
	appClient     dynamic.ResourceInterface
	appProjClient dynamic.ResourceInterface
	mux           *http.ServeMux
}

func copyStringMap(in map[string]string) map[string]string {
	out := make(map[string]string)
	for k, v := range in {
		out[k] = v
	}
	return out
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
		return s.listSubscriptions(cmd.Service, cmd.Recipient)
	case cmd.Subscribe != nil:
		return s.updateSubscription(cmd.Service, cmd.Recipient, true, *cmd.Subscribe)
	case cmd.Unsubscribe != nil:
		return s.updateSubscription(cmd.Service, cmd.Recipient, false, *cmd.Unsubscribe)
	default:
		return "", errors.New("unknown command")
	}
}

func annotationsPatch(old map[string]string, new map[string]string) map[string]*string {
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

func (s *server) updateSubscription(service string, recipient string, subscribe bool, opts UpdateSubscription) (string, error) {
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
	obj, err := client.Get(context.Background(), name, v1.GetOptions{})
	if err != nil {
		return "", err
	}
	oldAnnotations := copyStringMap(obj.GetAnnotations())
	annotations := controller.Subscriptions(obj.GetAnnotations())
	if subscribe {
		annotations.Subscribe(opts.Trigger, service, recipient)
	} else {
		annotations.Unsubscribe(opts.Trigger, service, recipient)
	}
	annotationsPatch := annotationsPatch(oldAnnotations, annotations)
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
		_, err = client.Patch(context.Background(), name, types.MergePatchType, patchData, v1.PatchOptions{})
		if err != nil {
			return "", err
		}
	}

	return "subscription updated", nil
}

func (s *server) listSubscriptions(service string, recipient string) (string, error) {
	appList, err := s.appClient.List(context.Background(), v1.ListOptions{})
	if err != nil {
		return "", err
	}
	var apps []string
	for _, app := range appList.Items {
		if controller.Subscriptions(app.GetAnnotations()).Has(service, recipient) {
			apps = append(apps, fmt.Sprintf("%s/%s", app.GetNamespace(), app.GetName()))
		}
	}
	appProjList, err := s.appProjClient.List(context.Background(), v1.ListOptions{})
	if err != nil {
		return "", err
	}
	var appProjs []string
	for _, appProj := range appProjList.Items {
		if controller.Subscriptions(appProj.GetAnnotations()).Has(service, recipient) {
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

func (s *server) AddAdapter(pattern string, adapter Adapter) {
	s.mux.HandleFunc(pattern, s.handler(adapter))
}

func (s *server) Serve(port int) error {
	return http.ListenAndServe(fmt.Sprintf(":%d", port), s.mux)
}
