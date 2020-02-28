package notifiers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidIconEmoij(t *testing.T) {
	assert.Equal(t, true, validIconEmoij.MatchString(":slack:"))
	assert.Equal(t, true, validIconEmoij.MatchString(":chart_with_upwards_trend:"))
	assert.Equal(t, false, validIconEmoij.MatchString("http://lorempixel.com/48/48"))
}

func TestValidIconURL(t *testing.T) {
	assert.Equal(t, true, isValidIconURL("http://lorempixel.com/48/48"))
	assert.Equal(t, true, isValidIconURL("https://lorempixel.com/48/48"))
	assert.Equal(t, false, isValidIconURL("favicon.ico"))
	assert.Equal(t, false, isValidIconURL("ftp://favicon.ico"))
	assert.Equal(t, false, isValidIconURL("ftp://lorempixel.com/favicon.ico"))
}
