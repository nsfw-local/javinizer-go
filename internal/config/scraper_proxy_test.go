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

	t.Run("returns direct (empty) when override is explicitly disabled", func(t *testing.T) {
		override := &ProxyConfig{
			Enabled: false,
		}
		resolved := ResolveScraperProxy(global, override)
		assert.Equal(t, "", resolved.URL)
		assert.Equal(t, "", resolved.Username)
		assert.Equal(t, "", resolved.Password)
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

func TestResolveScraperProxyMode(t *testing.T) {
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

	t.Run("global disabled returns direct", func(t *testing.T) {
		disabledGlobal := ProxyConfig{Enabled: false}
		mode := ResolveScraperProxyMode(disabledGlobal, &ProxyConfig{Enabled: true, Profile: "main"})
		assert.Equal(t, ScraperProxyModeDirect, mode)
	})

	t.Run("nil override returns inherit", func(t *testing.T) {
		mode := ResolveScraperProxyMode(global, nil)
		assert.Equal(t, ScraperProxyModeInherit, mode)
	})

	t.Run("disabled override returns direct", func(t *testing.T) {
		override := &ProxyConfig{Enabled: false}
		mode := ResolveScraperProxyMode(global, override)
		assert.Equal(t, ScraperProxyModeDirect, mode)
	})

	t.Run("enabled without profile returns inherit", func(t *testing.T) {
		override := &ProxyConfig{Enabled: true}
		mode := ResolveScraperProxyMode(global, override)
		assert.Equal(t, ScraperProxyModeInherit, mode)
	})

	t.Run("enabled with profile returns specific", func(t *testing.T) {
		override := &ProxyConfig{Enabled: true, Profile: "backup"}
		mode := ResolveScraperProxyMode(global, override)
		assert.Equal(t, ScraperProxyModeSpecific, mode)
	})

	t.Run("enabled with nonexistent profile returns specific", func(t *testing.T) {
		// Note: mode only checks if profile string is non-empty, existence is checked in ResolveScraperProxy
		override := &ProxyConfig{Enabled: true, Profile: "nonexistent"}
		mode := ResolveScraperProxyMode(global, override)
		assert.Equal(t, ScraperProxyModeSpecific, mode)
	})
}

func TestResolveScraperProxy_WithMode(t *testing.T) {
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

	t.Run("direct mode returns empty profile", func(t *testing.T) {
		// Direct mode when global is disabled
		disabledGlobal := ProxyConfig{Enabled: false}
		override := &ProxyConfig{Enabled: true, Profile: "main"}
		profile := ResolveScraperProxy(disabledGlobal, override)
		assert.Equal(t, "", profile.URL)
	})

	t.Run("disabled override returns direct with empty profile", func(t *testing.T) {
		// Direct mode when scraper explicitly disabled (even with global enabled)
		override := &ProxyConfig{Enabled: false}
		profile := ResolveScraperProxy(global, override)
		assert.Equal(t, "", profile.URL)
	})

	t.Run("inherit mode returns global default", func(t *testing.T) {
		override := &ProxyConfig{Enabled: true}
		profile := ResolveScraperProxy(global, override)
		assert.Equal(t, "http://main-proxy.example.com:8080", profile.URL)
	})

	t.Run("specific mode returns profile with credential inheritance", func(t *testing.T) {
		// Profile with URL only, should inherit credentials from global
		globalWithCreds := ProxyConfig{
			Enabled:        true,
			DefaultProfile: "main",
			Profiles: map[string]ProxyProfile{
				"main": {
					URL:      "http://main-proxy.example.com:8080",
					Username: "global-user",
					Password: "global-pass",
				},
				"specific": {
					URL: "http://specific-proxy.example.com:9090",
					// No credentials - should inherit from global
				},
			},
		}
		override := &ProxyConfig{Enabled: true, Profile: "specific"}
		profile := ResolveScraperProxy(globalWithCreds, override)
		assert.Equal(t, "http://specific-proxy.example.com:9090", profile.URL)
		assert.Equal(t, "global-user", profile.Username)
		assert.Equal(t, "global-pass", profile.Password)
	})

	t.Run("specific mode with nonexistent profile falls back to inherit", func(t *testing.T) {
		override := &ProxyConfig{Enabled: true, Profile: "nonexistent"}
		profile := ResolveScraperProxy(global, override)
		// Falls back to global default since profile doesn't exist
		assert.Equal(t, "http://main-proxy.example.com:8080", profile.URL)
		assert.Equal(t, "main-user", profile.Username)
		assert.Equal(t, "main-pass", profile.Password)
	})
}
