package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProxyConfig_Loading(t *testing.T) {
	yamlContent := `
scrapers:
  proxy:
    enabled: true
    url: "http://proxy.example.com:8080"
    username: "user"
    password: "pass"

output:
  download_proxy:
    enabled: true
    url: "socks5://localhost:1080"
`
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(cfgPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify scraper proxy
	if !cfg.Scrapers.Proxy.Enabled {
		t.Error("Expected scraper proxy to be enabled")
	}
	if cfg.Scrapers.Proxy.URL != "http://proxy.example.com:8080" {
		t.Errorf("Expected proxy URL 'http://proxy.example.com:8080', got '%s'", cfg.Scrapers.Proxy.URL)
	}
	if cfg.Scrapers.Proxy.Username != "user" {
		t.Errorf("Expected username 'user', got '%s'", cfg.Scrapers.Proxy.Username)
	}
	if cfg.Scrapers.Proxy.Password != "pass" {
		t.Errorf("Expected password 'pass', got '%s'", cfg.Scrapers.Proxy.Password)
	}

	// Verify download proxy
	if !cfg.Output.DownloadProxy.Enabled {
		t.Error("Expected download proxy to be enabled")
	}
	if cfg.Output.DownloadProxy.URL != "socks5://localhost:1080" {
		t.Errorf("Expected download proxy URL 'socks5://localhost:1080', got '%s'", cfg.Output.DownloadProxy.URL)
	}
}

func TestProxyConfig_Disabled(t *testing.T) {
	yamlContent := `
scrapers:
  proxy:
    enabled: false

output:
  download_proxy:
    enabled: false
`
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(cfgPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Scrapers.Proxy.Enabled {
		t.Error("Expected scraper proxy to be disabled")
	}
	if cfg.Output.DownloadProxy.Enabled {
		t.Error("Expected download proxy to be disabled")
	}
}

func TestProxyConfig_DefaultValues(t *testing.T) {
	cfg := DefaultConfig()

	// Verify default scraper proxy
	if cfg.Scrapers.Proxy.Enabled {
		t.Error("Expected scraper proxy to be disabled by default")
	}
	if cfg.Scrapers.Proxy.URL != "" {
		t.Error("Expected scraper proxy URL to be empty by default")
	}

	// Verify default download proxy
	if cfg.Output.DownloadProxy.Enabled {
		t.Error("Expected download proxy to be disabled by default")
	}
	if cfg.Output.DownloadProxy.URL != "" {
		t.Error("Expected download proxy URL to be empty by default")
	}
}

func TestProxyConfig_OnlyScraperProxy(t *testing.T) {
	yamlContent := `
scrapers:
  proxy:
    enabled: true
    url: "http://proxy.example.com:8080"
`
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(cfgPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify scraper proxy is set
	if !cfg.Scrapers.Proxy.Enabled {
		t.Error("Expected scraper proxy to be enabled")
	}

	// Verify download proxy remains at default (disabled)
	if cfg.Output.DownloadProxy.Enabled {
		t.Error("Expected download proxy to be disabled by default")
	}
}

func TestProxyConfig_WithAuthentication(t *testing.T) {
	yamlContent := `
scrapers:
  proxy:
    enabled: true
    url: "http://proxy.example.com:8080"
    username: "testuser"
    password: "testpass123"
`
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(cfgPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Scrapers.Proxy.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", cfg.Scrapers.Proxy.Username)
	}
	if cfg.Scrapers.Proxy.Password != "testpass123" {
		t.Errorf("Expected password 'testpass123', got '%s'", cfg.Scrapers.Proxy.Password)
	}
}

func TestProxyConfig_SOCKS5Proxy(t *testing.T) {
	yamlContent := `
scrapers:
  proxy:
    enabled: true
    url: "socks5://localhost:1080"
    username: "socksuser"
    password: "sockspass"
`
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(cfgPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if !cfg.Scrapers.Proxy.Enabled {
		t.Error("Expected scraper proxy to be enabled")
	}
	if cfg.Scrapers.Proxy.URL != "socks5://localhost:1080" {
		t.Errorf("Expected SOCKS5 URL 'socks5://localhost:1080', got '%s'", cfg.Scrapers.Proxy.URL)
	}
	if cfg.Scrapers.Proxy.Username != "socksuser" {
		t.Errorf("Expected username 'socksuser', got '%s'", cfg.Scrapers.Proxy.Username)
	}
}

func TestProxyConfig_EmptyURL(t *testing.T) {
	yamlContent := `
scrapers:
  proxy:
    enabled: true
    url: ""
`
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(cfgPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Should still load successfully
	if !cfg.Scrapers.Proxy.Enabled {
		t.Error("Expected proxy enabled flag to be true")
	}
	if cfg.Scrapers.Proxy.URL != "" {
		t.Errorf("Expected empty URL, got '%s'", cfg.Scrapers.Proxy.URL)
	}
}

func TestProxyConfig_SeparateProxies(t *testing.T) {
	yamlContent := `
scrapers:
  proxy:
    enabled: true
    url: "http://scraper-proxy.example.com:8080"

output:
  download_proxy:
    enabled: true
    url: "http://download-proxy.example.com:3128"
`
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(cfgPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify different proxies
	if cfg.Scrapers.Proxy.URL != "http://scraper-proxy.example.com:8080" {
		t.Errorf("Expected scraper proxy URL 'http://scraper-proxy.example.com:8080', got '%s'", cfg.Scrapers.Proxy.URL)
	}
	if cfg.Output.DownloadProxy.URL != "http://download-proxy.example.com:3128" {
		t.Errorf("Expected download proxy URL 'http://download-proxy.example.com:3128', got '%s'", cfg.Output.DownloadProxy.URL)
	}
}
