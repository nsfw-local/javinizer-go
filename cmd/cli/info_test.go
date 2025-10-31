package main

import (
	"path/filepath"
	"testing"

	"github.com/javinizer/javinizer-go/internal/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// TestRunInfo_DisplaysConfiguration verifies that runInfo displays config information
func TestRunInfo_DisplaysConfiguration(t *testing.T) {
	configPath, testCfg := createTestConfig(t,
		WithScraperPriority([]string{"r18dev", "dmm"}),
		WithOutputFolder("<ID> - <TITLE>"),
		WithOutputFile("<ID>"),
		WithDownloadCover(true),
	)

	withTempConfigFile(t, configPath, func() {
		// Load config into global variable
		err := loadConfig()
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		cmd := &cobra.Command{}

		stdout, _ := captureOutput(t, func() {
			err := runInfo(cmd, []string{}, testCfg)
			if err != nil {
				t.Fatalf("runInfo failed: %v", err)
			}
		})

		// Verify header
		assert.Contains(t, stdout, "=== Javinizer Configuration ===")

		// Verify config file path is shown
		assert.Contains(t, stdout, "Config file:")
		assert.Contains(t, stdout, configPath)

		// Verify database info
		assert.Contains(t, stdout, "Database:")
		assert.Contains(t, stdout, testCfg.Database.DSN)
		assert.Contains(t, stdout, testCfg.Database.Type)

		// Verify server info
		assert.Contains(t, stdout, "Server:")
		assert.Contains(t, stdout, testCfg.Server.Host)

		// Verify scrapers section
		assert.Contains(t, stdout, "Scrapers:")
		assert.Contains(t, stdout, "Priority:")
		assert.Contains(t, stdout, "r18dev")
		assert.Contains(t, stdout, "dmm")

		// Verify scraper status
		assert.Contains(t, stdout, "R18.dev:")
		assert.Contains(t, stdout, "DMM:")

		// Verify output settings
		assert.Contains(t, stdout, "Output:")
		assert.Contains(t, stdout, "Folder format:")
		assert.Contains(t, stdout, "<ID> - <TITLE>")
		assert.Contains(t, stdout, "File format:")
		assert.Contains(t, stdout, "<ID>")
		assert.Contains(t, stdout, "Download cover:")
		assert.Contains(t, stdout, "true")
	})
}

// TestRunInfo_ShowsScraperPriority verifies scraper priority is displayed correctly
func TestRunInfo_ShowsScraperPriority(t *testing.T) {
	tests := []struct {
		name     string
		priority []string
		contains []string
	}{
		{
			name:     "r18dev first",
			priority: []string{"r18dev", "dmm"},
			contains: []string{"r18dev", "dmm"},
		},
		{
			name:     "dmm only",
			priority: []string{"dmm"},
			contains: []string{"dmm"},
		},
		{
			name:     "custom priority order",
			priority: []string{"dmm", "r18dev"},
			contains: []string{"dmm", "r18dev"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath, testCfg := createTestConfig(t,
				WithScraperPriority(tt.priority),
			)

			withTempConfigFile(t, configPath, func() {
				err := loadConfig()
				if err != nil {
					t.Fatalf("Failed to load config: %v", err)
				}

				cmd := &cobra.Command{}

				stdout, _ := captureOutput(t, func() {
					err := runInfo(cmd, []string{}, testCfg)
					if err != nil {
						t.Fatalf("runInfo failed: %v", err)
					}
				})

				// Verify all expected scrapers are shown
				for _, scraper := range tt.contains {
					assert.Contains(t, stdout, scraper,
						"Expected scraper %s to be shown in priority", scraper)
				}
			})
		})
	}
}

// TestRunInfo_ShowsOutputConfiguration verifies output settings are displayed
func TestRunInfo_ShowsOutputConfiguration(t *testing.T) {
	tests := []struct {
		name           string
		folderFormat   string
		fileFormat     string
		downloadCover  bool
		downloadExtras bool
	}{
		{
			name:           "basic template",
			folderFormat:   "<ID>",
			fileFormat:     "<ID>",
			downloadCover:  true,
			downloadExtras: false,
		},
		{
			name:           "complex template",
			folderFormat:   "<ID> [<STUDIO>] - <TITLE> (<YEAR>)",
			fileFormat:     "<ID> - <TITLE>",
			downloadCover:  true,
			downloadExtras: true,
		},
		{
			name:           "minimal downloads",
			folderFormat:   "<ID>",
			fileFormat:     "<ID>",
			downloadCover:  false,
			downloadExtras: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath, testCfg := createTestConfig(t,
				WithOutputFolder(tt.folderFormat),
				WithOutputFile(tt.fileFormat),
				WithDownloadCover(tt.downloadCover),
				// Note: WithDownloadExtrafanart not in helpers, testing with default
			)

			withTempConfigFile(t, configPath, func() {
				err := loadConfig()
				if err != nil {
					t.Fatalf("Failed to load config: %v", err)
				}

				cmd := &cobra.Command{}

				stdout, _ := captureOutput(t, func() {
					err := runInfo(cmd, []string{}, testCfg)
					if err != nil {
						t.Fatalf("runInfo failed: %v", err)
					}
				})

				// Verify folder format
				assert.Contains(t, stdout, "Folder format:")
				assert.Contains(t, stdout, tt.folderFormat)

				// Verify file format
				assert.Contains(t, stdout, "File format:")
				assert.Contains(t, stdout, tt.fileFormat)

				// Verify download settings
				assert.Contains(t, stdout, "Download cover:")
				if tt.downloadCover {
					assert.Contains(t, stdout, "true")
				} else {
					assert.Contains(t, stdout, "false")
				}
			})
		})
	}
}

// TestRunInfo_ShowsDatabasePath verifies database configuration is displayed
func TestRunInfo_ShowsDatabasePath(t *testing.T) {
	tmpDir := t.TempDir()
	customDBPath := filepath.Join(tmpDir, "custom_database.db")

	configPath, testCfg := createTestConfig(t,
		WithDatabaseDSN(customDBPath),
	)

	withTempConfigFile(t, configPath, func() {
		err := loadConfig()
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		cmd := &cobra.Command{}

		stdout, _ := captureOutput(t, func() {
			err := runInfo(cmd, []string{}, testCfg)
			if err != nil {
				t.Fatalf("runInfo failed: %v", err)
			}
		})

		// Verify custom database path is shown
		assert.Contains(t, stdout, "Database:")
		assert.Contains(t, stdout, customDBPath)
		assert.Contains(t, stdout, "sqlite")
	})
}

// TestRunInfo_WithDefaultConfig verifies info works with default config
func TestRunInfo_WithDefaultConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create config with all defaults
	testCfg := config.DefaultConfig()
	testCfg.Database.DSN = filepath.Join(tmpDir, "test.db")
	err := config.Save(testCfg, configPath)
	if err != nil {
		t.Fatalf("Failed to save default config: %v", err)
	}

	withTempConfigFile(t, configPath, func() {
		err := loadConfig()
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		cmd := &cobra.Command{}

		stdout, _ := captureOutput(t, func() {
			err := runInfo(cmd, []string{}, testCfg)
			if err != nil {
				t.Fatalf("runInfo failed: %v", err)
			}
		})

		// Verify basic sections are present even with defaults
		assert.Contains(t, stdout, "Javinizer Configuration")
		assert.Contains(t, stdout, "Config file:")
		assert.Contains(t, stdout, "Database:")
		assert.Contains(t, stdout, "Server:")
		assert.Contains(t, stdout, "Scrapers:")
		assert.Contains(t, stdout, "Output:")

		// Verify no errors or panics occurred
		assert.NotContains(t, stdout, "error")
		assert.NotContains(t, stdout, "panic")
	})
}
