package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestConfigYAMLLoadAndRoundTrip tests that configs/config.yaml can be loaded
// and round-tripped successfully. This ensures that the default generated config.yaml
// will work for users without config.yaml for the first time.
func TestConfigYAMLLoadAndRoundTrip(t *testing.T) {
	// Find the config.yaml relative to the repo root
	repoRoot := filepath.Join("..", "..", "configs", "config.yaml")

	// Load the config from repo root
	cfg, err := Load(repoRoot)
	if err != nil {
		t.Fatalf("Failed to load config.yaml: %v", err)
	}

	// Verify critical fields are loaded
	if cfg.Scrapers.Priority == nil || len(cfg.Scrapers.Priority) == 0 {
		t.Error("Scrapers.Priority should not be empty")
	}

	// Verify all scrapers are loaded
	expectedScrapers := []string{"r18dev", "dmm", "mgstage", "javlibrary", "javdb", "javbus", "jav321", "tokyohot", "aventertainment", "dlgetchu", "libredmm", "caribbeancom", "fc2"}
	for _, scraper := range expectedScrapers {
		if _, ok := cfg.Scrapers.Overrides[scraper]; !ok {
			t.Errorf("Scraper %q not found in Overrides", scraper)
		}
	}

	// Save to a temp file
	tmpDir := t.TempDir()
	tmpPath := filepath.Join(tmpDir, "config.yaml")
	err = Save(cfg, tmpPath)
	if err != nil {
		t.Fatalf("Failed to save config to temp file: %v", err)
	}

	// Load the saved config
	reloaded, err := Load(tmpPath)
	if err != nil {
		t.Fatalf("Failed to load re-saved config: %v", err)
	}

	// Verify key fields are preserved after round-trip
	if len(reloaded.Scrapers.Priority) != len(cfg.Scrapers.Priority) {
		t.Errorf("Priority length mismatch after round-trip: got %d, want %d",
			len(reloaded.Scrapers.Priority), len(cfg.Scrapers.Priority))
	}

	// Verify scrapers are still present after round-trip
	for _, scraper := range expectedScrapers {
		if _, ok := reloaded.Scrapers.Overrides[scraper]; !ok {
			t.Errorf("Scraper %q not found in Overrides after round-trip", scraper)
		}
	}
}

// TestConfigYAMLScraperFlareSolverr tests that all scrapers with flaresolverr
// in config.yaml can be loaded correctly.
func TestConfigYAMLScraperFlareSolverr(t *testing.T) {
	repoRoot := filepath.Join("..", "..", "configs", "config.yaml")

	cfg, err := Load(repoRoot)
	if err != nil {
		t.Fatalf("Failed to load config.yaml: %v", err)
	}

	// Scrapers that should have flaresolverr in config
	scrapersWithFlareSolverr := []string{"r18dev", "dmm", "mgstage", "javlibrary", "javdb", "javbus", "jav321", "tokyohot", "aventertainment", "dlgetchu", "libredmm", "caribbeancom", "fc2"}

	for _, scraper := range scrapersWithFlareSolverr {
		scraperCfg, ok := cfg.Scrapers.Overrides[scraper]
		if !ok {
			t.Errorf("Scraper %q not found in Overrides", scraper)
			continue
		}
		if scraperCfg == nil {
			t.Errorf("Scraper %q config is nil", scraper)
			continue
		}
		// UseFlareSolverr should be accessible
		t.Logf("%s: UseFlareSolverr=%v", scraper, scraperCfg.UseFlareSolverr)
	}
}

// TestGeneratedConfigLoadable tests that the generated config.yaml is loadable
// by simulating a fresh install (loading the file that would be generated).
func TestGeneratedConfigLoadable(t *testing.T) {
	// This test validates that the configs/config.yaml shipped with the project
	// is always loadable. If this test fails, users with a fresh install will
	// see errors like "field X not found in type Y".
	repoRoot := filepath.Join("..", "..", "configs", "config.yaml")

	// Check file exists
	if _, err := os.Stat(repoRoot); os.IsNotExist(err) {
		t.Skip("configs/config.yaml not found, skipping integration test")
	}

	cfg, err := Load(repoRoot)
	if err != nil {
		t.Fatalf("configs/config.yaml is not loadable: %v", err)
	}

	// Verify no scraper configs are nil (would indicate unmarshal issue)
	for name, scraperCfg := range cfg.Scrapers.Overrides {
		if scraperCfg == nil {
			t.Errorf("Scraper %q has nil config (possible unmarshal issue)", name)
		}
	}

	// Verify UseFlareSolverr is accessible for all scrapers
	for name, scraperCfg := range cfg.Scrapers.Overrides {
		if scraperCfg != nil {
			_ = scraperCfg.UseFlareSolverr // Just verify it's accessible
			t.Logf("%s: UseFlareSolverr=%v", name, scraperCfg.UseFlareSolverr)
		}
	}

	// Validate the config to catch issues like invalid proxy profiles
	// This is critical because Load() doesn't validate by default
	if err := cfg.Validate(); err != nil {
		t.Errorf("configs/config.yaml validation failed: %v", err)
	}
}
