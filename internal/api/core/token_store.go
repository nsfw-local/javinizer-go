package core

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"
)

// VerificationToken represents a successful proxy test that can be used for save authorization
type VerificationToken struct {
	Token      string    `json:"token"`
	Scope      string    `json:"scope"`       // "global", "flaresolverr", or "profile:{name}"
	ConfigHash string    `json:"config_hash"` // Hash of config at test time
	ExpiresAt  time.Time `json:"expires_at"`
	CreatedAt  time.Time `json:"created_at"`
}

const TokenValidityDuration = 5 * time.Minute

// TokenStore manages verification tokens in-memory
type TokenStore struct {
	mu     sync.RWMutex
	tokens map[string]VerificationToken
}

// NewTokenStore creates a new token store with background cleanup
func NewTokenStore() *TokenStore {
	ts := &TokenStore{
		tokens: make(map[string]VerificationToken),
	}

	// Start background cleanup every 10 minutes
	go ts.backgroundCleanup()
	return ts
}

func (s *TokenStore) backgroundCleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.CleanupExpired()
	}
}

// Create generates a new verification token for the given scope and config hash
func (s *TokenStore) Create(scope string, configHash string) VerificationToken {
	token := generateToken()
	vt := VerificationToken{
		Token:      token,
		Scope:      scope,
		ConfigHash: configHash,
		ExpiresAt:  time.Now().Add(TokenValidityDuration),
		CreatedAt:  time.Now(),
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens[token] = vt
	return vt
}

// Validate checks if a token is valid for the given scope and config hash
func (s *TokenStore) Validate(token string, scope string, configHash string) bool {
	s.mu.RLock()
	vt, ok := s.tokens[token]
	if !ok {
		s.mu.RUnlock()
		return false
	}

	// Check expiration - let CleanupExpired handle deletion
	if vt.ExpiresAt.Before(time.Now()) {
		s.mu.RUnlock()
		return false
	}

	// Check scope
	if vt.Scope != scope {
		s.mu.RUnlock()
		return false
	}

	// Check config hash
	if vt.ConfigHash != configHash {
		s.mu.RUnlock()
		return false
	}

	s.mu.RUnlock()
	return true
}

// CleanupExpired removes expired tokens from the store
func (s *TokenStore) CleanupExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for token, vt := range s.tokens {
		if vt.ExpiresAt.Before(now) {
			delete(s.tokens, token)
		}
	}
}

func generateToken() string {
	bytes := make([]byte, 16)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// HashProxyConfig creates a hash of proxy config for comparison using SHA-256
func HashProxyConfig(proxyConfig interface{}) string {
	// Use canonical JSON for consistent hashing
	data, _ := json.Marshal(proxyConfig)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:8]) // First 8 bytes (16 hex chars) is sufficient
}
