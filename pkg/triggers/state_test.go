package triggers

import (
	"strconv"
	"testing"

	"github.com/argoproj-labs/argocd-notifications/pkg/services"

	"github.com/stretchr/testify/assert"
)

func TestNotificationState_Truncate(t *testing.T) {
	state := State{}
	for i := 0; i < 5; i++ {
		state[strconv.Itoa(i)] = int64(i)
	}

	state.Truncate(3)

	assert.Equal(t, State{"2": 2, "3": 3, "4": 4}, state)
}

func TestSetAlreadyNotified(t *testing.T) {
	dest := services.Destination{Service: "slack", Recipient: "my-channel"}

	state := State{}
	changed := state.SetAlreadyNotified("app-synced", ConditionResult{Key: "0"}, dest, true)

	assert.True(t, changed)
	_, ok := state["app-synced:0:slack:my-channel"]
	assert.True(t, ok)

	changed = state.SetAlreadyNotified("app-synced", ConditionResult{Key: "0"}, dest, true)
	assert.False(t, changed)

	changed = state.SetAlreadyNotified("app-synced", ConditionResult{Key: "0"}, dest, false)
	assert.True(t, changed)
	_, ok = state["app-synced:0:slack:my-channel"]
	assert.False(t, ok)
}

func TestSetAlreadyNotified_OncePerItem(t *testing.T) {
	dest := services.Destination{Service: "slack", Recipient: "my-channel"}

	state := State{}
	changed := state.SetAlreadyNotified("app-synced", ConditionResult{OncePer: "abc", Key: "0"}, dest, true)

	assert.True(t, changed)
	_, ok := state["abc:app-synced:0:slack:my-channel"]
	assert.True(t, ok)

	changed = state.SetAlreadyNotified("app-synced", ConditionResult{OncePer: "abc", Key: "0"}, dest, false)
	assert.False(t, changed)
	_, ok = state["abc:app-synced:0:slack:my-channel"]
	assert.True(t, ok)
}
