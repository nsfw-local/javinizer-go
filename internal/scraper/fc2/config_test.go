package fc2

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
			errMsg:  "fc2: config is nil",
		},
		{
			name: "disabled scraper is valid",
			cfg: &config.ScraperSettings{
				Enabled: false,
			},
			wantErr: false,
		},
		{
			name: "RateLimit -1 is invalid",
			cfg: &config.ScraperSettings{
				Enabled:   true,
				RateLimit: -1,
			},
			wantErr: true,
			errMsg:  "fc2: rate_limit must be non-negative, got -1",
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
			name: "RetryCount -1 is invalid",
			cfg: &config.ScraperSettings{
				Enabled:    true,
				RetryCount: -1,
			},
			wantErr: true,
			errMsg:  "fc2: retry_count must be non-negative, got -1",
		},
		{
			name: "RetryCount 0 is valid",
			cfg: &config.ScraperSettings{
				Enabled:    true,
				RetryCount: 0,
			},
			wantErr: false,
		},
		{
			name: "RetryCount 3 is valid",
			cfg: &config.ScraperSettings{
				Enabled:    true,
				RetryCount: 3,
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
			errMsg:  "fc2: timeout must be non-negative, got -1",
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
				BaseURL: "https://fc2.com",
			},
			wantErr: false,
		},
		{
			name: "invalid base URL (no protocol) is invalid",
			cfg: &config.ScraperSettings{
				Enabled: true,
				BaseURL: "fc2.com",
			},
			wantErr: true,
			errMsg:  "fc2.base_url must be a valid HTTP or HTTPS URL",
		},
		{
			name: "FlareSolverr enabled without name",
			cfg: &config.ScraperSettings{
				Enabled:         true,
				UseFlareSolverr: true,
			},
			wantErr: false,
		},
		{
			name: "FlareSolverr enabled is valid",
			cfg: &config.ScraperSettings{
				Enabled:         true,
				UseFlareSolverr: true,
			},
			wantErr: false,
		},
		{
			name: "all valid fields",
			cfg: &config.ScraperSettings{
				Enabled:    true,
				RateLimit:  500,
				RetryCount: 3,
				Timeout:    60,
			},
			wantErr: false,
		},
	}

	c := &FC2Config{}
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
