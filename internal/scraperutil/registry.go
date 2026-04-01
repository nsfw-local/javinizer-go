package scraperutil

// ValidatorFunc is a function that validates a scraper config.
// Uses any to break the import cycle between config and scraper packages.
// The validator function should type-assert the any to *config.ScraperSettings internally.
type ValidatorFunc func(any) error

// validatorRegistry holds validator functions for config validation delegation.
var validatorRegistry = make(map[string]ValidatorFunc)

// RegisterValidator registers a scraper config validator function.
// This is called from each scraper package's init() function.
func RegisterValidator(name string, fn ValidatorFunc) {
	validatorRegistry[name] = fn
}

// GetValidator returns the validator function for the named scraper.
func GetValidator(name string) ValidatorFunc {
	return validatorRegistry[name]
}

// ResetValidators clears the validator registry.
// Primarily used for test isolation.
func ResetValidators() {
	validatorRegistry = make(map[string]ValidatorFunc)
}

// ScraperConfigAccessor is a function that returns the scraper-specific
// *ScraperSettings from the shared ScrapersConfig.
// Uses any to break the import cycle between scraperutil and config packages.
// The accessor function is defined in the scraper package (which already imports config).
type ScraperConfigAccessor func(any) any

// scraperConfigRegistry holds config field accessors for plug-n-play scraper discovery.
var scraperConfigRegistry = make(map[string]ScraperConfigAccessor)

// RegisterScraperConfig registers a scraper config field accessor.
// Called from each scraper package's init() function.
// The accessor function handles the type assertion from any to *config.ScrapersConfig.
func RegisterScraperConfig(name string, accessor ScraperConfigAccessor) {
	scraperConfigRegistry[name] = accessor
}

// GetScraperConfigs returns a copy of all registered scraper config accessors.
func GetScraperConfigs() map[string]ScraperConfigAccessor {
	result := make(map[string]ScraperConfigAccessor, len(scraperConfigRegistry))
	for k, v := range scraperConfigRegistry {
		result[k] = v
	}
	return result
}

// ResetScraperConfigs clears the scraper config registry.
// Primarily used for test isolation.
func ResetScraperConfigs() {
	scraperConfigRegistry = make(map[string]ScraperConfigAccessor)
}

// ConfigFactory is a function that returns a new empty scraper config instance.
// The returned value is a pointer to a struct that implements yaml.Unmarshaler.
// Uses func() any to break the import cycle between scraperutil and config packages.
type ConfigFactory func() any

// configFactoryRegistry holds factory functions for creating scraper config instances.
var configFactoryRegistry = make(map[string]ConfigFactory)

// RegisterConfigFactory registers a scraper config factory function.
// Called from each scraper package's init() function.
// The factory returns a new empty instance of the scraper's config struct.
func RegisterConfigFactory(name string, factory ConfigFactory) {
	configFactoryRegistry[name] = factory
}

// GetConfigFactory returns the factory function for the named scraper.
func GetConfigFactory(name string) ConfigFactory {
	return configFactoryRegistry[name]
}

// ResetConfigFactories clears the config factory registry.
// Primarily used for test isolation.
func ResetConfigFactories() {
	configFactoryRegistry = make(map[string]ConfigFactory)
}

// FlattenFunc converts a scraper-specific flat config to unified *config.ScraperSettings.
// Uses any to break the import cycle between scraperutil and config packages.
// The function type-asserts the any to the scraper's concrete config type internally.
type FlattenFunc func(any) any

// flattenRegistry maps scraper name to flatten function.
var flattenRegistry = map[string]FlattenFunc{}

// RegisterFlattenFunc registers a flatten function for a scraper.
func RegisterFlattenFunc(name string, fn FlattenFunc) {
	flattenRegistry[name] = fn
}

// GetFlattenFunc returns the flatten function for a scraper.
func GetFlattenFunc(name string) FlattenFunc {
	return flattenRegistry[name]
}

// ResetFlattenFuncs clears the flatten registry.
// Primarily used for test isolation.
func ResetFlattenFuncs() {
	flattenRegistry = make(map[string]FlattenFunc)
}

// ScraperConfigInterface is implemented by scraper-specific config types
// (both from scraper packages and config package). This breaks the type
// dependency cycle and allows FlattenFunc to work across packages.
//
// Implement this interface on config package types by adding wrapper methods.
// Scraper package types already have compatible fields.
type ScraperConfigInterface interface {
	IsEnabled() bool
	GetUserAgent() string
	GetRequestDelay() int
	GetMaxRetries() int
	GetProxy() any         // Returns *config.ProxyConfig
	GetDownloadProxy() any // Returns *config.ProxyConfig
}

// ScraperConfigFlattenFunc converts a ScraperConfigInterface to *config.ScraperSettings.
// Uses any to break the import cycle since scraperutil cannot import config.
type ScraperConfigFlattenFunc func(ScraperConfigInterface) any

// GetPriorities returns scraper names sorted by priority (highest first).
// Derives from defaultScraperSettingsRegistry to maintain single source of truth.
// This registry is populated by scraper packages calling RegisterDefaultScraperDefaults,
// which chains from scraper.RegisterScraperDefaults in scraper/plugin.go.
func GetPriorities() []string {
	if len(defaultScraperSettingsRegistry) == 0 {
		return nil
	}

	type pair struct {
		name     string
		priority int
	}
	pairs := make([]pair, 0, len(defaultScraperSettingsRegistry))
	for name, def := range defaultScraperSettingsRegistry {
		pairs = append(pairs, pair{name: name, priority: def.priority})
	}

	// Simple bubble sort (list is small, typically 13 scrapers)
	for i := 0; i < len(pairs); i++ {
		for j := i + 1; j < len(pairs); j++ {
			if pairs[j].priority > pairs[i].priority {
				pairs[i], pairs[j] = pairs[j], pairs[i]
			}
		}
	}

	result := make([]string, len(pairs))
	for i, p := range pairs {
		result[i] = p.name
	}
	return result
}

// defaultScraperSettingsRegistry holds (settings any, priority int) for each scraper.
// settings is stored as any to avoid importing config package.
var defaultScraperSettingsRegistry = map[string]struct {
	settings any
	priority int
}{}

// RegisterDefaultScraperSettings registers a scraper's default settings and priority.
// Called from each scraper package's init() function.
// Settings is stored as any to avoid import cycle with config package.
func RegisterDefaultScraperSettings(name string, settings any, priority int) {
	defaultScraperSettingsRegistry[name] = struct {
		settings any
		priority int
	}{settings: settings, priority: priority}
}

// GetDefaultScraperSettings returns a copy of all registered scraper settings as map[string]any.
// Callers (config.go) type-assert the any to config.ScraperSettings.
func GetDefaultScraperSettings() map[string]any {
	result := make(map[string]any, len(defaultScraperSettingsRegistry))
	for k, v := range defaultScraperSettingsRegistry {
		result[k] = v.settings
	}
	return result
}

// ResetDefaults clears the default scraper settings registry.
// Primarily used for test isolation.
// Note: priorityRegistry removed - GetPriorities now derives from scraper.GetRegisteredDefaults()
func ResetDefaults() {
	defaultScraperSettingsRegistry = map[string]struct {
		settings any
		priority int
	}{}
}
