package sync

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func NewExprs() map[string]interface{} {
	return map[string]interface{}{
		"GetInfoItem": func(app map[string]interface{}, name string) string {
			res, err := getInfoItem(app, name)
			if err != nil {
				panic(err)
			}
			return res
		},
	}
}

func getInfoItem(app map[string]interface{}, name string) (string, error) {
	un := unstructured.Unstructured{Object: app}
	operation, ok, _ := unstructured.NestedMap(app, "operation")
	if !ok {
		operation, ok, _ = unstructured.NestedMap(app, "status", "operationState", "operation")
	}
	if !ok {
		return "", fmt.Errorf("application '%s' has no operation", un.GetName())
	}

	infoItems, ok := operation["info"].([]interface{})
	if !ok {
		return "", fmt.Errorf("application '%s' has no info items", un.GetName())
	}
	for _, infoItem := range infoItems {
		item, ok := infoItem.(map[string]interface{})
		if !ok {
			continue
		}
		if item["name"] == name {
			res, ok := item["value"].(string)
			if !ok {
				return "", fmt.Errorf("application '%s' has invalid value of info item '%s'", un.GetName(), name)
			}
			return res, nil
		}
	}
	return "", fmt.Errorf("application '%s' has no info item with name '%s'", un.GetName(), name)
}
