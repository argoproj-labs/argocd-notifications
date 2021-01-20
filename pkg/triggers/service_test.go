package triggers

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	svc, err := NewService(map[string][]Condition{
		"my-trigger": {{
			When: "var1 == 'abc'",
			Send: []string{"my-template"},
		}},
	})

	if !assert.NoError(t, err) {
		return
	}

	conditionKey := fmt.Sprintf("[0].%s", hash("var1 == 'abc'"))

	t.Run("Triggered", func(t *testing.T) {
		res, err := svc.Run("my-trigger", map[string]interface{}{"var1": "abc"})
		if assert.NoError(t, err) {
			return
		}
		assert.Equal(t, []ConditionResult{{
			Key:       conditionKey,
			Triggered: true,
			Templates: []string{"my-template"},
		}}, res)
	})

	t.Run("NotTriggered", func(t *testing.T) {
		res, err := svc.Run("my-trigger", map[string]interface{}{"var1": "bcd"})
		if assert.NoError(t, err) {
			return
		}
		assert.Equal(t, []ConditionResult{{
			Key:       conditionKey,
			Triggered: false,
			Templates: []string{"my-template"},
		}}, res)
	})

	t.Run("Failed", func(t *testing.T) {
		res, err := svc.Run("my-trigger", map[string]interface{}{})
		if assert.NoError(t, err) {
			return
		}
		assert.Equal(t, []ConditionResult{{
			Key:       conditionKey,
			Triggered: false,
			Templates: []string{"my-template"},
		}}, res)
	})
}

func TestRun_OncePerSet(t *testing.T) {
	revision := "123"
	svc, err := NewService(map[string][]Condition{
		"my-trigger": {{
			When:    "var1 == 'abc'",
			Send:    []string{"my-template"},
			OncePer: "revision",
		}},
	})

	if !assert.NoError(t, err) {
		return
	}

	conditionKey := fmt.Sprintf("%s:[0].%s", revision, hash("var1 == 'abc'"))

	t.Run("Triggered", func(t *testing.T) {
		res, err := svc.Run("my-trigger", map[string]interface{}{"var1": "abc", "revision": "123"})
		if assert.NoError(t, err) {
			return
		}
		assert.Equal(t, []ConditionResult{{
			Key:       conditionKey,
			Triggered: true,
			Templates: []string{"my-template"},
			OncePer:   revision,
		}}, res)
	})

	t.Run("NotTriggered", func(t *testing.T) {
		res, err := svc.Run("my-trigger", map[string]interface{}{"var1": "bcd"})
		if assert.NoError(t, err) {
			return
		}
		assert.Equal(t, []ConditionResult{{
			Key:       conditionKey,
			Triggered: false,
			Templates: []string{"my-template"},
			OncePer:   revision,
		}}, res)
	})
}
