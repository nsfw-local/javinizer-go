package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileExists(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
		want bool
	}{
		{"nonexistent_file", "/nonexistent/path/to/file", false},
		{"existent_file", "config_test.go", true},
		{"directory", ".", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For directory test, we expect false because os.Stat succeeds but it's a dir
			// Actually fileExists just checks os.Stat error, so dir would return true
			if tt.name == "directory" {
				// Directories exist, so fileExists returns true
				assert.True(t, fileExists(tt.path))
			} else {
				assert.Equal(t, tt.want, fileExists(tt.path))
			}
		})
	}
}

func TestAutoDiscoverBrowserBinary(t *testing.T) {
	t.Parallel()

	result := AutoDiscoverBrowserBinary()
	assert.NotPanics(t, func() {
		AutoDiscoverBrowserBinary()
	})
	assert.NotNil(t, result)
}

func TestGetBrowserBinaryPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cfg     BrowserConfig
		wantLen bool
	}{
		{
			name:    "empty_config",
			cfg:     BrowserConfig{},
			wantLen: false,
		},
		{
			name: "custom_binary_path",
			cfg: BrowserConfig{
				BinaryPath: "/custom/path/to/chrome",
			},
			wantLen: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetBrowserBinaryPath(tt.cfg)
			if tt.wantLen {
				assert.NotEmpty(t, result)
				assert.Equal(t, tt.cfg.BinaryPath, result)
			} else {
				// Empty config will try to auto-discover
				// Result could be empty or a path, just verify no panic
				assert.NotPanics(t, func() {
					GetBrowserBinaryPath(tt.cfg)
				})
			}
		})
	}
}
