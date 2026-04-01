package configutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateFlareSolverrConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     FlareSolverrConfig
		wantErr bool
		errMsg  string
	}{
		{
			name:    "disabled is valid",
			cfg:     FlareSolverrConfig{Enabled: false},
			wantErr: false,
		},
		{
			name:    "enabled without URL returns error",
			cfg:     FlareSolverrConfig{Enabled: true, URL: ""},
			wantErr: true,
			errMsg:  "flaresolverr.url is required when flaresolverr is enabled",
		},
		{
			name: "enabled with all valid fields is valid",
			cfg: FlareSolverrConfig{
				Enabled:    true,
				URL:        "http://localhost:8191/v1",
				Timeout:    30,
				MaxRetries: 3,
				SessionTTL: 300,
			},
			wantErr: false,
		},
		{
			name: "timeout 0 returns error",
			cfg: FlareSolverrConfig{
				Enabled:    true,
				URL:        "http://localhost:8191/v1",
				Timeout:    0,
				MaxRetries: 3,
				SessionTTL: 300,
			},
			wantErr: true,
			errMsg:  "flaresolverr.timeout must be between 1 and 300",
		},
		{
			name: "timeout 301 returns error",
			cfg: FlareSolverrConfig{
				Enabled:    true,
				URL:        "http://localhost:8191/v1",
				Timeout:    301,
				MaxRetries: 3,
				SessionTTL: 300,
			},
			wantErr: true,
			errMsg:  "flaresolverr.timeout must be between 1 and 300",
		},
		{
			name: "timeout 1 is valid",
			cfg: FlareSolverrConfig{
				Enabled:    true,
				URL:        "http://localhost:8191/v1",
				Timeout:    1,
				MaxRetries: 3,
				SessionTTL: 300,
			},
			wantErr: false,
		},
		{
			name: "timeout 300 is valid",
			cfg: FlareSolverrConfig{
				Enabled:    true,
				URL:        "http://localhost:8191/v1",
				Timeout:    300,
				MaxRetries: 3,
				SessionTTL: 300,
			},
			wantErr: false,
		},
		{
			name: "max_retries -1 returns error",
			cfg: FlareSolverrConfig{
				Enabled:    true,
				URL:        "http://localhost:8191/v1",
				Timeout:    30,
				MaxRetries: -1,
				SessionTTL: 300,
			},
			wantErr: true,
			errMsg:  "flaresolverr.max_retries must be between 0 and 10",
		},
		{
			name: "max_retries 11 returns error",
			cfg: FlareSolverrConfig{
				Enabled:    true,
				URL:        "http://localhost:8191/v1",
				Timeout:    30,
				MaxRetries: 11,
				SessionTTL: 300,
			},
			wantErr: true,
			errMsg:  "flaresolverr.max_retries must be between 0 and 10",
		},
		{
			name: "max_retries 0 is valid",
			cfg: FlareSolverrConfig{
				Enabled:    true,
				URL:        "http://localhost:8191/v1",
				Timeout:    30,
				MaxRetries: 0,
				SessionTTL: 300,
			},
			wantErr: false,
		},
		{
			name: "max_retries 10 is valid",
			cfg: FlareSolverrConfig{
				Enabled:    true,
				URL:        "http://localhost:8191/v1",
				Timeout:    30,
				MaxRetries: 10,
				SessionTTL: 300,
			},
			wantErr: false,
		},
		{
			name: "session_ttl 59 returns error",
			cfg: FlareSolverrConfig{
				Enabled:    true,
				URL:        "http://localhost:8191/v1",
				Timeout:    30,
				MaxRetries: 3,
				SessionTTL: 59,
			},
			wantErr: true,
			errMsg:  "flaresolverr.session_ttl must be between 60 and 3600",
		},
		{
			name: "session_ttl 3601 returns error",
			cfg: FlareSolverrConfig{
				Enabled:    true,
				URL:        "http://localhost:8191/v1",
				Timeout:    30,
				MaxRetries: 3,
				SessionTTL: 3601,
			},
			wantErr: true,
			errMsg:  "flaresolverr.session_ttl must be between 60 and 3600",
		},
		{
			name: "session_ttl 60 is valid",
			cfg: FlareSolverrConfig{
				Enabled:    true,
				URL:        "http://localhost:8191/v1",
				Timeout:    30,
				MaxRetries: 3,
				SessionTTL: 60,
			},
			wantErr: false,
		},
		{
			name: "session_ttl 3600 is valid",
			cfg: FlareSolverrConfig{
				Enabled:    true,
				URL:        "http://localhost:8191/v1",
				Timeout:    30,
				MaxRetries: 3,
				SessionTTL: 3600,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFlareSolverrConfig("flaresolverr", tt.cfg)
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

func TestValidateRequestDelay(t *testing.T) {
	tests := []struct {
		name    string
		delay   int
		wantErr bool
		errMsg  string
	}{
		{
			name:    "negative returns error",
			delay:   -1,
			wantErr: true,
			errMsg:  "request_delay.request_delay must be non-negative",
		},
		{
			name:    "zero is valid",
			delay:   0,
			wantErr: false,
		},
		{
			name:    "500 is valid",
			delay:   500,
			wantErr: false,
		},
		{
			name:    "5000 is valid",
			delay:   5000,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRequestDelay("request_delay", tt.delay)
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

func TestValidateHTTPBaseURL(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty string is valid (optional)",
			raw:     "",
			wantErr: false,
		},
		{
			name:    "http URL is valid",
			raw:     "http://example.com",
			wantErr: false,
		},
		{
			name:    "https URL is valid",
			raw:     "https://example.com",
			wantErr: false,
		},
		{
			name:    "ftp URL returns error",
			raw:     "ftp://example.com",
			wantErr: true,
			errMsg:  "base_url must be a valid HTTP or HTTPS URL",
		},
		{
			name:    "invalid returns error",
			raw:     "invalid",
			wantErr: true,
			errMsg:  "base_url must be a valid HTTP or HTTPS URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateHTTPBaseURL("base_url", tt.raw)
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
