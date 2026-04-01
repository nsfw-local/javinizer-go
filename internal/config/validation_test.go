package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestRejectUnknownProxyFields(t *testing.T) {
	tests := []struct {
		name      string
		yamlInput string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid proxy config with profile",
			yamlInput: "enabled: true\nprofile: main",
			wantErr:   false,
		},
		{
			name:      "legacy url field rejected",
			yamlInput: "enabled: true\nurl: http://proxy.example.com",
			wantErr:   true,
			errMsg:    "field 'url' is no longer supported",
		},
		{
			name:      "legacy username field rejected",
			yamlInput: "enabled: true\nusername: user",
			wantErr:   true,
			errMsg:    "field 'username' is no longer supported",
		},
		{
			name:      "legacy password field rejected",
			yamlInput: "enabled: true\npassword: secret",
			wantErr:   true,
			errMsg:    "field 'password' is no longer supported",
		},
		{
			name:      "legacy use_main_proxy field rejected",
			yamlInput: "enabled: true\nuse_main_proxy: true",
			wantErr:   true,
			errMsg:    "field 'use_main_proxy' is no longer supported",
		},
		{
			name:      "empty proxy config is valid",
			yamlInput: "{}",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var node yaml.Node
			err := yaml.Unmarshal([]byte(tt.yamlInput), &node)
			require.NoError(t, err)

			var targetNode *yaml.Node
			if len(node.Content) > 0 {
				targetNode = node.Content[0]
			}

			err = rejectUnknownProxyFields(targetNode, "test.proxy")
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
