package bot

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/argoproj-labs/argocd-notifications/shared/clients"
	"github.com/argoproj-labs/argocd-notifications/shared/recipients"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		return s.listSubscriptions(cmd.ListSubscriptions.Channel)
	default:
		return "", errors.New("unknown command")
	}
}

func sliceHasString(items []string, item string) bool {
	for i := range items {
		if items[i] == item {
			return true
		}
	}
	return false
}

func (s *server) listSubscriptions(receiver string) (string, error) {
	appList, err := s.appClient.List(v1.ListOptions{})
	if err != nil {
		return "", err
	}
	var apps []string
	for _, app := range appList.Items {
		if sliceHasString(recipients.GetRecipientsFromAnnotations(app.GetAnnotations(), ""), receiver) {
			apps = append(apps, fmt.Sprintf("%s/%s", app.GetNamespace(), app.GetName()))
		}
	}
	appProjList, err := s.appProjClient.List(v1.ListOptions{})
	if err != nil {
		return "", err
	}
	var appProjs []string
	for _, appProj := range appProjList.Items {
		if sliceHasString(recipients.GetRecipientsFromAnnotations(appProj.GetAnnotations(), ""), receiver) {
			appProjs = append(appProjs, fmt.Sprintf("%s/%s", appProj.GetNamespace(), appProj.GetName()))
		}
	}
	response := fmt.Sprintf("The %s has no subscriptions.", receiver)
	if len(apps) > 0 || len(appProjs) > 0 {
		response = fmt.Sprintf("The %s is subscribed to %d applications and %d projects.",
			receiver, len(apps), len(appProjs))
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
