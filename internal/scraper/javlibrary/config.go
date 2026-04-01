package javlibrary

import (
	"fmt"
	"strings"

	"github.com/javinizer/javinizer-go/internal/config"
	"github.com/javinizer/javinizer-go/internal/configutil"
	"github.com/javinizer/javinizer-go/internal/scraperutil"
)

// Config holds JavLibrary scraper configuration.
// YAML tags are defined here for unmarshaling via config.ScrapersConfig.
type JavLibraryConfig struct {
	Enabled       bool                      `yaml:"enabled" json:"enabled"`
	Language      string                    `yaml:"language" json:"language"`
	RequestDelay  int                       `yaml:"request_delay" json:"request_delay"`
	BaseURL       string                    `yaml:"base_url" json:"base_url"`
	CfClearance   string                    `yaml:"cf_clearance" json:"cf_clearance"` // Cloudflare clearance cookie
	CfBm          string                    `yaml:"cf_bm" json:"cf_bm"`               // Cloudflare BM cookie
	UserAgent     string                    `yaml:"user_agent" json:"user_agent"`
	Proxy         *config.ProxyConfig       `yaml:"proxy,omitempty" json:"proxy,omitempty"`
	DownloadProxy *config.ProxyConfig       `yaml:"download_proxy,omitempty" json:"download_proxy,omitempty"`
	Priority      int                       `yaml:"priority" json:"priority"`         // Scraper's priority (higher = higher priority)
	FlareSolverr  config.FlareSolverrConfig `yaml:"flaresolverr" json:"flaresolverr"` // FlareSolverr config for Cloudflare bypass
}

func init() {
	scraperutil.RegisterValidator("javlibrary", func(a any) error {
		return (&JavLibraryConfig{}).ValidateConfig(a.(*config.ScraperSettings))
	})
	// PLUGIN-01: Register config field accessor for registry-based iteration
	// Note: getter methods were removed in Phase 01. The normalize function will
	// fall back to scraperConfigs map directly if this returns nil.
	scraperutil.RegisterScraperConfig("javlibrary", func(a any) any {
		return nil
	})
	// TASK 5: Register ConfigFactory for UnmarshalYAML
	scraperutil.RegisterConfigFactory("javlibrary", func() any {
		return &JavLibraryConfig{}
	})
	// TASK 3: Register flatten function for registry-based type conversion
	scraperutil.RegisterFlattenFunc("javlibrary", func(cfg any) any {
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
		// Use type assertion to access JavLibrary-specific fields
		if jlCfg, ok := cfg.(*JavLibraryConfig); ok {
			return &config.ScraperSettings{
				Enabled:   c.IsEnabled(),
				Language:  jlCfg.Language,
				RateLimit: c.GetRequestDelay(),
				Extra: map[string]any{
					"base_url":     jlCfg.BaseURL,
					"cf_clearance": jlCfg.CfClearance,
					"cf_bm":        jlCfg.CfBm,
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
func (c *JavLibraryConfig) IsEnabled() bool { return c.Enabled }

// GetUserAgent implements scraperutil.ScraperConfigInterface.
func (c *JavLibraryConfig) GetUserAgent() string { return c.UserAgent }

// GetRequestDelay implements scraperutil.ScraperConfigInterface.
func (c *JavLibraryConfig) GetRequestDelay() int { return c.RequestDelay }

// GetMaxRetries implements scraperutil.ScraperConfigInterface.
func (c *JavLibraryConfig) GetMaxRetries() int { return 0 }

// GetProxy implements scraperutil.ScraperConfigInterface.
func (c *JavLibraryConfig) GetProxy() any { return c.Proxy }

// GetDownloadProxy implements scraperutil.ScraperConfigInterface.
func (c *JavLibraryConfig) GetDownloadProxy() any { return c.DownloadProxy }

// ValidateConfig implements config.ConfigValidator for JavLibraryConfig.
func (c *JavLibraryConfig) ValidateConfig(sc *config.ScraperSettings) error {
	if sc == nil {
		return fmt.Errorf("javlibrary: config is nil")
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
	case "cn":
		// Valid
	case "tw":
		// Valid
	default:
		return fmt.Errorf("javlibrary: language must be 'en', 'ja', 'cn', or 'tw', got %q", sc.Language)
	}
	// Validate rate limit
	if sc.RateLimit < 0 {
		return fmt.Errorf("javlibrary: rate_limit must be non-negative, got %d", sc.RateLimit)
	}
	// Validate timeout
	if sc.Timeout < 0 {
		return fmt.Errorf("javlibrary: timeout must be non-negative, got %d", sc.Timeout)
	}
	// Validate base URL if set
	if err := configutil.ValidateHTTPBaseURL("javlibrary.base_url", sc.BaseURL); err != nil {
		return err
	}
	// Validate FlareSolverr config if enabled
	if sc.FlareSolverr.Enabled {
		if sc.FlareSolverr.URL == "" {
			return fmt.Errorf("javlibrary.flaresolverr.url is required when flaresolverr is enabled")
		}
		if sc.FlareSolverr.Timeout < 1 || sc.FlareSolverr.Timeout > 300 {
			return fmt.Errorf("javlibrary.flaresolverr.timeout must be between 1 and 300")
		}
		if sc.FlareSolverr.MaxRetries < 0 || sc.FlareSolverr.MaxRetries > 10 {
			return fmt.Errorf("javlibrary.flaresolverr.max_retries must be between 0 and 10")
		}
		if sc.FlareSolverr.SessionTTL < 60 || sc.FlareSolverr.SessionTTL > 3600 {
			return fmt.Errorf("javlibrary.flaresolverr.session_ttl must be between 60 and 3600")
		}
	}
	return nil
}
