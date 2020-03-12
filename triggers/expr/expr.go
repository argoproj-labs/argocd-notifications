package expr

import (
	"github.com/argoproj-labs/argocd-notifications/triggers/expr/repo"
	"github.com/argoproj-labs/argocd-notifications/triggers/expr/time"
)

var helpers = map[string]interface{}{}

func init() {
	helpers = make(map[string]interface{})
	register("time", time.NewExprs())
	register("repo", repo.NewExprs())
}

func register(namespace string, entry map[string]interface{}) {
	helpers[namespace] = entry
}

func Spawn() map[string]interface{} {
	clone := make(map[string]interface{})
	for namespace, helper := range helpers {
		clone[namespace] = helper
	}

	return clone
}
