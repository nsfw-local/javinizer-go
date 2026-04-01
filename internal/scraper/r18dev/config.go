package r18dev

import (
	"fmt"
	"strings"

	"github.com/javinizer/javinizer-go/internal/config"
	"github.com/javinizer/javinizer-go/internal/scraperutil"
)

// R18DevConfig holds R18.dev scraper configuration.
// YAML tags are defined here for unmarshaling via config.ScrapersConfig.
type R18DevConfig struct {
	Enabled           bool                      `yaml:"enabled" json:"enabled"`
	Language          string                    `yaml:"language" json:"language"`                                 // Language code: en, ja (default: en)
	RequestDelay      int                       `yaml:"request_delay" json:"request_delay"`                       // Delay between requests in milliseconds (0 = no delay)
	MaxRetries        int                       `yaml:"max_retries" json:"max_retries"`                           // Maximum number of retry attempts for rate-limited requests
	RespectRetryAfter bool                      `yaml:"respect_retry_after" json:"respect_retry_after"`           // Whether to respect Retry-After header from server
	UserAgent         string                    `yaml:"user_agent" json:"user_agent"`                             // Custom User-Agent for this scraper
	Proxy             *config.ProxyConfig       `yaml:"proxy,omitempty" json:"proxy,omitempty"`                   // Optional scraper-specific proxy override
	DownloadProxy     *config.ProxyConfig       `yaml:"download_proxy,omitempty" json:"download_proxy,omitempty"` // Optional scraper-specific download proxy override
	Priority          int                       `yaml:"priority" json:"priority"`                                 // Scraper's priority (higher = higher priority)
	FlareSolverr      config.FlareSolverrConfig `yaml:"flaresolverr" json:"flaresolverr"`                         // FlareSolverr config for Cloudflare bypass
}

func init() {
	scraperutil.RegisterValidator("r18dev", func(a any) error {
		return (&R18DevConfig{}).ValidateConfig(a.(*config.ScraperSettings))
	})
	// PLUGIN-01: Register config field accessor for registry-based iteration
	// Note: getter methods were removed in Phase 01. The normalize function will
	// fall back to scraperConfigs map directly if this returns nil.
	scraperutil.RegisterScraperConfig("r18dev", func(a any) any {
		return nil
	})
	// TASK 5: Register ConfigFactory for UnmarshalYAML
	scraperutil.RegisterConfigFactory("r18dev", func() any {
		return &R18DevConfig{}
	})
	// TASK 3: Register flatten function for registry-based type conversion
	scraperutil.RegisterFlattenFunc("r18dev", func(cfg any) any {
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
		// Use type assertion to access R18Dev-specific RespectRetryAfter field
		if r18Cfg, ok := cfg.(*R18DevConfig); ok {
			return &config.ScraperSettings{
				Enabled:   c.IsEnabled(),
				Language:  "", // r18dev scraper type doesn't have language field
				RateLimit: c.GetRequestDelay(),
				Extra: map[string]any{
					"respect_retry_after": r18Cfg.RespectRetryAfter,
				},
				Proxy:         proxyVal,
				DownloadProxy: downloadProxyVal,
				FlareSolverr:  r18Cfg.FlareSolverr,
			}
		}
		return nil
	})
}

// IsEnabled implements scraperutil.ScraperConfigInterface.
func (c *R18DevConfig) IsEnabled() bool { return c.Enabled }

// GetUserAgent implements scraperutil.ScraperConfigInterface.
func (c *R18DevConfig) GetUserAgent() string { return c.UserAgent }

// GetRequestDelay implements scraperutil.ScraperConfigInterface.
func (c *R18DevConfig) GetRequestDelay() int { return c.RequestDelay }

// GetMaxRetries implements scraperutil.ScraperConfigInterface.
func (c *R18DevConfig) GetMaxRetries() int { return 0 }

// GetProxy implements scraperutil.ScraperConfigInterface.
func (c *R18DevConfig) GetProxy() any { return c.Proxy }

// GetDownloadProxy implements scraperutil.ScraperConfigInterface.
func (c *R18DevConfig) GetDownloadProxy() any { return c.DownloadProxy }

// ValidateConfig implements config.ConfigValidator for R18DevConfig.
func (c *R18DevConfig) ValidateConfig(sc *config.ScraperSettings) error {
	if sc == nil {
		return fmt.Errorf("r18dev: config is nil")
	}
	if !sc.Enabled {
		return nil // Disabled is valid
	}
	// Validate language if set
	switch strings.ToLower(strings.TrimSpace(sc.Language)) {
	case "", "en":
		// Valid
	case "ja":
		// Valid
	default:
		return fmt.Errorf("r18dev: language must be 'en' or 'ja', got %q", sc.Language)
	}
	// Validate rate limit
	if sc.RateLimit < 0 {
		return fmt.Errorf("r18dev: rate_limit must be non-negative, got %d", sc.RateLimit)
	}
	// Validate retry count
	if sc.RetryCount < 0 {
		return fmt.Errorf("r18dev: retry_count must be non-negative, got %d", sc.RetryCount)
	}
	// Validate timeout
	if sc.Timeout < 0 {
		return fmt.Errorf("r18dev: timeout must be non-negative, got %d", sc.Timeout)
	}
	// Validate FlareSolverr config if enabled
	if sc.FlareSolverr.Enabled {
		if sc.FlareSolverr.URL == "" {
			return fmt.Errorf("r18dev.flaresolverr.url is required when flaresolverr is enabled")
		}
		if sc.FlareSolverr.Timeout < 1 || sc.FlareSolverr.Timeout > 300 {
			return fmt.Errorf("r18dev.flaresolverr.timeout must be between 1 and 300")
		}
		if sc.FlareSolverr.MaxRetries < 0 || sc.FlareSolverr.MaxRetries > 10 {
			return fmt.Errorf("r18dev.flaresolverr.max_retries must be between 0 and 10")
		}
		if sc.FlareSolverr.SessionTTL < 60 || sc.FlareSolverr.SessionTTL > 3600 {
			return fmt.Errorf("r18dev.flaresolverr.session_ttl must be between 60 and 3600")
		}
	}
	return nil
}
