package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateCommonSettings_NilConfig(t *testing.T) {
	err := ValidateCommonSettings("r18dev", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "r18dev: config is nil")
}

func TestValidateCommonSettings_DisabledConfig(t *testing.T) {
	err := ValidateCommonSettings("r18dev", &ScraperSettings{Enabled: false})
	assert.NoError(t, err)
}

func TestValidateCommonSettings_InvalidRateLimit(t *testing.T) {
	err := ValidateCommonSettings("r18dev", &ScraperSettings{Enabled: true, RateLimit: -1})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rate_limit must be non-negative")
}

func TestValidateCommonSettings_InvalidRetryCount(t *testing.T) {
	err := ValidateCommonSettings("javbus", &ScraperSettings{Enabled: true, RetryCount: -1})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "retry_count must be non-negative")
}

func TestValidateCommonSettings_InvalidTimeout(t *testing.T) {
	err := ValidateCommonSettings("dmm", &ScraperSettings{Enabled: true, Timeout: -1})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout must be non-negative")
}

func TestValidateCommonSettings_ValidConfig(t *testing.T) {
	err := ValidateCommonSettings("r18dev", &ScraperSettings{Enabled: true, RateLimit: 100, RetryCount: 3, Timeout: 30})
	assert.NoError(t, err)
}

func TestValidateCommonSettings_ScraperNamePrefix(t *testing.T) {
	testCases := []struct {
		name        string
		scraperName string
		expected    string
	}{
		{"r18dev prefix", "r18dev", "r18dev:"},
		{"javbus prefix", "javbus", "javbus:"},
		{"dmm prefix", "dmm", "dmm:"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateCommonSettings(tc.scraperName, nil)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.expected)
		})
	}
}
