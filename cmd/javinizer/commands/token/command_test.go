package token

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCommand(t *testing.T) {
	cmd := NewCommand()
	assert.Equal(t, "token", cmd.Use)
	assert.NotNil(t, cmd)

	subCommands := cmd.Commands()
	subNames := make([]string, len(subCommands))
	for i, sc := range subCommands {
		subNames[i] = sc.Name()
	}
	assert.Contains(t, subNames, "create")
	assert.Contains(t, subNames, "revoke")
	assert.Contains(t, subNames, "list")
}

func TestShortID(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"long id truncated", "abcdef1234567890", "abcdef12"},
		{"short id unchanged", "abc", "abc"},
		{"exactly 8 chars", "12345678", "12345678"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, shortID(tc.input))
		})
	}
}

func TestPrintJSON(t *testing.T) {
	cmd := &cobra.Command{}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	data := map[string]string{"key": "value"}
	err := printJSON(cmd, data)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, `"key"`)
	assert.Contains(t, output, `"value"`)
}

func TestTokenCreateResult_Fields(t *testing.T) {
	result := tokenCreateResult{
		ID:          "test-id",
		Name:        "test-name",
		TokenPrefix: "abcd1234",
		Token:       "jv_abcd1234567890abcdef1234567890",
	}
	assert.Equal(t, "test-id", result.ID)
	assert.Equal(t, "test-name", result.Name)
	assert.Equal(t, "abcd1234", result.TokenPrefix)
	assert.True(t, strings.HasPrefix(result.Token, "jv_"))
}

func TestTokenRevokeResult_Fields(t *testing.T) {
	result := tokenRevokeResult{
		ID:      "test-id",
		Prefix:  "abcd1234",
		Revoked: true,
	}
	assert.Equal(t, "test-id", result.ID)
	assert.Equal(t, "abcd1234", result.Prefix)
	assert.True(t, result.Revoked)
}

func TestTokenListEntry_Fields(t *testing.T) {
	entry := tokenListEntry{
		ID:          "test-id",
		Name:        "test-name",
		TokenPrefix: "abcd1234",
	}
	assert.Equal(t, "test-id", entry.ID)
	assert.Equal(t, "test-name", entry.Name)
	assert.Equal(t, "abcd1234", entry.TokenPrefix)
}

func TestTokenListEntry_JSONSerialization(t *testing.T) {
	entry := tokenListEntry{
		ID:          "test-id",
		Name:        "test-name",
		TokenPrefix: "abcd1234",
	}
	data, err := json.Marshal(entry)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"id"`)
	assert.Contains(t, string(data), `"name"`)
	assert.Contains(t, string(data), `"token_prefix"`)
}
