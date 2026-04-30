package token

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateToken(t *testing.T) {
	testCases := []struct {
		name     string
		assertFn func(t *testing.T, fullToken string, prefix string)
	}{
		{
			name: "has jv_ prefix",
			assertFn: func(t *testing.T, fullToken string, prefix string) {
				assert.True(t, strings.HasPrefix(fullToken, TokenPrefix), "token should start with %s", TokenPrefix)
			},
		},
		{
			name: "total length is prefix plus 32 hex chars",
			assertFn: func(t *testing.T, fullToken string, prefix string) {
				expectedLen := len(TokenPrefix) + 32
				assert.Len(t, fullToken, expectedLen, "token should be %d chars", expectedLen)
			},
		},
		{
			name: "prefix is 8 hex characters",
			assertFn: func(t *testing.T, fullToken string, prefix string) {
				assert.Len(t, prefix, 8, "prefix should be 8 chars")
				for _, c := range prefix {
					assert.True(t, (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f'),
						"prefix char %q should be hex", c)
				}
			},
		},
		{
			name: "uniqueness across calls",
			assertFn: func(t *testing.T, fullToken string, prefix string) {
				otherFull, otherPrefix, err := GenerateToken()
				require.NoError(t, err)
				assert.NotEqual(t, fullToken, otherFull, "consecutive tokens should differ")
				assert.NotEqual(t, prefix, otherPrefix, "consecutive prefixes should differ")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fullToken, prefix, err := GenerateToken()
			require.NoError(t, err)
			tc.assertFn(t, fullToken, prefix)
		})
	}
}

func TestHashToken(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "deterministic SHA-256",
			input:    "jv_abcdef1234567890abcdef1234567890",
			expected: "353092f043bfa9facd7dcc21496a91028f1dc70e260edbde8f23f504cfa80ab5",
		},
		{
			name:     "empty string produces valid hash",
			input:    "",
			expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := HashToken(tc.input)
			assert.Len(t, got, 64, "SHA-256 hex should be 64 chars")
			assert.Equal(t, tc.expected, got)
		})
	}

	t.Run("same input produces same hash", func(t *testing.T) {
		input := "jv_test1234567890abcdef"
		first := HashToken(input)
		second := HashToken(input)
		assert.Equal(t, first, second)
	})

	t.Run("different inputs produce different hashes", func(t *testing.T) {
		h1 := HashToken("jv_token_a")
		h2 := HashToken("jv_token_b")
		assert.NotEqual(t, h1, h2)
	})
}

func TestGenerateToken_HashRoundTrip(t *testing.T) {
	fullToken, _, err := GenerateToken()
	require.NoError(t, err)

	directHash := HashToken(fullToken)
	assert.Len(t, directHash, 64, "SHA-256 hex should be 64 chars")
	assert.NotEmpty(t, directHash)

	secondHash := HashToken(fullToken)
	assert.Equal(t, directHash, secondHash, "hashing the same token twice should produce identical results")
}

func TestGenerateToken_PrefixMatchesHexSubstring(t *testing.T) {
	fullToken, prefix, err := GenerateToken()
	require.NoError(t, err)

	hexPart := fullToken[len(TokenPrefix):]
	assert.Equal(t, hexPart[:8], prefix, "prefix should equal the first 8 hex chars after jv_")
}
