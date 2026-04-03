package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// validateHTTPBaseURL validates that a URL has an http or https scheme and a host.
func validateHTTPBaseURL(path, raw string) error {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return fmt.Errorf("%s must be a valid http(s) URL", path)
	}
	if (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return fmt.Errorf("%s must be a valid http(s) URL", path)
	}
	return nil
}

func validateProxyProfileConfig(c *Config) error {
	if c == nil {
		return nil
	}

	// Ensure Overrides is populated before validation.
	// This must be called here (not just in Validate()) so that direct calls
	// to validateProxyProfileConfig pick up any flat config modifications.
	if c.Scrapers.Overrides == nil {
		c.Scrapers.NormalizeScraperConfigs()
	}

	profiles := c.Scrapers.Proxy.Profiles

	if err := validateNoLegacyProxyDirectFields("scrapers.proxy", &c.Scrapers.Proxy); err != nil {
		return err
	}
	if c.Scrapers.Proxy.Enabled && c.Scrapers.Proxy.DefaultProfile == "" {
		return fmt.Errorf("scrapers.proxy.default_profile is required when scrapers.proxy.enabled is true")
	}

	if c.Scrapers.Proxy.DefaultProfile != "" {
		if _, ok := profiles[c.Scrapers.Proxy.DefaultProfile]; !ok {
			return fmt.Errorf("scrapers.proxy.default_profile references unknown profile %q", c.Scrapers.Proxy.DefaultProfile)
		}
	}

	// CONF-04: Generic scraper proxy profile validation — iterates Overrides map.
	// NO hardcoded scraper-name branches.
	if err := c.validateScraperProxyProfiles(); err != nil {
		return err
	}

	// Validate output.download_proxy (not a scraper, special case)
	if err := validateProxyProfileRef("output.download_proxy", &c.Output.DownloadProxy, profiles); err != nil {
		return err
	}

	return nil
}

// validateScraperProxyProfiles validates proxy profiles for all scrapers generically.
// Uses c.Scrapers.Overrides map — NO hardcoded scraper-name branches.
// CONF-04: Adding a new scraper only requires adding it to flats map in normalizeScraperConfigs().
func (c *Config) validateScraperProxyProfiles() error {
	// Always re-normalize to pick up any modifications made to flat configs
	// after the previous NormalizeScraperConfigs() call (e.g., in tests).
	c.Scrapers.NormalizeScraperConfigs()

	for name, sc := range c.Scrapers.Overrides {
		path := "scrapers." + name

		if sc.Proxy != nil {
			if err := validateProxyProfileRef(path+".proxy", sc.Proxy, c.Scrapers.Proxy.Profiles); err != nil {
				return err
			}
		}

		if sc.DownloadProxy != nil {
			if err := validateProxyProfileRef(path+".download_proxy", sc.DownloadProxy, c.Scrapers.Proxy.Profiles); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateProxyProfileRef(path string, proxyCfg *ProxyConfig, profiles map[string]ProxyProfile) error {
	if proxyCfg == nil {
		return nil
	}

	if err := validateNoLegacyProxyDirectFields(path, proxyCfg); err != nil {
		return err
	}

	// enabled=true with empty profile means "inherit" mode - valid, no profile needed
	// enabled=true with non-empty profile means "specific" mode - profile is required
	if proxyCfg.Enabled && proxyCfg.Profile != "" {
		if _, ok := profiles[proxyCfg.Profile]; !ok {
			return fmt.Errorf("%s.profile references unknown profile %q", path, proxyCfg.Profile)
		}
	}
	return nil
}

// rejectUnknownProxyFields checks if the raw YAML node contains legacy proxy fields
// that are no longer supported. Returns an error if any are found.
// nolint:unused // Will be integrated into UnmarshalYAML in Task 2
func rejectUnknownProxyFields(node *yaml.Node, context string) error {
	if node == nil {
		return nil
	}

	if node.Kind != yaml.MappingNode {
		return nil
	}

	legacyFields := []string{"url", "username", "password", "use_main_proxy"}

	for i := 0; i < len(node.Content); i += 2 {
		if i < len(node.Content) {
			keyNode := node.Content[i]
			key := strings.ToLower(keyNode.Value)

			for _, legacy := range legacyFields {
				if key == legacy {
					return fmt.Errorf(
						"%s: field '%s' is no longer supported. "+
							"Use 'profile: <name>' to reference a proxy profile from scrapers.proxy.profiles instead",
						context, keyNode.Value,
					)
				}
			}
		}
	}

	return nil
}

func validateNoLegacyProxyDirectFields(path string, proxyCfg *ProxyConfig) error {
	if proxyCfg == nil {
		return nil
	}
	// Legacy field validation is now handled at the YAML parsing level
	// via rejectUnknownProxyFields() for url/username/password/use_main_proxy
	return nil
}

// validateFlareSolverrConfig validates FlareSolverr configuration
func validateFlareSolverrConfig(path string, cfg FlareSolverrConfig) error {
	if !cfg.Enabled {
		return nil
	}
	if cfg.URL == "" {
		return fmt.Errorf("%s.url is required when flaresolverr is enabled", path)
	}
	if cfg.Timeout < 1 || cfg.Timeout > 300 {
		return fmt.Errorf("%s.timeout must be between 1 and 300", path)
	}
	if cfg.MaxRetries < 0 || cfg.MaxRetries > 10 {
		return fmt.Errorf("%s.max_retries must be between 0 and 10", path)
	}
	if cfg.SessionTTL < 60 || cfg.SessionTTL > 3600 {
		return fmt.Errorf("%s.session_ttl must be between 60 and 3600", path)
	}
	return nil
}

// validateBrowserConfig validates Browser configuration
func validateBrowserConfig(path string, cfg BrowserConfig) error {
	if !cfg.Enabled {
		return nil // Disabled is valid
	}

	if cfg.Timeout < 1 || cfg.Timeout > 300 {
		return fmt.Errorf("%s.timeout must be between 1 and 300 seconds", path)
	}

	if cfg.MaxRetries < 0 || cfg.MaxRetries > 10 {
		return fmt.Errorf("%s.max_retries must be between 0 and 10", path)
	}

	if cfg.WindowWidth < 640 || cfg.WindowWidth > 3840 {
		return fmt.Errorf("%s.window_width must be between 640 and 3840", path)
	}

	if cfg.WindowHeight < 480 || cfg.WindowHeight > 2160 {
		return fmt.Errorf("%s.window_height must be between 480 and 2160", path)
	}

	if cfg.SlowMo < 0 || cfg.SlowMo > 5000 {
		return fmt.Errorf("%s.slow_mo must be between 0 and 5000", path)
	}

	// If binary_path is set, validate it exists
	if cfg.BinaryPath != "" {
		if _, err := os.Stat(cfg.BinaryPath); err != nil {
			return fmt.Errorf("%s.binary_path does not exist: %s", path, cfg.BinaryPath)
		}
	}

	return nil
}
