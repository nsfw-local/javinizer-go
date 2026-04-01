package config

import (
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"testing"
)

func TestDebugLoadFlow(t *testing.T) {
	repoRoot := filepath.Join("..", "..", "configs", "config.yaml")

	data, err := os.ReadFile(repoRoot)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}

	// Check what's in the YAML for scrapers section
	var node yaml.Node
	yaml.Unmarshal(data, &node)

	// Find scrapers node
	for i := 0; i < len(node.Content[0].Content); i += 2 {
		key := node.Content[0].Content[i].Value
		if key == "scrapers" {
			t.Logf("Found scrapers at index %d", i)
			scrapersNode := node.Content[0].Content[i+1]
			t.Logf("Scrapers node has %d content items", len(scrapersNode.Content))

			// Count keys
			keyCount := 0
			for j := 0; j < len(scrapersNode.Content); j += 2 {
				keyCount++
				if keyCount <= 10 {
					t.Logf("  scrapers key: %s", scrapersNode.Content[j].Value)
				}
			}
			t.Logf("  ... (total %d keys)", keyCount)
		}
	}

	// Now load and check
	cfg, err := Load(repoRoot)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}

	t.Logf("\nAfter Load:")
	t.Logf("  Overrides: %d keys", len(cfg.Scrapers.Overrides))
	t.Logf("  Proxy.Profiles: %+v", cfg.Scrapers.Proxy.Profiles)
}
