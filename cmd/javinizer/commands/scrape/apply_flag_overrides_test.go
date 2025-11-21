package scrape

import (
	"fmt"
	"testing"

	"github.com/javinizer/javinizer-go/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			cfg := &config.Config{
				Scrapers: config.ScrapersConfig{
					DMM: config.DMMConfig{
						ScrapeActress: !tt.expected, // Start with opposite value
					},
				},
			}

			// Set the flag
			err := cmd.Flags().Set(tt.flagName, fmt.Sprintf("%t", tt.flagVal))
			require.NoError(t, err)

			// Apply overrides
			ApplyFlagOverrides(cmd, cfg)

			assert.Equal(t, tt.expected, cfg.Scrapers.DMM.ScrapeActress,
				"Flag %s should set ScrapeActress to %t", tt.flagName, tt.expected)
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
		{"enable headless (deprecated)", "headless", true, true},
		{"disable no-headless (deprecated)", "no-headless", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewCommand()
			cfg := &config.Config{
				Scrapers: config.ScrapersConfig{
					DMM: config.DMMConfig{
						EnableBrowser: !tt.expected, // Opposite default
					},
				},
			}

			err := cmd.Flags().Set(tt.flagName, fmt.Sprintf("%t", tt.flagVal))
			require.NoError(t, err)

			ApplyFlagOverrides(cmd, cfg)

			assert.Equal(t, tt.expected, cfg.Scrapers.DMM.EnableBrowser,
				"Flag %s should set EnableBrowser to %t", tt.flagName, tt.expected)
		})
	}
}

// TestApplyFlagOverrides_BrowserTimeout tests browser-timeout flag overrides
func TestApplyFlagOverrides_BrowserTimeout(t *testing.T) {
	tests := []struct {
		name        string
		flagName    string
		flagVal     int
		defaultVal  int
		expectedVal int
		desc        string
	}{
		{"set browser-timeout to 45", "browser-timeout", 45, 30, 45, "Should override config"},
		{"set browser-timeout to 0 (no override)", "browser-timeout", 0, 30, 30, "Zero should not override"},
		{"set headless-timeout (deprecated)", "headless-timeout", 60, 30, 60, "Deprecated flag should still work"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewCommand()
			cfg := &config.Config{
				Scrapers: config.ScrapersConfig{
					DMM: config.DMMConfig{
						BrowserTimeout: tt.defaultVal,
					},
				},
			}

			err := cmd.Flags().Set(tt.flagName, fmt.Sprintf("%d", tt.flagVal))
			require.NoError(t, err)

			ApplyFlagOverrides(cmd, cfg)

			assert.Equal(t, tt.expectedVal, cfg.Scrapers.DMM.BrowserTimeout, tt.desc)
		})
	}
}

// TestApplyFlagOverrides_ActressDatabase tests actress-db flag overrides
func TestApplyFlagOverrides_ActressDatabase(t *testing.T) {
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
			cfg := &config.Config{
				Metadata: config.MetadataConfig{
					ActressDatabase: config.ActressDatabaseConfig{
						Enabled: !tt.expected,
					},
				},
			}

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
			cfg := &config.Config{
				Metadata: config.MetadataConfig{
					GenreReplacement: config.GenreReplacementConfig{
						Enabled: !tt.expected,
					},
				},
			}

			err := cmd.Flags().Set(tt.flagName, fmt.Sprintf("%t", tt.flagVal))
			require.NoError(t, err)

			ApplyFlagOverrides(cmd, cfg)

			assert.Equal(t, tt.expected, cfg.Metadata.GenreReplacement.Enabled,
				"Flag %s should set GenreReplacement.Enabled to %t", tt.flagName, tt.expected)
		})
	}
}

// TestApplyFlagOverrides_NoFlagsSet tests behavior when no flags are set (config unchanged)
func TestApplyFlagOverrides_NoFlagsSet(t *testing.T) {
	cmd := NewCommand()
	cfg := &config.Config{
		Scrapers: config.ScrapersConfig{
			DMM: config.DMMConfig{
				ScrapeActress:  true,
				EnableBrowser:  false,
				BrowserTimeout: 30,
			},
		},
		Metadata: config.MetadataConfig{
			ActressDatabase: config.ActressDatabaseConfig{
				Enabled: true,
			},
			GenreReplacement: config.GenreReplacementConfig{
				Enabled: false,
			},
		},
	}

	// Save original values
	originalScrapeActress := cfg.Scrapers.DMM.ScrapeActress
	originalEnableBrowser := cfg.Scrapers.DMM.EnableBrowser
	originalBrowserTimeout := cfg.Scrapers.DMM.BrowserTimeout
	originalActressDB := cfg.Metadata.ActressDatabase.Enabled
	originalGenreReplacement := cfg.Metadata.GenreReplacement.Enabled

	// Don't set any flags - just call ApplyFlagOverrides
	ApplyFlagOverrides(cmd, cfg)

	// Config should be unchanged
	assert.Equal(t, originalScrapeActress, cfg.Scrapers.DMM.ScrapeActress)
	assert.Equal(t, originalEnableBrowser, cfg.Scrapers.DMM.EnableBrowser)
	assert.Equal(t, originalBrowserTimeout, cfg.Scrapers.DMM.BrowserTimeout)
	assert.Equal(t, originalActressDB, cfg.Metadata.ActressDatabase.Enabled)
	assert.Equal(t, originalGenreReplacement, cfg.Metadata.GenreReplacement.Enabled)
}

// TestApplyFlagOverrides_MultipleFlags tests combining multiple flag overrides
func TestApplyFlagOverrides_MultipleFlags(t *testing.T) {
	cmd := NewCommand()
	cfg := &config.Config{
		Scrapers: config.ScrapersConfig{
			DMM: config.DMMConfig{
				ScrapeActress:  false,
				EnableBrowser:  false,
				BrowserTimeout: 30,
			},
		},
		Metadata: config.MetadataConfig{
			ActressDatabase: config.ActressDatabaseConfig{
				Enabled: false,
			},
			GenreReplacement: config.GenreReplacementConfig{
				Enabled: false,
			},
		},
	}

	// Set multiple flags
	require.NoError(t, cmd.Flags().Set("scrape-actress", "true"))
	require.NoError(t, cmd.Flags().Set("browser", "true"))
	require.NoError(t, cmd.Flags().Set("browser-timeout", "45"))
	require.NoError(t, cmd.Flags().Set("actress-db", "true"))
	require.NoError(t, cmd.Flags().Set("genre-replacement", "true"))

	ApplyFlagOverrides(cmd, cfg)

	// Verify all overrides applied
	assert.True(t, cfg.Scrapers.DMM.ScrapeActress, "ScrapeActress should be enabled")
	assert.True(t, cfg.Scrapers.DMM.EnableBrowser, "EnableBrowser should be enabled")
	assert.Equal(t, 45, cfg.Scrapers.DMM.BrowserTimeout, "BrowserTimeout should be 45")
	assert.True(t, cfg.Metadata.ActressDatabase.Enabled, "ActressDatabase should be enabled")
	assert.True(t, cfg.Metadata.GenreReplacement.Enabled, "GenreReplacement should be enabled")
}

// TestApplyFlagOverrides_BrowserPrecedence tests that new --browser flag takes precedence over deprecated --headless
func TestApplyFlagOverrides_BrowserPrecedence(t *testing.T) {
	tests := []struct {
		name         string
		browserFlag  bool
		headlessFlag bool
		expected     bool
		desc         string
	}{
		{"browser overrides headless (both true)", true, false, true, "New flag should win"},
		{"browser overrides headless (browser false, headless true)", false, true, false, "New flag should win"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewCommand()
			cfg := &config.Config{
				Scrapers: config.ScrapersConfig{
					DMM: config.DMMConfig{
						EnableBrowser: false,
					},
				},
			}

			// Set both flags (browser should take precedence)
			require.NoError(t, cmd.Flags().Set("browser", fmt.Sprintf("%t", tt.browserFlag)))
			require.NoError(t, cmd.Flags().Set("headless", fmt.Sprintf("%t", tt.headlessFlag)))

			ApplyFlagOverrides(cmd, cfg)

			assert.Equal(t, tt.expected, cfg.Scrapers.DMM.EnableBrowser, tt.desc)
		})
	}
}

// TestApplyFlagOverrides_BoolFalseFlagBehavior tests that setting a bool flag to false ONLY changes config when flag is changed
func TestApplyFlagOverrides_BoolFalseFlagBehavior(t *testing.T) {
	cmd := NewCommand()
	cfg := &config.Config{
		Scrapers: config.ScrapersConfig{
			DMM: config.DMMConfig{
				ScrapeActress: true, // Start as true
			},
		},
	}

	// Don't set scrape-actress flag at all
	// Just call ApplyFlagOverrides - config should remain true
	ApplyFlagOverrides(cmd, cfg)

	assert.True(t, cfg.Scrapers.DMM.ScrapeActress,
		"ScrapeActress should remain true when flag not set")

	// Now explicitly set scrape-actress to false (this won't work because flag only enables)
	// The actual way to disable is using no-scrape-actress
	cmd2 := NewCommand()
	cfg2 := &config.Config{
		Scrapers: config.ScrapersConfig{
			DMM: config.DMMConfig{
				ScrapeActress: true,
			},
		},
	}

	require.NoError(t, cmd2.Flags().Set("no-scrape-actress", "true"))
	ApplyFlagOverrides(cmd2, cfg2)

	assert.False(t, cfg2.Scrapers.DMM.ScrapeActress,
		"no-scrape-actress flag should disable ScrapeActress")
}
