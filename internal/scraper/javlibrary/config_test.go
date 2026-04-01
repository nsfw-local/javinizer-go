package javlibrary

import (
	"testing"

	"github.com/javinizer/javinizer-go/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.ScraperSettings
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil config returns error",
			cfg:     nil,
			wantErr: true,
			errMsg:  "javlibrary: config is nil",
		},
		{
			name: "disabled scraper is valid",
			cfg: &config.ScraperSettings{
				Enabled: false,
			},
			wantErr: false,
		},
		{
			name: "language en is valid",
			cfg: &config.ScraperSettings{
				Enabled:  true,
				Language: "en",
			},
			wantErr: false,
		},
		{
			name: "language ja is valid",
			cfg: &config.ScraperSettings{
				Enabled:  true,
				Language: "ja",
			},
			wantErr: false,
		},
		{
			name: "language cn is valid",
			cfg: &config.ScraperSettings{
				Enabled:  true,
				Language: "cn",
			},
			wantErr: false,
		},
		{
			name: "language tw is valid",
			cfg: &config.ScraperSettings{
				Enabled:  true,
				Language: "tw",
			},
			wantErr: false,
		},
		{
			name: "language empty is valid (defaults to en)",
			cfg: &config.ScraperSettings{
				Enabled:  true,
				Language: "",
			},
			wantErr: false,
		},
		{
			name: "language case insensitive EN is valid",
			cfg: &config.ScraperSettings{
				Enabled:  true,
				Language: "EN",
			},
			wantErr: false,
		},
		{
			name: "language case insensitive JA is valid",
			cfg: &config.ScraperSettings{
				Enabled:  true,
				Language: "Ja",
			},
			wantErr: false,
		},
		{
			name: "language de is invalid",
			cfg: &config.ScraperSettings{
				Enabled:  true,
				Language: "de",
			},
			wantErr: true,
			errMsg:  "javlibrary: language must be 'en', 'ja', 'cn', or 'tw', got \"de\"",
		},
		{
			name: "language fr is invalid",
			cfg: &config.ScraperSettings{
				Enabled:  true,
				Language: "fr",
			},
			wantErr: true,
			errMsg:  "javlibrary: language must be 'en', 'ja', 'cn', or 'tw', got \"fr\"",
		},
		{
			name: "RateLimit -1 is invalid",
			cfg: &config.ScraperSettings{
				Enabled:   true,
				RateLimit: -1,
			},
			wantErr: true,
			errMsg:  "javlibrary: rate_limit must be non-negative, got -1",
		},
		{
			name: "RateLimit 0 is valid",
			cfg: &config.ScraperSettings{
				Enabled:   true,
				RateLimit: 0,
			},
			wantErr: false,
		},
		{
			name: "RateLimit 1000 is valid",
			cfg: &config.ScraperSettings{
				Enabled:   true,
				RateLimit: 1000,
			},
			wantErr: false,
		},
		{
			name: "Timeout -1 is invalid",
			cfg: &config.ScraperSettings{
				Enabled: true,
				Timeout: -1,
			},
			wantErr: true,
			errMsg:  "javlibrary: timeout must be non-negative, got -1",
		},
		{
			name: "Timeout 0 is valid",
			cfg: &config.ScraperSettings{
				Enabled: true,
				Timeout: 0,
			},
			wantErr: false,
		},
		{
			name: "Timeout 30 is valid",
			cfg: &config.ScraperSettings{
				Enabled: true,
				Timeout: 30,
			},
			wantErr: false,
		},
		{
			name: "valid http base URL is valid",
			cfg: &config.ScraperSettings{
				Enabled: true,
				BaseURL: "http://javlibrary.com",
			},
			wantErr: false,
		},
		{
			name: "valid https base URL is valid",
			cfg: &config.ScraperSettings{
				Enabled: true,
				BaseURL: "https://javlibrary.com",
			},
			wantErr: false,
		},
		{
			name: "invalid base URL (no protocol) is invalid",
			cfg: &config.ScraperSettings{
				Enabled: true,
				BaseURL: "javlibrary.com",
			},
			wantErr: true,
			errMsg:  "javlibrary.base_url must be a valid HTTP or HTTPS URL",
		},
		{
			name: "FlareSolverr enabled without URL is invalid",
			cfg: &config.ScraperSettings{
				Enabled: true,
				FlareSolverr: config.FlareSolverrConfig{
					Enabled: true,
					URL:     "",
				},
			},
			wantErr: true,
			errMsg:  "javlibrary.flaresolverr.url is required when flaresolverr is enabled",
		},
		{
			name: "FlareSolverr enabled with URL is valid",
			cfg: &config.ScraperSettings{
				Enabled: true,
				FlareSolverr: config.FlareSolverrConfig{
					Enabled:    true,
					URL:        "http://localhost:8191/v1",
					Timeout:    30,
					MaxRetries: 3,
					SessionTTL: 300,
				},
			},
			wantErr: false,
		},
		{
			name: "FlareSolverr timeout out of range is invalid",
			cfg: &config.ScraperSettings{
				Enabled: true,
				FlareSolverr: config.FlareSolverrConfig{
					Enabled:    true,
					URL:        "http://localhost:8191/v1",
					Timeout:    0,
					MaxRetries: 3,
					SessionTTL: 300,
				},
			},
			wantErr: true,
			errMsg:  "javlibrary.flaresolverr.timeout must be between 1 and 300",
		},
		{
			name: "all valid fields",
			cfg: &config.ScraperSettings{
				Enabled:   true,
				Language:  "ja",
				RateLimit: 500,
				Timeout:   60,
			},
			wantErr: false,
		},
	}

	c := &JavLibraryConfig{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := c.ValidateConfig(tt.cfg)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Equal(t, tt.errMsg, err.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
