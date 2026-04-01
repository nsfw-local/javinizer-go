package javbus

import (
	"fmt"
	"strings"

	"github.com/javinizer/javinizer-go/internal/config"
	"github.com/javinizer/javinizer-go/internal/configutil"
	"github.com/javinizer/javinizer-go/internal/scraperutil"
)

// Config holds JavBus scraper configuration.
// YAML tags are defined here for unmarshaling via config.ScrapersConfig.
type JavBusConfig struct {
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
	scraperutil.RegisterValidator("javbus", func(a any) error {
		return (&JavBusConfig{}).ValidateConfig(a.(*config.ScraperSettings))
	})
	// PLUGIN-01: Register config field accessor for registry-based iteration
	// Note: getter methods were removed in Phase 01. The normalize function will
	// fall back to scraperConfigs map directly if this returns nil.
	scraperutil.RegisterScraperConfig("javbus", func(a any) any {
		return nil
	})
	// TASK 5: Register ConfigFactory for UnmarshalYAML
	scraperutil.RegisterConfigFactory("javbus", func() any {
		return &JavBusConfig{}
	})
	// TASK 3: Register flatten function for registry-based type conversion
	scraperutil.RegisterFlattenFunc("javbus", func(cfg any) any {
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
		// Use type assertion to access JavBus-specific fields
		if javbusCfg, ok := cfg.(*JavBusConfig); ok {
			return &config.ScraperSettings{
				Enabled:   c.IsEnabled(),
				Language:  "",
				RateLimit: c.GetRequestDelay(),
				Extra: map[string]any{
					"base_url": "https://www.javbus.com",
				},
				Proxy:         proxyVal,
				DownloadProxy: downloadProxyVal,
				FlareSolverr:  javbusCfg.FlareSolverr,
			}
		}
		return nil
	})
}

// IsEnabled implements scraperutil.ScraperConfigInterface.
func (c *JavBusConfig) IsEnabled() bool { return c.Enabled }

// GetUserAgent implements scraperutil.ScraperConfigInterface.
func (c *JavBusConfig) GetUserAgent() string { return c.UserAgent }

// GetRequestDelay implements scraperutil.ScraperConfigInterface.
func (c *JavBusConfig) GetRequestDelay() int { return c.RequestDelay }

// GetMaxRetries implements scraperutil.ScraperConfigInterface.
func (c *JavBusConfig) GetMaxRetries() int { return c.MaxRetries }

// GetProxy implements scraperutil.ScraperConfigInterface.
func (c *JavBusConfig) GetProxy() any { return c.Proxy }

// GetDownloadProxy implements scraperutil.ScraperConfigInterface.
func (c *JavBusConfig) GetDownloadProxy() any { return c.DownloadProxy }

// ValidateConfig implements config.ConfigValidator for JavBusConfig.
func (c *JavBusConfig) ValidateConfig(sc *config.ScraperSettings) error {
	if sc == nil {
		return fmt.Errorf("javbus: config is nil")
	}
	if !sc.Enabled {
		return nil // Disabled is valid
	}
	// Validate language
	switch strings.ToLower(strings.TrimSpace(sc.Language)) {
	case "", "en":
		// Valid
	case "ja":
		// Valid
	case "zh":
		// Valid
	default:
		return fmt.Errorf("javbus: language must be 'en', 'ja', or 'zh', got %q", sc.Language)
	}
	// Validate rate limit
	if sc.RateLimit < 0 {
		return fmt.Errorf("javbus: rate_limit must be non-negative, got %d", sc.RateLimit)
	}
	// Validate retry count
	if sc.RetryCount < 0 {
		return fmt.Errorf("javbus: retry_count must be non-negative, got %d", sc.RetryCount)
	}
	// Validate timeout
	if sc.Timeout < 0 {
		return fmt.Errorf("javbus: timeout must be non-negative, got %d", sc.Timeout)
	}
	// Validate base URL if set
	if err := configutil.ValidateHTTPBaseURL("javbus.base_url", sc.BaseURL); err != nil {
		return err
	}
	return nil
}
