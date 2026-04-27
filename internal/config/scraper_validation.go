package config

import "fmt"

func ValidateCommonSettings(scraperName string, sc *ScraperSettings) error {
	if sc == nil {
		return fmt.Errorf("%s: config is nil", scraperName)
	}
	if !sc.Enabled {
		return nil
	}
	if sc.RateLimit < 0 {
		return fmt.Errorf("%s: rate_limit must be non-negative, got %d", scraperName, sc.RateLimit)
	}
	if sc.RetryCount < 0 {
		return fmt.Errorf("%s: retry_count must be non-negative, got %d", scraperName, sc.RetryCount)
	}
	if sc.Timeout < 0 {
		return fmt.Errorf("%s: timeout must be non-negative, got %d", scraperName, sc.Timeout)
	}
	return nil
}
