package history

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTruncatePath(t *testing.T) {
	testCases := []struct {
		name     string
		path     string
		maxLen   int
		expected string
	}{
		{"short path unchanged", "/foo/bar", 20, "/foo/bar"},
		{"long path truncated", "/very/long/path/that/exceeds/max", 15, ".../exceeds/max"},
		{"exact length unchanged", "/foo", 4, "/foo"},
		{"empty path", "", 10, ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, truncatePath(tc.path, tc.maxLen))
		})
	}
}

func TestPercentage(t *testing.T) {
	testCases := []struct {
		name     string
		part     int64
		total    int64
		expected float64
	}{
		{"fifty percent", 50, 100, 50.0},
		{"zero total", 10, 0, 0},
		{"zero part", 0, 100, 0},
		{"full", 100, 100, 100.0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, percentage(tc.part, tc.total))
		})
	}
}
