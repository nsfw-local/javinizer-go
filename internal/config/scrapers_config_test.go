package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToInt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     any
		wantValue int
		wantOK    bool
	}{
		{"int", int(42), 42, true},
		{"int8", int8(8), 8, true},
		{"int16", int16(16), 16, true},
		{"int32", int32(32), 32, true},
		{"int64", int64(64), 64, true},
		{"uint", uint(42), 42, true},
		{"uint8", uint8(8), 8, true},
		{"uint16", uint16(16), 16, true},
		{"uint32", uint32(32), 32, true},
		{"uint64", uint64(64), 64, true},
		{"float32_whole", float32(42.0), 42, true},
		{"float64_whole", float64(42.0), 42, true},
		{"float32_fractional", float32(42.5), 0, false},
		{"float64_fractional", float64(42.5), 0, false},
		{"string", "42", 0, false},
		{"nil", nil, 0, false},
		{"bool", true, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValue, gotOK := toInt(tt.input)
			assert.Equal(t, tt.wantValue, gotValue)
			assert.Equal(t, tt.wantOK, gotOK)
		})
	}
}
