package config_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	config "github.com/javinizer/javinizer-go/internal/config"
	_ "github.com/javinizer/javinizer-go/internal/scraper/dmm"
	"github.com/javinizer/javinizer-go/internal/scraperutil"
)

func TestLoadFailsOnInvalidScrapersTimeoutType(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")
	content := `
scrapers:
  timeout_seconds: "bad"
`

	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	_, err := config.Load(cfgPath)
	if err == nil {
		t.Fatal("expected load to fail for invalid timeout_seconds type")
	}
	if !strings.Contains(err.Error(), "timeout_seconds must be an integer") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoadFailsOnUnknownScraperName(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")
	content := `
scrapers:
  customscraper:
    enabled: true
    custom_flag: true
`

	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	_, err := config.Load(cfgPath)
	if err == nil {
		t.Fatal("expected load to fail for unknown scraper name")
	}
	if !strings.Contains(err.Error(), `unknown scraper "customscraper"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

type untaggedPluginConfig struct {
	Enabled    bool
	CustomFlag bool
}

func TestLoadAllowsKnownScraperWithUntaggedExportedFields(t *testing.T) {
	scraperName := fmt.Sprintf("untagged_%d", time.Now().UnixNano())
	scraperutil.RegisterConfigFactory(scraperName, func() any {
		return &untaggedPluginConfig{}
	})

	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")
	content := fmt.Sprintf(`
scrapers:
  %s:
    enabled: true
    customflag: true
`, scraperName)

	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("expected load to succeed, got error: %v", err)
	}

	got, ok := cfg.Scrapers.Overrides[scraperName]
	if !ok || got == nil {
		t.Fatalf("expected scraper %q in overrides", scraperName)
	}
	if got.Extra == nil {
		t.Fatalf("expected extra map for scraper %q", scraperName)
	}
	if v, ok := got.Extra["customflag"].(bool); !ok || !v {
		t.Fatalf("expected customflag=true in extra, got: %#v", got.Extra["customflag"])
	}
}
