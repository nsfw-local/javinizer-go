package httpclient

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/javinizer/javinizer-go/internal/config"
	"golang.org/x/net/proxy"
)

// SanitizeProxyURL removes credentials from proxy URL for safe logging
func SanitizeProxyURL(proxyURL string) string {
	u, err := url.Parse(proxyURL)
	if err != nil {
		return proxyURL // Return as-is if unparseable
	}
	if u.User != nil {
		// Replace user info with [REDACTED]
		u.User = url.User("[REDACTED]")
	}
	return u.String()
}

// NewTransport creates an http.Transport with optional proxy support
func NewTransport(proxyConfig *config.ProxyConfig) (*http.Transport, error) {
	// Clone default transport to preserve Go's safety timeouts
	// (DialContext timeout, TLSHandshakeTimeout, ExpectContinueTimeout, etc.)
	transport := http.DefaultTransport.(*http.Transport).Clone()

	if proxyConfig != nil && proxyConfig.Enabled && proxyConfig.URL != "" {
		proxyURL, err := url.Parse(proxyConfig.URL)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL: %w", err)
		}

		// Handle authentication
		if proxyConfig.Username != "" && proxyConfig.Password != "" {
			proxyURL.User = url.UserPassword(proxyConfig.Username, proxyConfig.Password)
		}

		// Check if SOCKS5
		if proxyURL.Scheme == "socks5" {
			// Use golang.org/x/net/proxy for SOCKS5
			var auth *proxy.Auth
			if proxyConfig.Username != "" && proxyConfig.Password != "" {
				auth = &proxy.Auth{
					User:     proxyConfig.Username,
					Password: proxyConfig.Password,
				}
			}
			dialer, err := proxy.SOCKS5("tcp", proxyURL.Host, auth, proxy.Direct)
			if err != nil {
				return nil, fmt.Errorf("failed to create SOCKS5 dialer: %w", err)
			}
			// Use DialContext to honor request cancellation and timeouts
			// Check if dialer supports DialContext (it does - proxy.Dialer implements ContextDialer)
			if contextDialer, ok := dialer.(proxy.ContextDialer); ok {
				transport.DialContext = contextDialer.DialContext
			} else {
				// Fallback: wrap Dial with context (shouldn't happen with proxy.SOCKS5)
				transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
					return dialer.Dial(network, addr)
				}
			}
			// Clear transport.Proxy to prevent HTTP_PROXY env vars from overriding SOCKS5
			transport.Proxy = nil
		} else {
			// HTTP/HTTPS proxy
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	}

	return transport, nil
}

// NewHTTPClient creates a standard http.Client with proxy support
func NewHTTPClient(proxyConfig *config.ProxyConfig, timeout time.Duration) (*http.Client, error) {
	transport, err := NewTransport(proxyConfig)
	if err != nil {
		return nil, err
	}

	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}, nil
}

// NewRestyClient creates a resty.Client with proxy support
func NewRestyClient(proxyConfig *config.ProxyConfig, timeout time.Duration, retries int) (*resty.Client, error) {
	transport, err := NewTransport(proxyConfig)
	if err != nil {
		return nil, err
	}

	client := resty.New()
	client.SetTimeout(timeout)
	client.SetRetryCount(retries)
	client.SetTransport(transport)

	return client, nil
}
