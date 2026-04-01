package libredmm

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
			errMsg:  "libredmm: config is nil",
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
			errMsg:  "libredmm: rate_limit must be non-negative, got -1",
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
			errMsg:  "libredmm: timeout must be non-negative, got -1",
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
				BaseURL: "https://libredmm.com",
			},
			wantErr: false,
		},
		{
			name: "valid https base URL is valid",
			cfg: &config.ScraperSettings{
				Enabled: true,
				BaseURL: "https://libredmm.com",
			},
			wantErr: false,
		},
		{
			name: "invalid base URL (no protocol) is invalid",
			cfg: &config.ScraperSettings{
				Enabled: true,
				BaseURL: "libredmm.com",
			},
			wantErr: true,
			errMsg:  "libredmm.base_url must be a valid HTTP or HTTPS URL",
		},
		{
			name: "all valid fields",
			cfg: &config.ScraperSettings{
				Enabled:   true,
				RateLimit: 500,
				Timeout:   60,
			},
			wantErr: false,
		},
	}

	c := &LibreDMMConfig{}
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
