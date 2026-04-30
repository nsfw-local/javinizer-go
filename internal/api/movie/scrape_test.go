package movie

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScraperPanicError(t *testing.T) {
	err := &scraperPanicError{message: "test panic"}
	assert.Equal(t, "test panic", err.Error())
}

func TestIsScraperPanicError(t *testing.T) {
	t.Run("true for scraperPanicError", func(t *testing.T) {
		assert.True(t, isScraperPanicError(&scraperPanicError{message: "panic"}))
	})

	t.Run("false for other error", func(t *testing.T) {
		assert.False(t, isScraperPanicError(errors.New("other")))
	})
}
