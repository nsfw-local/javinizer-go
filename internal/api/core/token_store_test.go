package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTokenStore_CreateAndValidate(t *testing.T) {
	store := NewTokenStore()

	vt := store.Create("global", "hash123")
	assert.NotEmpty(t, vt.Token)
	assert.Equal(t, "global", vt.Scope)
	assert.Equal(t, "hash123", vt.ConfigHash)
	assert.True(t, vt.ExpiresAt.After(time.Now()))

	// Valid token should pass
	assert.True(t, store.Validate(vt.Token, "global", "hash123"))

	// Wrong scope should fail
	assert.False(t, store.Validate(vt.Token, "flaresolverr", "hash123"))

	// Wrong hash should fail
	assert.False(t, store.Validate(vt.Token, "global", "different_hash"))

	// Non-existent token should fail
	assert.False(t, store.Validate("invalid_token", "global", "hash123"))
}

func TestTokenStore_ExpiredToken(t *testing.T) {
	store := NewTokenStore()

	vt := VerificationToken{
		Token:      "expired_token",
		Scope:      "global",
		ConfigHash: "hash123",
		ExpiresAt:  time.Now().Add(-1 * time.Minute),
		CreatedAt:  time.Now().Add(-10 * time.Minute),
	}

	store.tokens["expired_token"] = vt

	// Expired token should fail
	assert.False(t, store.Validate("expired_token", "global", "hash123"))
}

func TestTokenStore_CleanupExpired(t *testing.T) {
	store := NewTokenStore()

	store.tokens["expired"] = VerificationToken{
		Token:     "expired",
		ExpiresAt: time.Now().Add(-1 * time.Minute),
	}
	store.tokens["valid"] = VerificationToken{
		Token:     "valid",
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}

	store.CleanupExpired()

	_, hasExpired := store.tokens["expired"]
	_, hasValid := store.tokens["valid"]

	assert.False(t, hasExpired)
	assert.True(t, hasValid)
}

func TestHashProxyConfig(t *testing.T) {
	// Same config should produce same hash
	hash1 := HashProxyConfig("test-config")
	hash2 := HashProxyConfig("test-config")
	assert.Equal(t, hash1, hash2)

	// Different config should produce different hash
	hash3 := HashProxyConfig("different-config")
	assert.NotEqual(t, hash1, hash3)
}
