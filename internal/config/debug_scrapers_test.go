package config

import (
	"path/filepath"
	"testing"
)

func TestDebugScrapersInOverrides(t *testing.T) {
	repoRoot := filepath.Join("..", "..", "configs", "config.yaml")

	cfg, err := Load(repoRoot)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}

	t.Logf("Overrides: %d keys", len(cfg.Scrapers.Overrides))
	for k := range cfg.Scrapers.Overrides {
		t.Logf("  - %s", k)
	}

	t.Logf("Proxy.Profiles: %+v", cfg.Scrapers.Proxy.Profiles)

	// Check javlibrary
	if javCfg, ok := cfg.Scrapers.Overrides["javlibrary"]; ok {
		t.Logf("javlibrary.Proxy: %+v", javCfg.Proxy)
	} else {
		t.Logf("javlibrary not in Overrides!")
	}

	if err := cfg.Validate(); err != nil {
		t.Logf("Validate error: %v", err)
	}
}
