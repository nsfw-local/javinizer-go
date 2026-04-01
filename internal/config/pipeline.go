package config

import (
	"fmt"
	"strings"
)

type compatibilityRule struct {
	legacyMaxVersion int
	apply            func(cfg *Config, defaults *Config) (bool, error)
}

// appendMissingStrings keeps user ordering and appends default values that are
// missing from the existing list.
func appendMissingStrings(existing, defaults []string) ([]string, bool) {
	if len(existing) == 0 {
		return append([]string{}, defaults...), len(defaults) > 0
	}

	seen := make(map[string]bool, len(existing))
	for _, value := range existing {
		seen[value] = true
	}

	merged := append([]string{}, existing...)
	changed := false
	for _, value := range defaults {
		if seen[value] {
			continue
		}
		merged = append(merged, value)
		changed = true
	}
	return merged, changed
}

// legacyScraperPriorityBaseline returns the baseline list of all known scrapers.
// This is used for backward compatibility when older configs don't specify all scrapers.
func legacyScraperPriorityBaseline() []string {
	// Hardcoded list of all 13 known scrapers in priority order.
	// This replaces the old reflection-based approach which no longer works
	// since scraper configs are now stored in a map rather than struct fields.
	return []string{
		"r18dev",
		"dmm",
		"libredmm",
		"mgstage",
		"javlibrary",
		"javdb",
		"javbus",
		"jav321",
		"tokyohot",
		"aventertainment",
		"dlgetchu",
		"caribbeancom",
		"fc2",
	}
}

// configCompatibilityRules are applied idempotently for legacy configs.
// These should be exceptional; additive fields should rely on DefaultConfig().
var configCompatibilityRules = []compatibilityRule{
	{
		legacyMaxVersion: 2,
		apply: func(cfg *Config, _ *Config) (bool, error) {
			baseline := legacyScraperPriorityBaseline()
			merged, changed := appendMissingStrings(cfg.Scrapers.Priority, baseline)
			if changed {
				cfg.Scrapers.Priority = merged
			}
			return changed, nil
		},
	},
	{
		legacyMaxVersion: 2,
		apply: func(cfg *Config, defaults *Config) (bool, error) {
			// Preserve explicit false for update_enabled; only backfill interval.
			if cfg.System.UpdateCheckIntervalHours == 0 {
				cfg.System.UpdateCheckIntervalHours = defaults.System.UpdateCheckIntervalHours
				return true, nil
			}
			return false, nil
		},
	},
}

// applyCompatibilityRules upgrades legacy config behavior to current semantics.
// Returns true when any compatibility change is applied.
func applyCompatibilityRules(cfg *Config) (bool, error) {
	if cfg == nil {
		return false, nil
	}
	if cfg.ConfigVersion > CurrentConfigVersion {
		return false, fmt.Errorf(
			"config version %d is newer than supported version %d; please update Javinizer",
			cfg.ConfigVersion,
			CurrentConfigVersion,
		)
	}

	defaults := DefaultConfig()
	originalVersion := cfg.ConfigVersion
	changed := false

	for _, rule := range configCompatibilityRules {
		if originalVersion > rule.legacyMaxVersion {
			continue
		}
		ruleChanged, err := rule.apply(cfg, defaults)
		if err != nil {
			return false, err
		}
		changed = changed || ruleChanged
	}

	if cfg.ConfigVersion != CurrentConfigVersion {
		cfg.ConfigVersion = CurrentConfigVersion
		changed = true
	}

	return changed, nil
}

func normalizeField(value *string, defaultValue string, toLower bool) bool {
	if value == nil {
		return false
	}
	normalized := strings.TrimSpace(*value)
	if normalized == "" {
		normalized = defaultValue
	}
	if toLower {
		normalized = strings.ToLower(normalized)
	}
	if *value == normalized {
		return false
	}
	*value = normalized
	return true
}

func normalizeTranslationConfig(t *TranslationConfig) bool {
	if t == nil {
		return false
	}

	changed := false
	changed = normalizeField(&t.Provider, "openai", true) || changed
	changed = normalizeField(&t.SourceLanguage, "auto", false) || changed
	changed = normalizeField(&t.TargetLanguage, "ja", false) || changed
	changed = normalizeField(&t.OpenAI.BaseURL, "https://api.openai.com/v1", false) || changed
	changed = normalizeField(&t.OpenAI.Model, "gpt-4o-mini", false) || changed
	changed = normalizeField(&t.DeepL.Mode, "free", true) || changed
	changed = normalizeField(&t.Google.Mode, "free", true) || changed

	if t.TimeoutSeconds <= 0 {
		t.TimeoutSeconds = 60
		changed = true
	}

	return changed
}

// Normalize applies idempotent value normalization to config data.
func Normalize(cfg *Config) bool {
	if cfg == nil {
		return false
	}

	// Ensure Overrides is populated before accessing it.
	// This handles the case where a config was loaded via JSON (which doesn't
	// call NormalizeScraperConfigs like Load() does for YAML).
	cfg.Scrapers.NormalizeScraperConfigs()

	changed := false
	changed = normalizeField(&cfg.Database.Type, "sqlite", true) || changed
	if r18dev := cfg.Scrapers.Overrides["r18dev"]; r18dev != nil {
		changed = normalizeField(&r18dev.Language, "en", true) || changed
	}
	if javlib := cfg.Scrapers.Overrides["javlibrary"]; javlib != nil {
		changed = normalizeField(&javlib.Language, "en", true) || changed
	}

	if strings.TrimSpace(cfg.Scrapers.Referer) == "" {
		cfg.Scrapers.Referer = "https://www.dmm.co.jp/"
		changed = true
	}

	changed = normalizeTranslationConfig(&cfg.Metadata.Translation) || changed
	return changed
}

// Prepare runs compatibility migrations, normalization, and strict validation.
// Returns true when config data was changed during preparation.
func Prepare(cfg *Config) (bool, error) {
	if cfg == nil {
		return false, nil
	}

	compatChanged, err := applyCompatibilityRules(cfg)
	if err != nil {
		return false, err
	}

	normalized := Normalize(cfg)

	if err := cfg.Validate(); err != nil {
		return compatChanged || normalized, fmt.Errorf("invalid configuration: %w", err)
	}

	return compatChanged || normalized, nil
}
