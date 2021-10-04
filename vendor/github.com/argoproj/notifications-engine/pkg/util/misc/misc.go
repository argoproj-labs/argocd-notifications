package misc

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"

	"github.com/ghodss/yaml"
)

func PrintFormatted(input interface{}, output string, out io.Writer) error {
	switch output {
	case "json":
		data, err := json.MarshalIndent(input, "", "  ")
		if err != nil {
			return err
		}
		_, err = out.Write([]byte(string(data) + "\n"))
		return err
	case "yaml":
		data, err := yaml.Marshal(input)
		if err != nil {
			return err
		}
		_, err = out.Write(data)
		return err
	default:
		return fmt.Errorf("output '%s' is not supported", output)
	}
}

func IterateStringKeyMap(val interface{}, callback func(key string)) {
	keys := reflect.ValueOf(val).MapKeys()
	var sortedKeys []string
	for _, k := range keys {
		sortedKeys = append(sortedKeys, k.String())
	}
	sort.Slice(sortedKeys, func(i, j int) bool {
		return sortedKeys[i] < sortedKeys[j]
	})
	for i := range sortedKeys {
		callback(sortedKeys[i])
	}
}
