package triggers

//go:generate mockgen -destination=./mocks/triggers.go -package=mocks github.com/argoproj-labs/argocd-notifications/triggers Trigger

import (
	"fmt"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	exprHelpers "github.com/argoproj-labs/argocd-notifications/expr"
	"github.com/argoproj-labs/argocd-notifications/shared/argocd"
)

type NotificationTrigger struct {
	Name        string `json:"name,omitempty"`
	Condition   string `json:"condition,omitempty"`
	Description string `json:"description,omitempty"`
	Template    string `json:"template,omitempty"`
	Enabled     *bool  `json:"enabled,omitempty"`
}

func NewTrigger(t NotificationTrigger, argocdService argocd.Service) (Trigger, error) {
	if t.Condition == "" {
		return nil, fmt.Errorf("trigger '%s' condition is empty", t.Name)
	}
	condition, err := expr.Compile(t.Condition)
	if err != nil {
		return nil, fmt.Errorf("failed to parse trigger '%s' condition: %v", t.Name, err)
	}

	return &trigger{condition: condition, argocdService: argocdService}, nil
}

type Trigger interface {
	Triggered(app *unstructured.Unstructured) (bool, error)
	GetTemplate() string
}

type trigger struct {
	condition     *vm.Program
	argocdService argocd.Service
	template      string
}

func (t *trigger) Triggered(app *unstructured.Unstructured) (bool, error) {
	envs := map[string]interface{}{"app": app.Object}
	if res, err := expr.Run(t.condition, exprHelpers.Spawn(app, t.argocdService, envs)); err != nil {
		return false, err
	} else if boolRes, ok := res.(bool); ok {
		return boolRes, nil
	}
	return false, nil
}

func (t *trigger) GetTemplate() string {
	return t.template
}
