package config

// BrowserConfig holds global browser automation settings.
// Per-scraper overrides are configured via ScrapersConfig.Overrides with use_browser field.
type BrowserConfig struct {
	Enabled      bool   `yaml:"enabled" json:"enabled"`             // Master kill-switch (default: false)
	BinaryPath   string `yaml:"binary_path" json:"binary_path"`     // Chrome/Chromium path (auto-discovered if empty)
	Timeout      int    `yaml:"timeout" json:"timeout"`             // Operation timeout in seconds (default: 30)
	MaxRetries   int    `yaml:"max_retries" json:"max_retries"`     // Retry attempts (default: 3)
	Headless     bool   `yaml:"headless" json:"headless"`           // Run headless (default: true)
	StealthMode  bool   `yaml:"stealth_mode" json:"stealth_mode"`   // Anti-detection measures (default: true)
	WindowWidth  int    `yaml:"window_width" json:"window_width"`   // Viewport width (default: 1920)
	WindowHeight int    `yaml:"window_height" json:"window_height"` // Viewport height (default: 1080)
	SlowMo       int    `yaml:"slow_mo" json:"slow_mo"`             // Slow motion delay ms (default: 0)
	BlockImages  bool   `yaml:"block_images" json:"block_images"`   // Block images for speed (default: true)
	BlockCSS     bool   `yaml:"block_css" json:"block_css"`         // Block CSS for speed (default: false)
	UserAgent    string `yaml:"user_agent" json:"user_agent"`       // Override UA (empty = use scraper's)
	DebugVisible bool   `yaml:"debug_visible" json:"debug_visible"` // Show browser window for debugging (default: false)
}
