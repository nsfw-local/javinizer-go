package javlibrary

import (
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/javinizer/javinizer-go/internal/config"
	httpclient "github.com/javinizer/javinizer-go/internal/httpclient"
	"github.com/javinizer/javinizer-go/internal/logging"
)

// NewHTTPClient creates an HTTP client and FlareSolverr for the javlibrary scraper.
// HTTP-01, HTTP-03: Per-scraper HTTP client and FlareSolverr ownership.
// The bool parameter (useFlareSolverr) indicates whether to enable FlareSolverr
// based on scraperCfg.UseFlareSolverr from the javlibrary config.
// Returns client, flaresolverr, and error.
func NewHTTPClient(cfg *config.ScraperSettings, globalProxy *config.ProxyConfig, globalFlareSolverr config.FlareSolverrConfig, useFlareSolverr bool) (*resty.Client, *httpclient.FlareSolverr, error) {
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

	// When useFlareSolverr is true (passed from javlibrary.go based on
	// scraperCfg.UseFlareSolverr), the FlareSolverr client will be initialized.
	// Proxy profile is used directly by NewRestyClientWithFlareSolverr via buildFlareSolverrRequestProxy.
	var client *resty.Client
	var fs *httpclient.FlareSolverr
	var err error

	if useFlareSolverr {
		// Use global proxy profile if scraper-specific proxy URL is empty but global proxy has one
		proxyForFS := proxyCfg
		if proxyCfg.URL == "" {
			globalProfile := config.ResolveGlobalProxy(globalProxyVal)
			if globalProfile != nil && globalProfile.URL != "" {
				proxyForFS = globalProfile
			}
		}

		// Pass FlareSolverr config separately since it's no longer embedded in ProxyConfig
		client, fs, err = httpclient.NewRestyClientWithFlareSolverr(
			proxyForFS,
			cfg.FlareSolverr,
			timeout,
			retryCount,
		)
		if err != nil {
			logging.Errorf("JavLibrary: Failed to create FlareSolverr: %v", err)
			// Fall through to create client without FlareSolverr
			client, err = httpclient.NewRestyClient(proxyCfg, timeout, retryCount)
		}
	} else {
		client, err = httpclient.NewRestyClient(proxyCfg, timeout, retryCount)
	}
	if err != nil {
		return nil, nil, err
	}

	return client, fs, nil
}
