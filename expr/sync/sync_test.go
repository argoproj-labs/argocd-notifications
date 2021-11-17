package sync

import (
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/argoproj-labs/argocd-notifications/testing"
)

var (
	infoItems = []interface{}{
		map[string]interface{}{
			"name":  "name1",
			"value": "val1",
		},
	}
)

func TestGetInfoItem_RequestedOperation(t *testing.T) {
	app := NewApp("test")
	app.Object["operation"] = map[string]interface{}{
		"info": infoItems,
	}

	val, err := getInfoItem(app.Object, "name1")
	assert.NoError(t, err)
	assert.Equal(t, "val1", val)
}

func TestGetInfoItem_CompletedOperation(t *testing.T) {
	app := NewApp("test")
	app.Object["status"] = map[string]interface{}{
		"operationState": map[string]interface{}{
			"operation": map[string]interface{}{
				"info": infoItems,
			},
		},
	}

	val, err := getInfoItem(app.Object, "name1")
	assert.NoError(t, err)
	assert.Equal(t, "val1", val)
}

func TestGetInfoItem_NoOperation(t *testing.T) {
	app := NewApp("test")
	_, err := getInfoItem(app.Object, "name1")
	assert.Error(t, err)
}
