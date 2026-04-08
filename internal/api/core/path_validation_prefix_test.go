package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathHasPrefix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		path   string
		prefix string
		want   bool
	}{
		{"exact_match", "/dev", "/dev", true},
		{"has_prefix", "/dev/null", "/dev", true},
		{"no_match", "/home/user", "/dev", false},
		{"empty_prefix", "/dev/null", "", true},
		{"empty_path", "", "/dev", false},
		{"trailing_slash", "/dev/", "/dev", true},
		{"longer_prefix_than_path", "/dev", "/dev/null", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pathHasPrefix(tt.path, tt.prefix)
			assert.Equal(t, tt.want, got)
		})
	}
}
