package misc

import (
	"encoding/json"
	"fmt"
	"io"

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
