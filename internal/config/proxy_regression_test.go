package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Regression tests for issues found in code review rounds 1-6

// Round 1: Frontend/backend Direct mode mismatch
// Issue: Backend returned 'inherit' when scraper override was disabled
// Fix: ResolveScraperProxyMode now returns Direct when !scraperOverride.Enabled
func TestRegression_DirectModeWhenScraperDisabled(t *testing.T) {
	global := ProxyConfig{
		Enabled:        true,
		DefaultProfile: "main",
		Profiles: map[string]ProxyProfile{
			"main": {URL: "http://main-proxy.example.com:8080"},
		},
	}

	// Scraper explicitly disabled → Direct
	override := &ProxyConfig{Enabled: false}
	mode := ResolveScraperProxyMode(global, override)
	assert.Equal(t, ScraperProxyModeDirect, mode, "Disabled scraper should use Direct mode")

	// Global disabled forces Direct regardless of scraper override
	disabledGlobal := ProxyConfig{Enabled: false}
	mode = ResolveScraperProxyMode(disabledGlobal, &ProxyConfig{Enabled: true, Profile: "main"})
	assert.Equal(t, ScraperProxyModeDirect, mode, "Global disabled should force Direct mode")
}

// Round 2: Nil override semantics mismatch
// Issue: Frontend treated missing proxy config as 'direct', backend as 'inherit'
// Fix: Frontend now matches backend - nil override = inherit
func TestRegression_NilOverrideIsInherit(t *testing.T) {
	global := ProxyConfig{
		Enabled:        true,
		DefaultProfile: "main",
		Profiles: map[string]ProxyProfile{
			"main": {URL: "http://main-proxy.example.com:8080"},
		},
	}

	// Nil override → Inherit
	mode := ResolveScraperProxyMode(global, nil)
	assert.Equal(t, ScraperProxyModeInherit, mode, "Nil override should be Inherit")

	profile := ResolveScraperProxy(global, nil)
	assert.Equal(t, "http://main-proxy.example.com:8080", profile.URL, "Should inherit global default")
}

// Round 3: Validation rejecting inherit mode
// Issue: validateProxyProfileRef rejected enabled=true with empty profile
// Fix: Validation now accepts enabled=true, profile="" as valid inherit mode
func TestRegression_ValidationAcceptsInheritMode(t *testing.T) {
	profiles := map[string]ProxyProfile{
		"main": {URL: "http://main-proxy.example.com:8080"},
	}

	// enabled=true with empty profile = inherit mode (valid)
	inheritConfig := &ProxyConfig{Enabled: true, Profile: ""}
	err := validateProxyProfileRef("scrapers.javlibrary.proxy", inheritConfig, profiles)
	assert.NoError(t, err, "Inherit mode (enabled=true, profile='') should be valid")

	// enabled=true with profile = specific mode (valid if profile exists)
	specificConfig := &ProxyConfig{Enabled: true, Profile: "main"}
	err = validateProxyProfileRef("scrapers.javlibrary.proxy", specificConfig, profiles)
	assert.NoError(t, err, "Specific mode with valid profile should be valid")

	// enabled=true with non-existent profile = invalid
	invalidConfig := &ProxyConfig{Enabled: true, Profile: "nonexistent"}
	err = validateProxyProfileRef("scrapers.javlibrary.proxy", invalidConfig, profiles)
	assert.Error(t, err, "Specific mode with invalid profile should fail validation")
}

// Round 4: Credential inheritance in specific mode
// Issue: resolveProxyProfileForTest didn't inherit credentials from global default
// Fix: Credential inheritance is handled in ResolveScraperProxy for specific mode
func TestRegression_CredentialInheritanceInSpecificMode(t *testing.T) {
	global := ProxyConfig{
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

	// Test specific mode with profile that has no credentials
	scraperOverride := &ProxyConfig{Enabled: true, Profile: "specific"}
	resolved := ResolveScraperProxy(global, scraperOverride)

	assert.Equal(t, "http://specific-proxy.example.com:9090", resolved.URL, "URL should come from specific profile")
	assert.Equal(t, "global-user", resolved.Username, "Username should inherit from global default")
	assert.Equal(t, "global-pass", resolved.Password, "Password should inherit from global default")
}

// Round 5: Mode mismatch for partial configs
// Issue: Frontend treated proxy: { profile: "x" } as specific, but backend requires enabled=true
// Fix: Frontend now requires enabled === true for specific/inherit modes
func TestRegression_EnabledMustBeExplicitlyTrue(t *testing.T) {
	global := ProxyConfig{
		Enabled: true,
		Profiles: map[string]ProxyProfile{
			"backup": {URL: "http://backup.example.com:8080"},
		},
	}

	// Partial config with profile but no enabled flag (enabled is undefined/false by default)
	partialConfig := &ProxyConfig{Profile: "backup"}
	mode := ResolveScraperProxyMode(global, partialConfig)
	assert.Equal(t, ScraperProxyModeDirect, mode, "Partial config without enabled=true should be Direct")

	// Same with enabled explicitly true
	explicitConfig := &ProxyConfig{Enabled: true, Profile: "backup"}
	mode = ResolveScraperProxyMode(global, explicitConfig)
	assert.Equal(t, ScraperProxyModeSpecific, mode, "Explicit enabled=true with profile should be Specific")
}
