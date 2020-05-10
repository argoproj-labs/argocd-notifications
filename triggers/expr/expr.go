package expr

import (
	"github.com/argoproj-labs/argocd-notifications/shared/argocd"
	"github.com/argoproj-labs/argocd-notifications/triggers/expr/repo"
	"github.com/argoproj-labs/argocd-notifications/triggers/expr/time"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var helpers = map[string]interface{}{}

func init() {
	helpers = make(map[string]interface{})
	register("time", time.NewExprs())
}

func register(namespace string, entry map[string]interface{}) {
	helpers[namespace] = entry
}

func Spawn(app *unstructured.Unstructured, argocdService argocd.Service) map[string]interface{} {
	clone := make(map[string]interface{})
	for namespace, helper := range helpers {
		clone[namespace] = helper
	}
	clone["repo"] = repo.NewExprs(argocdService, app)

	return clone
}
