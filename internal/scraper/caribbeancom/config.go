package caribbeancom

import (
	"fmt"
	"strings"

	"github.com/javinizer/javinizer-go/internal/config"
	"github.com/javinizer/javinizer-go/internal/configutil"
	"github.com/javinizer/javinizer-go/internal/scraperutil"
)

// Config holds Caribbeancom scraper configuration.
// YAML tags are defined here for unmarshaling via config.ScrapersConfig.
type CaribbeancomConfig struct {
	Enabled       bool                      `yaml:"enabled" json:"enabled"`
	Language      string                    `yaml:"language" json:"language"`
	RequestDelay  int                       `yaml:"request_delay" json:"request_delay"`
	MaxRetries    int                       `yaml:"max_retries" json:"max_retries"`
	BaseURL       string                    `yaml:"base_url" json:"base_url"`
	UserAgent     string                    `yaml:"user_agent" json:"user_agent"`
	Proxy         *config.ProxyConfig       `yaml:"proxy,omitempty" json:"proxy,omitempty"`
	DownloadProxy *config.ProxyConfig       `yaml:"download_proxy,omitempty" json:"download_proxy,omitempty"`
	Priority      int                       `yaml:"priority" json:"priority"` // Scraper's priority (higher = higher priority)
	FlareSolverr  config.FlareSolverrConfig `yaml:"flaresolverr" json:"flaresolverr"`
}

func init() {
	scraperutil.RegisterValidator("caribbeancom", func(a any) error {
		return (&CaribbeancomConfig{}).ValidateConfig(a.(*config.ScraperSettings))
	})
	// PLUGIN-01: Register config field accessor for registry-based iteration
	// Note: getter methods were removed in Phase 01. The normalize function will
	// fall back to scraperConfigs map directly if this returns nil.
	scraperutil.RegisterScraperConfig("caribbeancom", func(a any) any {
		return nil
	})
	// TASK 5: Register ConfigFactory for UnmarshalYAML
	scraperutil.RegisterConfigFactory("caribbeancom", func() any {
		return &CaribbeancomConfig{}
	})
	// TASK 3: Register flatten function for registry-based type conversion
	scraperutil.RegisterFlattenFunc("caribbeancom", func(cfg any) any {
		c, ok := cfg.(scraperutil.ScraperConfigInterface)
		if !ok {
			return nil
		}
		proxy := c.GetProxy()
		downloadProxy := c.GetDownloadProxy()
		var proxyVal, downloadProxyVal *config.ProxyConfig
		if proxy != nil {
			proxyVal = proxy.(*config.ProxyConfig)
		}
		if downloadProxy != nil {
			downloadProxyVal = downloadProxy.(*config.ProxyConfig)
		}
		// Use type assertion to access Caribbeancom-specific fields
		if caribbeancomCfg, ok := cfg.(*CaribbeancomConfig); ok {
			return &config.ScraperSettings{
				Enabled:   c.IsEnabled(),
				Language:  "",
				RateLimit: c.GetRequestDelay(),
				Extra: map[string]any{
					"base_url": "https://www.caribbeancom.com",
				},
				Proxy:         proxyVal,
				DownloadProxy: downloadProxyVal,
				FlareSolverr:  caribbeancomCfg.FlareSolverr,
			}
		}
		return nil
	})
}

// IsEnabled implements scraperutil.ScraperConfigInterface.
func (c *CaribbeancomConfig) IsEnabled() bool { return c.Enabled }

// GetUserAgent implements scraperutil.ScraperConfigInterface.
func (c *CaribbeancomConfig) GetUserAgent() string { return c.UserAgent }

// GetRequestDelay implements scraperutil.ScraperConfigInterface.
func (c *CaribbeancomConfig) GetRequestDelay() int { return c.RequestDelay }

// GetMaxRetries implements scraperutil.ScraperConfigInterface.
func (c *CaribbeancomConfig) GetMaxRetries() int { return c.MaxRetries }

// GetProxy implements scraperutil.ScraperConfigInterface.
func (c *CaribbeancomConfig) GetProxy() any { return c.Proxy }

// GetDownloadProxy implements scraperutil.ScraperConfigInterface.
func (c *CaribbeancomConfig) GetDownloadProxy() any { return c.DownloadProxy }

// ValidateConfig implements config.ConfigValidator for CaribbeancomConfig.
func (c *CaribbeancomConfig) ValidateConfig(sc *config.ScraperSettings) error {
	if sc == nil {
		return fmt.Errorf("caribbeancom: config is nil")
	}
	if !sc.Enabled {
		return nil // Disabled is valid
	}
	// Validate language if set
	switch strings.ToLower(strings.TrimSpace(sc.Language)) {
	case "", "ja", "en":
		// Valid
	default:
		return fmt.Errorf("caribbeancom: language must be 'ja' or 'en', got %q", sc.Language)
	}
	// Validate rate limit
	if sc.RateLimit < 0 {
		return fmt.Errorf("caribbeancom: rate_limit must be non-negative, got %d", sc.RateLimit)
	}
	// Validate retry count
	if sc.RetryCount < 0 {
		return fmt.Errorf("caribbeancom: retry_count must be non-negative, got %d", sc.RetryCount)
	}
	// Validate timeout
	if sc.Timeout < 0 {
		return fmt.Errorf("caribbeancom: timeout must be non-negative, got %d", sc.Timeout)
	}
	// Validate base URL if set
	if err := configutil.ValidateHTTPBaseURL("caribbeancom.base_url", sc.BaseURL); err != nil {
		return err
	}
	// Validate FlareSolverr config if enabled
	if sc.FlareSolverr.Enabled {
		if sc.FlareSolverr.URL == "" {
			return fmt.Errorf("caribbeancom.flaresolverr.url is required when flaresolverr is enabled")
		}
		if sc.FlareSolverr.Timeout < 1 || sc.FlareSolverr.Timeout > 300 {
			return fmt.Errorf("caribbeancom.flaresolverr.timeout must be between 1 and 300")
		}
		if sc.FlareSolverr.MaxRetries < 0 || sc.FlareSolverr.MaxRetries > 10 {
			return fmt.Errorf("caribbeancom.flaresolverr.max_retries must be between 0 and 10")
		}
		if sc.FlareSolverr.SessionTTL < 60 || sc.FlareSolverr.SessionTTL > 3600 {
			return fmt.Errorf("caribbeancom.flaresolverr.session_ttl must be between 60 and 3600")
		}
	}
	return nil
}
