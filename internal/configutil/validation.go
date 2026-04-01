package configutil

import (
	"fmt"
	"net/url"
	"strings"
)

// FlareSolverrConfig holds FlareSolverr configuration for bypassing Cloudflare
// Copied from internal/config to avoid circular dependency
type FlareSolverrConfig struct {
	Enabled    bool   `yaml:"enabled" json:"enabled"`         // Enable FlareSolverr for bypassing Cloudflare
	URL        string `yaml:"url" json:"url"`                 // FlareSolverr endpoint (default: http://localhost:8191/v1)
	Timeout    int    `yaml:"timeout" json:"timeout"`         // Request timeout in seconds (default: 30)
	MaxRetries int    `yaml:"max_retries" json:"max_retries"` // Max retry attempts for FlareSolverr calls (default: 3)
	SessionTTL int    `yaml:"session_ttl" json:"session_ttl"` // Session TTL in seconds (default: 300)
}

// ValidateFlareSolverrConfig validates FlareSolverr configuration
func ValidateFlareSolverrConfig(path string, cfg FlareSolverrConfig) error {
	if !cfg.Enabled {
		return nil
	}
	if cfg.URL == "" {
		return fmt.Errorf("%s.url is required when flaresolverr is enabled", path)
	}
	if cfg.Timeout < 1 || cfg.Timeout > 300 {
		return fmt.Errorf("%s.timeout must be between 1 and 300", path)
	}
	if cfg.MaxRetries < 0 || cfg.MaxRetries > 10 {
		return fmt.Errorf("%s.max_retries must be between 0 and 10", path)
	}
	if cfg.SessionTTL < 60 || cfg.SessionTTL > 3600 {
		return fmt.Errorf("%s.session_ttl must be between 60 and 3600", path)
	}
	return nil
}

// ValidateRequestDelay validates request delay configuration
func ValidateRequestDelay(path string, delay int) error {
	if delay < 0 {
		return fmt.Errorf("%s.request_delay must be non-negative", path)
	}
	return nil
}

// ValidateHTTPBaseURL validates that a base URL is a valid HTTP or HTTPS URL.
func ValidateHTTPBaseURL(path, raw string) error {
	if raw == "" {
		return nil // Base URL is optional
	}
	raw = strings.TrimSpace(raw)
	if !strings.HasPrefix(raw, "http://") && !strings.HasPrefix(raw, "https://") {
		return fmt.Errorf("%s must be a valid HTTP or HTTPS URL", path)
	}
	_, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("%s must be a valid URL: %w", path, err)
	}
	return nil
}
