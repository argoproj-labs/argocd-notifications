package controller

import (
	"strconv"
	"testing"

	"github.com/argoproj-labs/argocd-notifications/pkg/services"

	"github.com/stretchr/testify/assert"
)

func TestNotificationState_Truncate(t *testing.T) {
	state := notificationState{}
	for i := 0; i < 5; i++ {
		state[strconv.Itoa(i)] = int64(i)
	}

	state.truncate(3)

	assert.Equal(t, notificationState{"2": 2, "3": 3, "4": 4}, state)
}

func TestSetAlreadyNotified(t *testing.T) {
	dest := services.Destination{Service: "slack", Recipient: "my-channel"}

	state := notificationState{}
	changed := state.setAlreadyNotified("app-synced", "", dest, true)

	assert.True(t, changed)
	_, ok := state["app-synced:slack:my-channel"]
	assert.True(t, ok)

	changed = state.setAlreadyNotified("app-synced", "", dest, true)
	assert.False(t, changed)

	changed = state.setAlreadyNotified("app-synced", "", dest, false)
	assert.True(t, changed)
	_, ok = state["app-synced:slack:my-channel"]
	assert.False(t, ok)
}

func TestSetAlreadyNotified_OncePerItem(t *testing.T) {
	dest := services.Destination{Service: "slack", Recipient: "my-channel"}

	state := notificationState{}
	changed := state.setAlreadyNotified("app-synced", "abc", dest, true)

	assert.True(t, changed)
	_, ok := state["abc:app-synced:slack:my-channel"]
	assert.True(t, ok)

	changed = state.setAlreadyNotified("app-synced", "abc", dest, false)
	assert.False(t, changed)
	_, ok = state["abc:app-synced:slack:my-channel"]
	assert.True(t, ok)
}
