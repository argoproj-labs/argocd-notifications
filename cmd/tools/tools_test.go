package tools

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrintFormatterJson(t *testing.T) {
	var out bytes.Buffer
	err := printFormatted(map[string]string{
		"foo": "bar",
	}, "json", &out)

	assert.NoError(t, err)
	assert.Contains(t, out.String(), `{
  "foo": "bar"
}`)
}

func TestPrintFormatterYaml(t *testing.T) {
	var out bytes.Buffer
	err := printFormatted(map[string]string{
		"foo": "bar",
	}, "yaml", &out)

	assert.NoError(t, err)
	assert.Contains(t, out.String(), `foo: bar`)
}
