package expr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpr(t *testing.T) {
	namespaces := []string{
		"time",
		"repo",
	}

	for _, ns := range namespaces {
		helpers := Spawn(nil, nil)
		_, hasNamespace := helpers[ns]
		assert.True(t, hasNamespace)
	}
}
