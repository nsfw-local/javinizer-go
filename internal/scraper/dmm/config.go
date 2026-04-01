package dmm

import (
	"fmt"

	"github.com/javinizer/javinizer-go/internal/config"
	"github.com/javinizer/javinizer-go/internal/scraperutil"
)

// Config holds DMM/Fanza scraper configuration.
// YAML tags are defined here for unmarshaling via config.ScrapersConfig.
type DMMConfig struct {
	Enabled        bool                `yaml:"enabled" json:"enabled"`
	ScrapeActress  bool                `yaml:"scrape_actress" json:"scrape_actress"`
	EnableBrowser  bool                `yaml:"enable_browser" json:"enable_browser"`
	BrowserTimeout int                 `yaml:"browser_timeout" json:"browser_timeout"`
	RequestDelay   int                 `yaml:"request_delay" json:"request_delay"`
	MaxRetries     int                 `yaml:"max_retries" json:"max_retries"`
	UserAgent      string              `yaml:"user_agent" json:"user_agent"`
	Proxy          *config.ProxyConfig `yaml:"proxy,omitempty" json:"proxy,omitempty"`
	DownloadProxy  *config.ProxyConfig `yaml:"download_proxy,omitempty" json:"download_proxy,omitempty"`
	Priority       int                 `yaml:"priority" json:"priority"` // Scraper's priority (higher = higher priority)
}

func init() {
	scraperutil.RegisterValidator("dmm", func(a any) error {
		return (&DMMConfig{}).ValidateConfig(a.(*config.ScraperSettings))
	})
	// PLUGIN-01: Register config field accessor for registry-based iteration
	// Note: getter methods were removed in Phase 01. The normalize function will
	// fall back to scraperConfigs map directly if this returns nil.
	scraperutil.RegisterScraperConfig("dmm", func(a any) any {
		return nil
	})
	// TASK 5: Register ConfigFactory for UnmarshalYAML
	scraperutil.RegisterConfigFactory("dmm", func() any {
		return &DMMConfig{}
	})
	// TASK 3: Register flatten function for registry-based type conversion
	scraperutil.RegisterFlattenFunc("dmm", func(cfg any) any {
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
		// Use type assertion to access DMM-specific fields not in ScraperConfigInterface
		if dmmCfg, ok := cfg.(*DMMConfig); ok {
			return &config.ScraperSettings{
				Enabled:   c.IsEnabled(),
				RateLimit: c.GetRequestDelay(),
				Extra: map[string]any{
					"scrape_actress":  dmmCfg.ScrapeActress,
					"enable_browser":  dmmCfg.EnableBrowser,
					"browser_timeout": dmmCfg.BrowserTimeout,
				},
				Proxy:         proxyVal,
				DownloadProxy: downloadProxyVal,
			}
		}
		return nil
	})
}

// IsEnabled implements scraperutil.ScraperConfigInterface.
func (c *DMMConfig) IsEnabled() bool { return c.Enabled }

// GetUserAgent implements scraperutil.ScraperConfigInterface.
func (c *DMMConfig) GetUserAgent() string { return c.UserAgent }

// GetRequestDelay implements scraperutil.ScraperConfigInterface.
func (c *DMMConfig) GetRequestDelay() int { return c.RequestDelay }

// GetMaxRetries implements scraperutil.ScraperConfigInterface.
func (c *DMMConfig) GetMaxRetries() int { return c.MaxRetries }

// GetProxy implements scraperutil.ScraperConfigInterface.
func (c *DMMConfig) GetProxy() any { return c.Proxy }

// GetDownloadProxy implements scraperutil.ScraperConfigInterface.
func (c *DMMConfig) GetDownloadProxy() any { return c.DownloadProxy }

// ValidateConfig implements config.ConfigValidator for DMMConfig.
func (c *DMMConfig) ValidateConfig(sc *config.ScraperSettings) error {
	if sc == nil {
		return fmt.Errorf("dmm: config is nil")
	}
	if !sc.Enabled {
		return nil // Disabled is valid
	}
	// Validate rate limit
	if sc.RateLimit < 0 {
		return fmt.Errorf("dmm: rate_limit must be non-negative, got %d", sc.RateLimit)
	}
	// Validate retry count
	if sc.RetryCount < 0 {
		return fmt.Errorf("dmm: retry_count must be non-negative, got %d", sc.RetryCount)
	}
	// Validate timeout
	if sc.Timeout < 0 {
		return fmt.Errorf("dmm: timeout must be non-negative, got %d", sc.Timeout)
	}
	// Validate browser timeout if browser is enabled
	if sc.GetBoolExtra("enable_browser", false) {
		browserTimeout := sc.GetIntExtra("browser_timeout", 30)
		if browserTimeout < 1 {
			return fmt.Errorf("dmm: browser_timeout must be at least 1 second")
		}
	}
	return nil
}
