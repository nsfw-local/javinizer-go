package config

// ScraperSettings holds common scraper configuration fields used by the Scraper interface.
// Individual scraper configs embed this and add scraper-specific fields.
// CONF-01: All fields are present: Enabled, Timeout, RateLimit, RetryCount,
// FlareSolverr, UserAgent, Extra.
type ScraperSettings struct {
	Enabled       bool               `yaml:"enabled" json:"enabled"`
	Language      string             `yaml:"language" json:"language"`                                 // Language code varies by scraper
	Timeout       int                `yaml:"timeout" json:"timeout"`                                   // HTTP client timeout in seconds
	RateLimit     int                `yaml:"rate_limit" json:"rate_limit"`                             // Request delay in milliseconds (mirrors RequestDelay)
	RetryCount    int                `yaml:"retry_count" json:"retry_count"`                           // Max retries (mirrors MaxRetries)
	UserAgent     string             `yaml:"user_agent" json:"user_agent"`                             // Custom User-Agent; if empty, DefaultFakeUserAgent is used
	Proxy         *ProxyConfig       `yaml:"proxy,omitempty" json:"proxy,omitempty"`                   // Optional scraper-specific proxy override
	DownloadProxy *ProxyConfig       `yaml:"download_proxy,omitempty" json:"download_proxy,omitempty"` // Optional scraper-specific download proxy override
	FlareSolverr  FlareSolverrConfig `yaml:"flaresolverr" json:"flaresolverr"`                         // HTTP-03: FlareSolverr on ScraperSettings
	BaseURL       string             `yaml:"base_url,omitempty" json:"base_url,omitempty"`             // Base URL for the scraper
	Extra         map[string]any     `yaml:"extra,omitempty" json:"extra,omitempty"`                   // CONF-06: scraper-specific fields
}

// MarshalYAML preserves the full unified scraper settings shape so config
// save/load round-trips do not drop scraper-specific data.
// FlareSolverr is omitted when zero-value since it's typically configured globally.
//
// MAINTENANCE WARNING: The anonymous struct in the zero-flare branch must be kept
// in sync with ScraperSettings fields (excluding FlareSolverr). When adding new
// fields to ScraperSettings, update the anonymous struct and the assignment below.
func (s *ScraperSettings) MarshalYAML() (interface{}, error) {
	// Nil receiver guard
	if s == nil {
		return nil, nil
	}

	// Helper type to avoid infinite recursion
	// Check if FlareSolverr is zero-value (default/unconfigured)
	isZeroFlare := !s.FlareSolverr.Enabled &&
		s.FlareSolverr.URL == "" &&
		s.FlareSolverr.Timeout == 0 &&
		s.FlareSolverr.MaxRetries == 0 &&
		s.FlareSolverr.SessionTTL == 0

	if isZeroFlare {
		// Omit FlareSolverr field entirely when zero.
		// IMPORTANT: Keep this struct synchronized with ScraperSettings (minus FlareSolverr).
		return &struct {
			Enabled       bool           `yaml:"enabled"`
			Language      string         `yaml:"language"`
			Timeout       int            `yaml:"timeout"`
			RateLimit     int            `yaml:"rate_limit"`
			RetryCount    int            `yaml:"retry_count"`
			UserAgent     string         `yaml:"user_agent"`
			Proxy         *ProxyConfig   `yaml:"proxy,omitempty"`
			DownloadProxy *ProxyConfig   `yaml:"download_proxy,omitempty"`
			BaseURL       string         `yaml:"base_url,omitempty"`
			Extra         map[string]any `yaml:"extra,omitempty"`
		}{
			Enabled:       s.Enabled,
			Language:      s.Language,
			Timeout:       s.Timeout,
			RateLimit:     s.RateLimit,
			RetryCount:    s.RetryCount,
			UserAgent:     s.UserAgent,
			Proxy:         s.Proxy,
			DownloadProxy: s.DownloadProxy,
			BaseURL:       s.BaseURL,
			Extra:         s.Extra,
		}, nil
	}

	// Include FlareSolverr when it's non-zero
	type scraperSettingsNoFlare ScraperSettings
	return (*scraperSettingsNoFlare)(s), nil
}

// ToScraperSettings implements ScraperSettingsAdapter.
func (s *ScraperSettings) ToScraperSettings() *ScraperSettings {
	return s
}

// GetBoolExtra returns a boolean from Extra map with type safety.
func (sc *ScraperSettings) GetBoolExtra(key string, defaultVal bool) bool {
	if sc.Extra == nil {
		return defaultVal
	}
	v, ok := sc.Extra[key]
	if !ok {
		return defaultVal
	}
	b, ok := v.(bool)
	if !ok {
		return defaultVal
	}
	return b
}

// GetIntExtra returns an integer from Extra map with type safety.
// Handles int and float64 (JSON numbers are float64 when unmarshaled into map[string]any).
func (sc *ScraperSettings) GetIntExtra(key string, defaultVal int) int {
	if sc.Extra == nil {
		return defaultVal
	}
	v, ok := sc.Extra[key]
	if !ok {
		return defaultVal
	}
	// Handle int directly
	if i, ok := v.(int); ok {
		return i
	}
	// Handle float64 from JSON unmarshaling (JSON numbers become float64)
	if f, ok := v.(float64); ok {
		return int(f)
	}
	return defaultVal
}

// GetStringExtra returns a string from Extra map with type safety.
func (sc *ScraperSettings) GetStringExtra(key string, defaultVal string) string {
	if sc.Extra == nil {
		return defaultVal
	}
	v, ok := sc.Extra[key]
	if !ok {
		return defaultVal
	}
	s, ok := v.(string)
	if !ok {
		return defaultVal
	}
	return s
}

// GetBaseURL returns the base URL for the scraper, or an empty string if not set.
func (sc *ScraperSettings) GetBaseURL() string {
	return sc.BaseURL
}

// DeepCopy creates a deep copy of ScraperSettings, ensuring that pointer fields
// (Proxy, DownloadProxy) and map fields (Extra) are properly isolated from the original.
// This prevents mutation leaks when settings are shared between config instances.
func (s *ScraperSettings) DeepCopy() *ScraperSettings {
	if s == nil {
		return nil
	}
	copy := &ScraperSettings{
		Enabled:      s.Enabled,
		Language:     s.Language,
		Timeout:      s.Timeout,
		RateLimit:    s.RateLimit,
		RetryCount:   s.RetryCount,
		UserAgent:    s.UserAgent,
		FlareSolverr: s.FlareSolverr, // Value type, automatic copy
		BaseURL:      s.BaseURL,
	}

	// Deep copy Proxy if not nil (profile-based only)
	if s.Proxy != nil {
		copy.Proxy = &ProxyConfig{
			Enabled:        s.Proxy.Enabled,
			Profile:        s.Proxy.Profile,
			DefaultProfile: s.Proxy.DefaultProfile,
			Profiles:       deepCopyProxyProfiles(s.Proxy.Profiles),
		}
	}

	// Deep copy DownloadProxy if not nil (profile-based only)
	if s.DownloadProxy != nil {
		copy.DownloadProxy = &ProxyConfig{
			Enabled:        s.DownloadProxy.Enabled,
			Profile:        s.DownloadProxy.Profile,
			DefaultProfile: s.DownloadProxy.DefaultProfile,
			Profiles:       deepCopyProxyProfiles(s.DownloadProxy.Profiles),
		}
	}

	// Deep copy Extra map (shallow copy of values is acceptable per current usage)
	if s.Extra != nil {
		copy.Extra = make(map[string]any, len(s.Extra))
		for k, v := range s.Extra {
			copy.Extra[k] = v
		}
	}

	return copy
}

// deepCopyProxyProfiles creates a deep copy of the proxy profiles map
func deepCopyProxyProfiles(profiles map[string]ProxyProfile) map[string]ProxyProfile {
	if profiles == nil {
		return nil
	}

	copy := make(map[string]ProxyProfile, len(profiles))
	for k, v := range profiles {
		copy[k] = ProxyProfile{
			URL:      v.URL,
			Username: v.Username,
			Password: v.Password,
		}
	}
	return copy
}

// ScraperCommonConfig holds common scraper configuration fields used by ScraperConfigInterface.
// Embed this struct in all scraper-specific configs with `yaml:",inline"` to automatically
// satisfy the interface without boilerplate wrapper methods.
type ScraperCommonConfig struct {
	Enabled       bool            `yaml:"enabled"`
	RequestDelay  int             `yaml:"request_delay"`
	MaxRetries    int             `yaml:"max_retries"`
	UserAgent     UserAgentString `yaml:"user_agent"`
	Proxy         *ProxyConfig    `yaml:"proxy,omitempty"`
	DownloadProxy *ProxyConfig    `yaml:"download_proxy,omitempty"`
}

// IsEnabled implements ScraperConfigInterface.
func (c ScraperCommonConfig) IsEnabled() bool { return c.Enabled }

// GetUserAgent implements ScraperConfigInterface.
func (c ScraperCommonConfig) GetUserAgent() string { return c.UserAgent.Value }

// GetRequestDelay implements ScraperConfigInterface.
func (c ScraperCommonConfig) GetRequestDelay() int { return c.RequestDelay }

// GetMaxRetries implements ScraperConfigInterface.
func (c ScraperCommonConfig) GetMaxRetries() int { return c.MaxRetries }

// GetProxy implements ScraperConfigInterface.
func (c ScraperCommonConfig) GetProxy() any { return c.Proxy }

// GetDownloadProxy implements ScraperConfigInterface.
func (c ScraperCommonConfig) GetDownloadProxy() any { return c.DownloadProxy }
