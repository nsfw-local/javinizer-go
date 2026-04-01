package config

import (
	"strings"
)

// ProxyProfile holds reusable proxy connection settings.
type ProxyProfile struct {
	URL      string `yaml:"url" json:"url"`
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
}

// ProxyConfig holds HTTP/SOCKS5 proxy configuration
// All proxy settings are managed through profiles (scrapers.proxy.profiles).
// Use Profile field to reference a named profile for scraper-specific overrides.
type ProxyConfig struct {
	Enabled        bool                    `yaml:"enabled" json:"enabled"`                                     // Enable proxy for HTTP requests
	Profile        string                  `yaml:"profile,omitempty" json:"profile,omitempty"`                 // Named profile to use (for scraper-specific overrides)
	DefaultProfile string                  `yaml:"default_profile,omitempty" json:"default_profile,omitempty"` // Default profile name (for global scrapers.proxy)
	Profiles       map[string]ProxyProfile `yaml:"profiles,omitempty" json:"profiles,omitempty"`               // Named proxy profiles (global scrapers.proxy)
}

// ResolveScraperUserAgent resolves the effective User-Agent for a scraper.
// If the scraper-specific userAgent is non-empty, it is used; otherwise DefaultFakeUserAgent is returned.
func ResolveScraperUserAgent(userAgent string) string {
	if ua := strings.TrimSpace(userAgent); ua != "" {
		return ua
	}
	return DefaultFakeUserAgent
}

// ResolveScraperProxy returns the effective proxy profile for a scraper.
// When no scraper-specific override exists or it is disabled, the global proxy
// profile is used as a fallback (if enabled), preserving backward compatibility with
// existing configs that only have scrapers.proxy configured.
// When enabled, proxy profiles are applied first, then missing URL/credentials
// inherit from the globally resolved proxy profile.
// Note: FlareSolverr is handled separately via ScraperSettings.FlareSolverr and
// ScrapersConfig.FlareSolverr (global), not via ProxyConfig.
func ResolveScraperProxy(global ProxyConfig, scraperOverride *ProxyConfig) *ProxyProfile {
	globalProfile := ResolveGlobalProxy(global)

	// If no scraper override, fall back to global proxy if enabled.
	if scraperOverride == nil {
		if global.Enabled {
			return globalProfile
		}
		return &ProxyProfile{}
	}

	// Scraper override exists but is disabled - check if global is enabled.
	if !scraperOverride.Enabled {
		if global.Enabled {
			return globalProfile
		}
		return &ProxyProfile{}
	}

	// Scraper override is enabled - resolve its profile
	resolved := ProxyProfile{}
	if scraperOverride.Profile != "" {
		if profile, ok := global.Profiles[scraperOverride.Profile]; ok {
			resolved = profile
		}
	}

	// If proxy is enabled but URL is omitted, inherit global proxy
	// credentials so users can toggle per-scraper proxy usage without
	// duplicating global proxy values.
	if resolved.URL == "" {
		resolved.URL = globalProfile.URL
		if resolved.Username == "" {
			resolved.Username = globalProfile.Username
		}
		if resolved.Password == "" {
			resolved.Password = globalProfile.Password
		}
	}
	return &resolved
}

// ResolveGlobalProxy returns the effective global proxy profile, including the
// selected default profile when configured. Returns empty profile if proxy is disabled.
func ResolveGlobalProxy(global ProxyConfig) *ProxyProfile {
	if !global.Enabled {
		return &ProxyProfile{}
	}
	if global.DefaultProfile != "" {
		if profile, ok := global.Profiles[global.DefaultProfile]; ok {
			return &profile
		}
	}
	return &ProxyProfile{}
}
