package scrape

import (
	"fmt"
	"testing"

	"github.com/javinizer/javinizer-go/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	// Register scraper defaults for NormalizeScraperConfigs
	_ "github.com/javinizer/javinizer-go/internal/scraper/dmm"
)

// TestApplyFlagOverrides_ScrapeActress tests scrape-actress flag overrides
func TestApplyFlagOverrides_ScrapeActress(t *testing.T) {
	tests := []struct {
		name     string
		flagName string
		flagVal  bool
		expected bool
	}{
		{"enable scrape-actress", "scrape-actress", true, true},
		{"disable no-scrape-actress", "no-scrape-actress", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewCommand()
			cfg := config.DefaultConfig()
			// Normalize to populate Overrides map
			cfg.Scrapers.NormalizeScraperConfigs()
			// Initialize Extra map if nil and set initial value (opposite of expected)
			if cfg.Scrapers.Overrides["dmm"].Extra == nil {
				cfg.Scrapers.Overrides["dmm"].Extra = make(map[string]any)
			}
			cfg.Scrapers.Overrides["dmm"].Extra["scrape_actress"] = !tt.expected

			// Set the flag
			err := cmd.Flags().Set(tt.flagName, fmt.Sprintf("%t", tt.flagVal))
			require.NoError(t, err)

			// Apply overrides
			ApplyFlagOverrides(cmd, cfg)

			assert.Equal(t, tt.expected, cfg.Scrapers.Overrides["dmm"].GetBoolExtra("scrape_actress", false),
				"Flag %s should set scrape_actress to %t", tt.flagName, tt.expected)
		})
	}
}

// TestApplyFlagOverrides_BrowserMode tests browser flag overrides and deprecated backward compatibility
func TestApplyFlagOverrides_BrowserMode(t *testing.T) {
	tests := []struct {
		name     string
		flagName string
		flagVal  bool
		expected bool
	}{
		{"enable browser", "browser", true, true},
		{"disable no-browser", "no-browser", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewCommand()
			cfg := config.DefaultConfig()
			// Normalize to populate Overrides map
			cfg.Scrapers.NormalizeScraperConfigs()
			// Initialize Extra map if nil and set initial value (opposite of expected)
			if cfg.Scrapers.Overrides["dmm"].Extra == nil {
				cfg.Scrapers.Overrides["dmm"].Extra = make(map[string]any)
			}
			cfg.Scrapers.Overrides["dmm"].Extra["enable_browser"] = !tt.expected

			err := cmd.Flags().Set(tt.flagName, fmt.Sprintf("%t", tt.flagVal))
			require.NoError(t, err)

			ApplyFlagOverrides(cmd, cfg)

			assert.Equal(t, tt.expected, cfg.Scrapers.Overrides["dmm"].GetBoolExtra("enable_browser", false),
				"Flag %s should set enable_browser to %t", tt.flagName, tt.expected)
		})
	}
}

// TestApplyFlagOverrides_BrowserTimeout tests browser-timeout flag override
func TestApplyFlagOverrides_BrowserTimeout(t *testing.T) {
	cmd := NewCommand()
	cfg := config.DefaultConfig()

	// Normalize to populate Overrides map
	cfg.Scrapers.NormalizeScraperConfigs()
	// Initialize Extra map if nil and set initial timeout
	if cfg.Scrapers.Overrides["dmm"].Extra == nil {
		cfg.Scrapers.Overrides["dmm"].Extra = make(map[string]any)
	}
	cfg.Scrapers.Overrides["dmm"].Extra["browser_timeout"] = 30

	// Set the flag
	err := cmd.Flags().Set("browser-timeout", "60")
	require.NoError(t, err)

	ApplyFlagOverrides(cmd, cfg)

	assert.Equal(t, 60, cfg.Scrapers.Overrides["dmm"].GetIntExtra("browser_timeout", 0),
		"browser-timeout flag should override to 60")
}

// TestApplyFlagOverrides_ActressDB tests actress-db flag overrides
func TestApplyFlagOverrides_ActressDB(t *testing.T) {
	tests := []struct {
		name     string
		flagName string
		flagVal  bool
		expected bool
	}{
		{"enable actress-db", "actress-db", true, true},
		{"disable no-actress-db", "no-actress-db", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewCommand()
			cfg := config.DefaultConfig()
			cfg.Metadata.ActressDatabase.Enabled = !tt.expected

			err := cmd.Flags().Set(tt.flagName, fmt.Sprintf("%t", tt.flagVal))
			require.NoError(t, err)

			ApplyFlagOverrides(cmd, cfg)

			assert.Equal(t, tt.expected, cfg.Metadata.ActressDatabase.Enabled,
				"Flag %s should set ActressDatabase.Enabled to %t", tt.flagName, tt.expected)
		})
	}
}

// TestApplyFlagOverrides_GenreReplacement tests genre-replacement flag overrides
func TestApplyFlagOverrides_GenreReplacement(t *testing.T) {
	tests := []struct {
		name     string
		flagName string
		flagVal  bool
		expected bool
	}{
		{"enable genre-replacement", "genre-replacement", true, true},
		{"disable no-genre-replacement", "no-genre-replacement", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewCommand()
			cfg := config.DefaultConfig()
			cfg.Metadata.GenreReplacement.Enabled = !tt.expected

			err := cmd.Flags().Set(tt.flagName, fmt.Sprintf("%t", tt.flagVal))
			require.NoError(t, err)

			ApplyFlagOverrides(cmd, cfg)

			assert.Equal(t, tt.expected, cfg.Metadata.GenreReplacement.Enabled,
				"Flag %s should set GenreReplacement.Enabled to %t", tt.flagName, tt.expected)
		})
	}
}
