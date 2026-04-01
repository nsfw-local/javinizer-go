package libredmm

import (
	"fmt"

	"github.com/javinizer/javinizer-go/internal/config"
	"github.com/javinizer/javinizer-go/internal/configutil"
	"github.com/javinizer/javinizer-go/internal/scraperutil"
)

// Config holds LibreDMM scraper configuration.
// YAML tags are defined here for unmarshaling via config.ScrapersConfig.
type LibreDMMConfig struct {
	Enabled       bool                      `yaml:"enabled" json:"enabled"`
	RequestDelay  int                       `yaml:"request_delay" json:"request_delay"`
	BaseURL       string                    `yaml:"base_url" json:"base_url"`
	UserAgent     string                    `yaml:"user_agent" json:"user_agent"`
	Proxy         *config.ProxyConfig       `yaml:"proxy,omitempty" json:"proxy,omitempty"`
	DownloadProxy *config.ProxyConfig       `yaml:"download_proxy,omitempty" json:"download_proxy,omitempty"`
	Priority      int                       `yaml:"priority" json:"priority"` // Scraper's priority (higher = higher priority)
	FlareSolverr  config.FlareSolverrConfig `yaml:"flaresolverr" json:"flaresolverr"`
}

func init() {
	scraperutil.RegisterValidator("libredmm", func(a any) error {
		return (&LibreDMMConfig{}).ValidateConfig(a.(*config.ScraperSettings))
	})
	// PLUGIN-01: Register config field accessor for registry-based iteration
	// Note: getter methods were removed in Phase 01. The normalize function will
	// fall back to scraperConfigs map directly if this returns nil.
	scraperutil.RegisterScraperConfig("libredmm", func(a any) any {
		return nil
	})
	// TASK 5: Register ConfigFactory for UnmarshalYAML
	scraperutil.RegisterConfigFactory("libredmm", func() any {
		return &LibreDMMConfig{}
	})
	// TASK 3: Register flatten function for registry-based type conversion
	scraperutil.RegisterFlattenFunc("libredmm", func(cfg any) any {
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
		// Use type assertion to access LibreDMM-specific fields
		if libredmmCfg, ok := cfg.(*LibreDMMConfig); ok {
			return &config.ScraperSettings{
				Enabled:   c.IsEnabled(),
				RateLimit: c.GetRequestDelay(),
				Extra: map[string]any{
					"base_url": "https://www.libredmm.com",
				},
				Proxy:         proxyVal,
				DownloadProxy: downloadProxyVal,
				FlareSolverr:  libredmmCfg.FlareSolverr,
			}
		}
		return nil
	})
}

// IsEnabled implements scraperutil.ScraperConfigInterface.
func (c *LibreDMMConfig) IsEnabled() bool { return c.Enabled }

// GetUserAgent implements scraperutil.ScraperConfigInterface.
func (c *LibreDMMConfig) GetUserAgent() string { return c.UserAgent }

// GetRequestDelay implements scraperutil.ScraperConfigInterface.
func (c *LibreDMMConfig) GetRequestDelay() int { return c.RequestDelay }

// GetMaxRetries implements scraperutil.ScraperConfigInterface.
func (c *LibreDMMConfig) GetMaxRetries() int { return 0 }

// GetProxy implements scraperutil.ScraperConfigInterface.
func (c *LibreDMMConfig) GetProxy() any { return c.Proxy }

// GetDownloadProxy implements scraperutil.ScraperConfigInterface.
func (c *LibreDMMConfig) GetDownloadProxy() any { return c.DownloadProxy }

// ValidateConfig implements config.ConfigValidator for LibreDMMConfig.
func (c *LibreDMMConfig) ValidateConfig(sc *config.ScraperSettings) error {
	if sc == nil {
		return fmt.Errorf("libredmm: config is nil")
	}
	if !sc.Enabled {
		return nil // Disabled is valid
	}
	// Validate rate limit
	if sc.RateLimit < 0 {
		return fmt.Errorf("libredmm: rate_limit must be non-negative, got %d", sc.RateLimit)
	}
	// Validate timeout
	if sc.Timeout < 0 {
		return fmt.Errorf("libredmm: timeout must be non-negative, got %d", sc.Timeout)
	}
	// Validate base URL if set
	if err := configutil.ValidateHTTPBaseURL("libredmm.base_url", sc.BaseURL); err != nil {
		return err
	}
	return nil
}
