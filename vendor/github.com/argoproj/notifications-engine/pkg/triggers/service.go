package triggers

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"

	"github.com/argoproj/notifications-engine/pkg/util/text"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	log "github.com/sirupsen/logrus"
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
	compiledOncePer    map[string]*vm.Program
	triggers           map[string][]Condition
}

func NewService(triggers map[string][]Condition) (*service, error) {
	svc := service{
		compiledConditions: map[string]*vm.Program{},
		compiledOncePer:    map[string]*vm.Program{},
		triggers:           triggers,
	}
	for _, t := range triggers {
		for _, condition := range t {
			prog, err := expr.Compile(text.Coalesce(condition.When, "false"))
			if err != nil {
				return nil, err
			}
			svc.compiledConditions[condition.When] = prog

			if condition.OncePer != "" {
				prog, err := expr.Compile(condition.OncePer)
				if err != nil {
					return nil, err
				}
				svc.compiledOncePer[condition.OncePer] = prog
			}
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
		conditionResult := ConditionResult{
			Templates: condition.Send,
			Key:       fmt.Sprintf("[%d].%s", i, hash(condition.When)),
		}

		if prog, ok := svc.compiledConditions[condition.When]; !ok {
			return nil, fmt.Errorf("trigger configiration has changed after initialization")
		} else if val, err := expr.Run(prog, vars); err == nil {
			boolRes, ok := val.(bool)
			conditionResult.Triggered = ok && boolRes
		} else {
			log.Errorf("failed to execute when condition: %+v", err)
		}

		if prog, ok := svc.compiledOncePer[condition.OncePer]; ok {
			if val, err := expr.Run(prog, vars); err == nil {
				conditionResult.OncePer = fmt.Sprintf("%v", val)
			} else {
				log.Errorf("failed to execute oncePer condition: %+v", err)
			}
		}

		res = append(res, conditionResult)
	}

	return res, nil
}
