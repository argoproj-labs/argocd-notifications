package triggers

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
)

// Condition holds expression and template that must be used to create notification is expression is returns true
type Condition struct {
	OncePer     string   `json:"oncePer,omitempty"`
	When        string   `json:"when,omitempty"`
	Description string   `json:"description,omitempty"`
	Send        []string `json:"send,omitempty"`
}

type ConditionResult struct {
	Key       string
	OncePer   string
	Templates []string
	Triggered bool
}

type Service interface {
	// Executes given trigger name and return result of each condition
	Run(triggerName string, vars map[string]interface{}) ([]ConditionResult, error)
}

type service struct {
	compiledConditions map[string]*vm.Program
	triggers           map[string][]Condition
}

func NewService(triggers map[string][]Condition) (*service, error) {
	svc := service{map[string]*vm.Program{}, triggers}
	for _, t := range triggers {
		for _, condition := range t {
			prog, err := expr.Compile(condition.When)
			if err != nil {
				return nil, err
			}
			svc.compiledConditions[condition.When] = prog
		}
	}
	return &svc, nil
}

func hash(input string) string {
	h := sha1.New()
	_, _ = h.Write([]byte(input))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

func (svc *service) Run(triggerName string, vars map[string]interface{}) ([]ConditionResult, error) {
	t, ok := svc.triggers[triggerName]
	if !ok {
		return nil, fmt.Errorf("trigger '%s' is not configured", triggerName)
	}
	var res []ConditionResult
	for i, condition := range t {
		prog, ok := svc.compiledConditions[condition.When]
		if !ok {
			return nil, fmt.Errorf("trigger configiration has changed after initialization")
		}
		conditionResult := ConditionResult{
			Templates: condition.Send,
			Key:       fmt.Sprintf("[%d].%s", i, hash(condition.When)),
		}
		// ignore execution error and treat and false result
		val, err := expr.Run(prog, vars)
		if err == nil {
			boolRes, ok := val.(bool)
			conditionResult.Triggered = ok && boolRes
		}

		if condition.OncePer != "" {
			if oncePer, ok, err := unstructured.NestedFieldNoCopy(vars, strings.Split(condition.OncePer, ".")...); err == nil && ok {
				conditionResult.OncePer = fmt.Sprintf("%v", oncePer)
			}
		}
		res = append(res, conditionResult)
	}

	return res, nil
}
