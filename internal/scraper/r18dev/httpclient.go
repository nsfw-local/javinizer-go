package r18dev

import (
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/javinizer/javinizer-go/internal/config"
	"github.com/javinizer/javinizer-go/internal/httpclient"
)

// NewHTTPClient creates a new HTTP client for the R18.dev scraper.
// It accepts *ScraperSettings to enable generic HTTP client setup without scraper-name branching.
// HTTP-01: Per-scraper HTTP client ownership.
func NewHTTPClient(cfg *config.ScraperSettings, globalProxy *config.ProxyConfig, globalFlareSolverr config.FlareSolverrConfig) (*resty.Client, error) {
	// Handle nil globalProxy to avoid dereference panic
	globalProxyVal := config.ProxyConfig{}
	if globalProxy != nil {
		globalProxyVal = *globalProxy
	}
	// Resolve proxy per-scraper (HTTP-02)
	proxyCfg := config.ResolveScraperProxy(globalProxyVal, cfg.Proxy)

	// Use timeout from ScraperConfig, default to 30s
	timeout := time.Duration(cfg.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	// Use RetryCount from ScraperConfig, default to 3
	retryCount := cfg.RetryCount
	if retryCount == 0 {
		retryCount = 3
	}

	var client *resty.Client
	var err error

	// HTTP-03: Use global FlareSolverr if scraper has UseFlareSolverr enabled
	if globalFlareSolverr.Enabled && cfg.UseFlareSolverr {
		// Pass FlareSolverr config separately since it's no longer embedded in ProxyConfig
		client, _, err = httpclient.NewRestyClientWithFlareSolverr(
			proxyCfg,
			globalFlareSolverr,
			timeout,
			retryCount,
		)
	} else {
		client, err = httpclient.NewRestyClient(proxyCfg, timeout, retryCount)
	}

	if err != nil {
		return nil, err
	}

	// Apply UserAgent from ScraperConfig
	userAgent := config.ResolveScraperUserAgent(cfg.UserAgent)
	client.SetHeader("User-Agent", userAgent)
	client.SetHeader("Accept", "application/json, text/html, */*")
	client.SetHeader("Referer", "https://r18.dev/")

	return client, nil
}
