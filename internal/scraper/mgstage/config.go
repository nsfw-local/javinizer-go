package mgstage

import (
	"fmt"

	"github.com/javinizer/javinizer-go/internal/config"
	"github.com/javinizer/javinizer-go/internal/scraperutil"
)

// Config holds MGStage scraper configuration.
// YAML tags are defined here for unmarshaling via config.ScrapersConfig.
type MGStageConfig struct {
	Enabled       bool                      `yaml:"enabled" json:"enabled"`
	RequestDelay  int                       `yaml:"request_delay" json:"request_delay"`
	MaxRetries    int                       `yaml:"max_retries" json:"max_retries"`
	UserAgent     string                    `yaml:"user_agent" json:"user_agent"`
	Proxy         *config.ProxyConfig       `yaml:"proxy,omitempty" json:"proxy,omitempty"`
	DownloadProxy *config.ProxyConfig       `yaml:"download_proxy,omitempty" json:"download_proxy,omitempty"`
	Priority      int                       `yaml:"priority" json:"priority"` // Scraper's priority (higher = higher priority)
	FlareSolverr  config.FlareSolverrConfig `yaml:"flaresolverr" json:"flaresolverr"`
}

func init() {
	scraperutil.RegisterValidator("mgstage", func(a any) error {
		return (&MGStageConfig{}).ValidateConfig(a.(*config.ScraperSettings))
	})
	// PLUGIN-01: Register config field accessor for registry-based iteration
	// Note: getter methods were removed in Phase 01. The normalize function will
	// fall back to scraperConfigs map directly if this returns nil.
	scraperutil.RegisterScraperConfig("mgstage", func(a any) any {
		return nil
	})
	// TASK 5: Register ConfigFactory for UnmarshalYAML
	scraperutil.RegisterConfigFactory("mgstage", func() any {
		return &MGStageConfig{}
	})
	// TASK 3: Register flatten function for registry-based type conversion
	scraperutil.RegisterFlattenFunc("mgstage", func(cfg any) any {
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
		// Use type assertion to access MGStage-specific fields
		if mgstageCfg, ok := cfg.(*MGStageConfig); ok {
			return &config.ScraperSettings{
				Enabled:       c.IsEnabled(),
				RateLimit:     c.GetRequestDelay(),
				Extra:         map[string]any{},
				Proxy:         proxyVal,
				DownloadProxy: downloadProxyVal,
				FlareSolverr:  mgstageCfg.FlareSolverr,
			}
		}
		return nil
	})
}

// IsEnabled implements scraperutil.ScraperConfigInterface.
func (c *MGStageConfig) IsEnabled() bool { return c.Enabled }

// GetUserAgent implements scraperutil.ScraperConfigInterface.
func (c *MGStageConfig) GetUserAgent() string { return c.UserAgent }

// GetRequestDelay implements scraperutil.ScraperConfigInterface.
func (c *MGStageConfig) GetRequestDelay() int { return c.RequestDelay }

// GetMaxRetries implements scraperutil.ScraperConfigInterface.
func (c *MGStageConfig) GetMaxRetries() int { return c.MaxRetries }

// GetProxy implements scraperutil.ScraperConfigInterface.
func (c *MGStageConfig) GetProxy() any { return c.Proxy }

// GetDownloadProxy implements scraperutil.ScraperConfigInterface.
func (c *MGStageConfig) GetDownloadProxy() any { return c.DownloadProxy }

// ValidateConfig implements config.ConfigValidator for MGStageConfig.
func (c *MGStageConfig) ValidateConfig(sc *config.ScraperSettings) error {
	if sc == nil {
		return fmt.Errorf("mgstage: config is nil")
	}
	if !sc.Enabled {
		return nil // Disabled is valid
	}
	// Validate rate limit
	if sc.RateLimit < 0 {
		return fmt.Errorf("mgstage: rate_limit must be non-negative, got %d", sc.RateLimit)
	}
	// Validate retry count
	if sc.RetryCount < 0 {
		return fmt.Errorf("mgstage: retry_count must be non-negative, got %d", sc.RetryCount)
	}
	// Validate timeout
	if sc.Timeout < 0 {
		return fmt.Errorf("mgstage: timeout must be non-negative, got %d", sc.Timeout)
	}
	// MGStage doesn't typically use FlareSolverr, but validate if enabled
	if sc.FlareSolverr.Enabled {
		if sc.FlareSolverr.URL == "" {
			return fmt.Errorf("mgstage.flaresolverr.url is required when flaresolverr is enabled")
		}
		if sc.FlareSolverr.Timeout < 1 || sc.FlareSolverr.Timeout > 300 {
			return fmt.Errorf("mgstage.flaresolverr.timeout must be between 1 and 300")
		}
		if sc.FlareSolverr.MaxRetries < 0 || sc.FlareSolverr.MaxRetries > 10 {
			return fmt.Errorf("mgstage.flaresolverr.max_retries must be between 0 and 10")
		}
		if sc.FlareSolverr.SessionTTL < 60 || sc.FlareSolverr.SessionTTL > 3600 {
			return fmt.Errorf("mgstage.flaresolverr.session_ttl must be between 60 and 3600")
		}
	}
	return nil
}
