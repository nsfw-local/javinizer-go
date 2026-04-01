package dmm

import (
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/javinizer/javinizer-go/internal/config"
	"github.com/javinizer/javinizer-go/internal/httpclient"
)

// NewHTTPClient creates an HTTP client for the DMM scraper.
// HTTP-01: Per-scraper HTTP client ownership.
// Returns client, effective proxyProfile (for browser use), and error.
func NewHTTPClient(cfg *config.ScraperSettings, globalProxy *config.ProxyConfig, globalFlareSolverr config.FlareSolverrConfig) (*resty.Client, *config.ProxyProfile, error) {
	// Handle nil globalProxy to avoid dereference panic
	globalProxyVal := config.ProxyConfig{}
	if globalProxy != nil {
		globalProxyVal = *globalProxy
	}
	// Resolve proxy per-scraper (HTTP-02)
	proxyCfg := config.ResolveScraperProxy(globalProxyVal, cfg.Proxy)

	// Use timeout from ScraperSettings, default to 30s
	timeout := time.Duration(cfg.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	// Use RetryCount from ScraperConfig, default to 3
	retryCount := cfg.RetryCount
	if retryCount == 0 {
		retryCount = 3
	}

	client, err := httpclient.NewRestyClient(proxyCfg, timeout, retryCount)
	if err != nil {
		return nil, nil, err
	}

	// Apply UserAgent from ScraperConfig
	userAgent := config.ResolveScraperUserAgent(cfg.UserAgent)
	client.SetHeader("User-Agent", userAgent)
	client.SetHeader("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	client.SetHeader("Accept-Language", "en-US,en;q=0.9,ja;q=0.8")
	client.SetHeader("Accept-Encoding", "gzip, deflate, br")
	client.SetHeader("Connection", "keep-alive")
	client.SetHeader("Upgrade-Insecure-Requests", "1")
	client.SetHeader("Cookie", "age_check_done=1; cklg=ja")

	return client, proxyCfg, nil
}

// NewHTTPClientWithDefaults creates an HTTP client using DMM-specific defaults.
// Deprecated: This function is kept for backward compatibility but is no longer used.
// Scraper.New() now constructs ScraperConfig from ScraperSettings directly.
func NewHTTPClientWithDefaults(cfg *config.Config) (*resty.Client, *config.ProxyProfile, error) {
	// This function is deprecated. Use NewHTTPClient with ScraperSettings instead.
	// Keeping this stub to avoid breaking any external callers.
	return nil, nil, fmt.Errorf("NewHTTPClientWithDefaults is deprecated")
}
