package jav321

import (
	"fmt"

	"github.com/javinizer/javinizer-go/internal/config"
	"github.com/javinizer/javinizer-go/internal/configutil"
	"github.com/javinizer/javinizer-go/internal/scraperutil"
)

// Config holds Jav321 scraper configuration.
// YAML tags are defined here for unmarshaling via config.ScrapersConfig.
type Jav321Config struct {
	Enabled       bool                      `yaml:"enabled" json:"enabled"`
	RequestDelay  int                       `yaml:"request_delay" json:"request_delay"`
	MaxRetries    int                       `yaml:"max_retries" json:"max_retries"`
	BaseURL       string                    `yaml:"base_url" json:"base_url"`
	UserAgent     string                    `yaml:"user_agent" json:"user_agent"`
	Proxy         *config.ProxyConfig       `yaml:"proxy,omitempty" json:"proxy,omitempty"`
	DownloadProxy *config.ProxyConfig       `yaml:"download_proxy,omitempty" json:"download_proxy,omitempty"`
	Priority      int                       `yaml:"priority" json:"priority"`         // Scraper's priority (higher = higher priority)
	FlareSolverr  config.FlareSolverrConfig `yaml:"flaresolverr" json:"flaresolverr"` // FlareSolverr config for Cloudflare bypass
}

func init() {
	scraperutil.RegisterValidator("jav321", func(a any) error {
		return (&Jav321Config{}).ValidateConfig(a.(*config.ScraperSettings))
	})
	// PLUGIN-01: Register config field accessor for registry-based iteration
	// Note: getter methods were removed in Phase 01. The normalize function will
	// fall back to scraperConfigs map directly if this returns nil.
	scraperutil.RegisterScraperConfig("jav321", func(a any) any {
		return nil
	})
	// TASK 5: Register ConfigFactory for UnmarshalYAML
	scraperutil.RegisterConfigFactory("jav321", func() any {
		return &Jav321Config{}
	})
	// TASK 3: Register flatten function for registry-based type conversion
	scraperutil.RegisterFlattenFunc("jav321", func(cfg any) any {
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
		if jlCfg, ok := cfg.(*Jav321Config); ok {
			return &config.ScraperSettings{
				Enabled:   c.IsEnabled(),
				RateLimit: c.GetRequestDelay(),
				Extra: map[string]any{
					"base_url": "https://jp.jav321.com",
				},
				Proxy:         proxyVal,
				DownloadProxy: downloadProxyVal,
				FlareSolverr:  jlCfg.FlareSolverr,
			}
		}
		return nil
	})
}

// IsEnabled implements scraperutil.ScraperConfigInterface.
func (c *Jav321Config) IsEnabled() bool { return c.Enabled }

// GetUserAgent implements scraperutil.ScraperConfigInterface.
func (c *Jav321Config) GetUserAgent() string { return c.UserAgent }

// GetRequestDelay implements scraperutil.ScraperConfigInterface.
func (c *Jav321Config) GetRequestDelay() int { return c.RequestDelay }

// GetMaxRetries implements scraperutil.ScraperConfigInterface.
func (c *Jav321Config) GetMaxRetries() int { return c.MaxRetries }

// GetProxy implements scraperutil.ScraperConfigInterface.
func (c *Jav321Config) GetProxy() any { return c.Proxy }

// GetDownloadProxy implements scraperutil.ScraperConfigInterface.
func (c *Jav321Config) GetDownloadProxy() any { return c.DownloadProxy }

// ValidateConfig implements config.ConfigValidator for Jav321Config.
func (c *Jav321Config) ValidateConfig(sc *config.ScraperSettings) error {
	if sc == nil {
		return fmt.Errorf("jav321: config is nil")
	}
	if !sc.Enabled {
		return nil // Disabled is valid
	}
	// Validate rate limit
	if sc.RateLimit < 0 {
		return fmt.Errorf("jav321: rate_limit must be non-negative, got %d", sc.RateLimit)
	}
	// Validate retry count
	if sc.RetryCount < 0 {
		return fmt.Errorf("jav321: retry_count must be non-negative, got %d", sc.RetryCount)
	}
	// Validate timeout
	if sc.Timeout < 0 {
		return fmt.Errorf("jav321: timeout must be non-negative, got %d", sc.Timeout)
	}
	// Validate base URL if set
	if err := configutil.ValidateHTTPBaseURL("jav321.base_url", sc.BaseURL); err != nil {
		return err
	}
	// Validate FlareSolverr config if enabled
	if sc.FlareSolverr.Enabled {
		if sc.FlareSolverr.URL == "" {
			return fmt.Errorf("jav321.flaresolverr.url is required when flaresolverr is enabled")
		}
		if sc.FlareSolverr.Timeout < 1 || sc.FlareSolverr.Timeout > 300 {
			return fmt.Errorf("jav321.flaresolverr.timeout must be between 1 and 300")
		}
		if sc.FlareSolverr.MaxRetries < 0 || sc.FlareSolverr.MaxRetries > 10 {
			return fmt.Errorf("jav321.flaresolverr.max_retries must be between 0 and 10")
		}
		if sc.FlareSolverr.SessionTTL < 60 || sc.FlareSolverr.SessionTTL > 3600 {
			return fmt.Errorf("jav321.flaresolverr.session_ttl must be between 60 and 3600")
		}
	}
	return nil
}
