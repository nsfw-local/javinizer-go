package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveScraperProxy(t *testing.T) {
	global := ProxyConfig{
		Enabled:        true,
		DefaultProfile: "main",
		Profiles: map[string]ProxyProfile{
			"main": {
				URL:      "http://main-proxy.example.com:8080",
				Username: "main-user",
				Password: "main-pass",
			},
			"backup": {
				URL:      "http://backup-proxy.example.com:8080",
				Username: "backup-user",
				Password: "backup-pass",
			},
		},
	}

	t.Run("falls back to global proxy when override is nil", func(t *testing.T) {
		resolved := ResolveScraperProxy(global, nil)
		assert.Equal(t, "http://main-proxy.example.com:8080", resolved.URL)
		assert.Equal(t, "main-user", resolved.Username)
		assert.Equal(t, "main-pass", resolved.Password)
	})

	t.Run("falls back to global proxy when override is explicitly disabled", func(t *testing.T) {
		override := &ProxyConfig{
			Enabled: false,
		}
		resolved := ResolveScraperProxy(global, override)
		assert.Equal(t, "http://main-proxy.example.com:8080", resolved.URL)
		assert.Equal(t, "main-user", resolved.Username)
		assert.Equal(t, "main-pass", resolved.Password)
	})

	t.Run("nil override with disabled global returns empty config", func(t *testing.T) {
		disabledGlobal := ProxyConfig{
			Enabled: false,
		}
		resolved := ResolveScraperProxy(disabledGlobal, nil)
		assert.Equal(t, "", resolved.URL)
	})

	t.Run("disabled override with disabled global returns empty config", func(t *testing.T) {
		disabledGlobal := ProxyConfig{
			Enabled: false,
		}
		override := &ProxyConfig{
			Enabled: false,
		}
		resolved := ResolveScraperProxy(disabledGlobal, override)
		assert.Equal(t, "", resolved.URL)
	})

	t.Run("uses scraper override when profile specified", func(t *testing.T) {
		override := &ProxyConfig{
			Enabled: true,
			Profile: "backup",
		}

		resolved := ResolveScraperProxy(global, override)
		assert.Equal(t, "http://backup-proxy.example.com:8080", resolved.URL)
		assert.Equal(t, "backup-user", resolved.Username)
		assert.Equal(t, "backup-pass", resolved.Password)
	})

	t.Run("uses scraper profile when provided", func(t *testing.T) {
		override := &ProxyConfig{
			Enabled: true,
			Profile: "backup",
		}

		resolved := ResolveScraperProxy(global, override)
		assert.Equal(t, "http://backup-proxy.example.com:8080", resolved.URL)
		assert.Equal(t, "backup-user", resolved.Username)
		assert.Equal(t, "backup-pass", resolved.Password)
	})

	t.Run("inherits global proxy URL and credentials when enabled override omits URL", func(t *testing.T) {
		override := &ProxyConfig{
			Enabled: true,
			// URL/credentials intentionally omitted - will inherit from global
		}

		resolved := ResolveScraperProxy(global, override)
		assert.Equal(t, "http://main-proxy.example.com:8080", resolved.URL)
		assert.Equal(t, "main-user", resolved.Username)
		assert.Equal(t, "main-pass", resolved.Password)
	})
}

func TestResolveGlobalProxy(t *testing.T) {
	global := ProxyConfig{
		Enabled:        true,
		DefaultProfile: "main",
		Profiles: map[string]ProxyProfile{
			"main": {
				URL:      "http://main-proxy.example.com:8080",
				Username: "main-user",
				Password: "main-pass",
			},
		},
	}

	resolved := ResolveGlobalProxy(global)
	assert.Equal(t, "http://main-proxy.example.com:8080", resolved.URL)
	assert.Equal(t, "main-user", resolved.Username)
	assert.Equal(t, "main-pass", resolved.Password)
}
