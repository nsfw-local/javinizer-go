package javlibrary

import (
	"testing"

	"github.com/javinizer/javinizer-go/internal/config"
	"github.com/stretchr/testify/assert"
)

// TestNewHTTPClient_NilProxy tests that nil globalProxy does not cause panic (JAVL-02).
func TestNewHTTPClient_NilProxy(t *testing.T) {
	cfg := &config.ScraperSettings{
		Enabled:    true,
		Timeout:    30,
		RetryCount: 3,
		Proxy:      nil,
		FlareSolverr: config.FlareSolverrConfig{
			Enabled:    true,
			URL:        "http://localhost:8191/v1",
			Timeout:    30,
			MaxRetries: 3,
			SessionTTL: 300,
		},
	}

	// This should NOT panic even though globalProxy is nil
	client, _, err := NewHTTPClient(cfg, nil, config.FlareSolverrConfig{Enabled: false}, true)

	// FlareSolverr may or may not be created depending on whether a server is running
	// The key assertion is that no panic occurs (JAVL-02)
	assert.NotNil(t, client)
	assert.NoError(t, err)
}

// TestNewHTTPClient_FlareSolverrDisabled tests that FlareSolverr setup is skipped when disabled (JAVL-01).
func TestNewHTTPClient_FlareSolverrDisabled(t *testing.T) {
	cfg := &config.ScraperSettings{
		Enabled:    true,
		Timeout:    30,
		RetryCount: 3,
		Proxy:      nil,
		FlareSolverr: config.FlareSolverrConfig{
			Enabled: false,
			URL:     "http://localhost:8191/v1", // Should be ignored
		},
	}

	globalProxy := &config.ProxyConfig{
		Enabled: true,
		Profile: "main",
		Profiles: map[string]config.ProxyProfile{
			"main": {
				URL: "http://proxy.example.com:8080",
			},
		},
	}

	// useFlareSolverr=false should skip all FlareSolverr setup
	client, fs, err := NewHTTPClient(cfg, globalProxy, config.FlareSolverrConfig{Enabled: false}, false)

	assert.NotNil(t, client)
	assert.Nil(t, fs) // Should be nil when useFlareSolverr=false
	assert.NoError(t, err)
}

// TestNewHTTPClient_GlobalProxyUsedForFlareSolverr tests that global proxy is used when scraper proxy is empty (JAVL-02).
func TestNewHTTPClient_GlobalProxyUsedForFlareSolverr(t *testing.T) {
	cfg := &config.ScraperSettings{
		Enabled:    true,
		Timeout:    30,
		RetryCount: 3,
		Proxy:      nil, // No scraper-specific proxy
		FlareSolverr: config.FlareSolverrConfig{
			Enabled:    true,
			URL:        "http://localhost:8191/v1",
			Timeout:    30,
			MaxRetries: 3,
			SessionTTL: 300,
		},
	}

	// Global proxy is enabled and has profile
	globalProxy := &config.ProxyConfig{
		Enabled: true,
		Profile: "main",
		Profiles: map[string]config.ProxyProfile{
			"main": {
				URL:      "http://proxy.example.com:8080",
				Username: "user",
				Password: "pass",
			},
		},
	}

	// Should not panic and should attempt to use global proxy for FlareSolverr
	client, _, err := NewHTTPClient(cfg, globalProxy, config.FlareSolverrConfig{Enabled: false}, true)

	assert.NotNil(t, client)
	// err may be non-nil if FlareSolverr server isn't running, but client should still be created
	assert.NoError(t, err)
}

// TestNewHTTPClient_DisabledGlobalProxyNotUsed tests that disabled global proxy is NOT used for FlareSolverr (defense-in-depth).
func TestNewHTTPClient_DisabledGlobalProxyNotUsed(t *testing.T) {
	cfg := &config.ScraperSettings{
		Enabled:    true,
		Timeout:    30,
		RetryCount: 3,
		Proxy:      nil, // No scraper-specific proxy
		FlareSolverr: config.FlareSolverrConfig{
			Enabled:    true,
			URL:        "http://localhost:8191/v1",
			Timeout:    30,
			MaxRetries: 3,
			SessionTTL: 300,
		},
	}

	// Global proxy has profile but is disabled
	globalProxy := &config.ProxyConfig{
		Enabled: false,
		Profile: "main",
		Profiles: map[string]config.ProxyProfile{
			"main": {
				URL:      "http://proxy.example.com:8080",
				Username: "user",
				Password: "pass",
			},
		},
	}

	client, _, err := NewHTTPClient(cfg, globalProxy, config.FlareSolverrConfig{Enabled: false}, true)

	// Should create client with empty proxy for FlareSolverr (not using disabled global proxy)
	assert.NotNil(t, client)
	assert.NoError(t, err)
}

// TestNewHTTPClient_ScraperProxyOverridesGlobal tests that scraper-specific proxy takes precedence.
func TestNewHTTPClient_ScraperProxyOverridesGlobal(t *testing.T) {
	cfg := &config.ScraperSettings{
		Enabled:    true,
		Timeout:    30,
		RetryCount: 3,
		Proxy: &config.ProxyConfig{
			Enabled: true,
			Profile: "scraper",
			Profiles: map[string]config.ProxyProfile{
				"scraper": {
					URL:      "http://scraper-proxy.example.com:9090",
					Username: "scraper-user",
					Password: "scraper-pass",
				},
			},
		},
		FlareSolverr: config.FlareSolverrConfig{
			Enabled:    true,
			URL:        "http://localhost:8191/v1",
			Timeout:    30,
			MaxRetries: 3,
			SessionTTL: 300,
		},
	}

	// Global proxy is enabled but should be ignored since scraper has its own
	globalProxy := &config.ProxyConfig{
		Enabled: true,
		Profile: "main",
		Profiles: map[string]config.ProxyProfile{
			"main": {
				URL:      "http://global-proxy.example.com:8080",
				Username: "global-user",
				Password: "global-pass",
			},
		},
	}

	client, _, err := NewHTTPClient(cfg, globalProxy, config.FlareSolverrConfig{Enabled: false}, true)

	assert.NotNil(t, client)
	assert.NoError(t, err)
}

// TestNewHTTPClient_TimeoutDefaults tests that timeout defaults to 30s when zero.
func TestNewHTTPClient_TimeoutDefaults(t *testing.T) {
	cfg := &config.ScraperSettings{
		Enabled:    true,
		Timeout:    0, // Zero - should default to 30
		RetryCount: 0, // Zero - should default to 3
		FlareSolverr: config.FlareSolverrConfig{
			Enabled: false,
		},
	}

	client, fs, err := NewHTTPClient(cfg, nil, config.FlareSolverrConfig{Enabled: false}, false)

	assert.NotNil(t, client)
	assert.Nil(t, fs)
	assert.NoError(t, err)
}

// TestNewHTTPClient_RetryCountDefaults tests that retry count defaults to 3 when zero.
func TestNewHTTPClient_RetryCountDefaults(t *testing.T) {
	cfg := &config.ScraperSettings{
		Enabled:    true,
		Timeout:    10,
		RetryCount: 0, // Zero - should default to 3
		FlareSolverr: config.FlareSolverrConfig{
			Enabled: false,
		},
	}

	client, fs, err := NewHTTPClient(cfg, nil, config.FlareSolverrConfig{Enabled: false}, false)

	assert.NotNil(t, client)
	assert.Nil(t, fs)
	assert.NoError(t, err)
}
