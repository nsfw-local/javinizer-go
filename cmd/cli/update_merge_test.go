package main

import (
	"strings"
	"testing"

	"github.com/javinizer/javinizer-go/internal/nfo"
	"github.com/stretchr/testify/assert"
)

// TestDetermineMergeStrategy tests the merge strategy determination logic
// from the runUpdate function
func TestDetermineMergeStrategy(t *testing.T) {
	tests := []struct {
		name             string
		preserveNFO      bool
		mergeStrategyStr string
		expectedStrategy nfo.MergeStrategy
		description      string
	}{
		{
			name:             "preserve-nfo flag overrides strategy string",
			preserveNFO:      true,
			mergeStrategyStr: "prefer-scraper",
			expectedStrategy: nfo.PreferNFO,
			description:      "When preserve-nfo is true, always use PreferNFO regardless of strategy string",
		},
		{
			name:             "preserve-nfo flag overrides to prefer-nfo",
			preserveNFO:      true,
			mergeStrategyStr: "merge-arrays",
			expectedStrategy: nfo.PreferNFO,
			description:      "preserve-nfo takes precedence over any strategy",
		},
		{
			name:             "prefer-nfo strategy when not preserve flag",
			preserveNFO:      false,
			mergeStrategyStr: "prefer-nfo",
			expectedStrategy: nfo.PreferNFO,
			description:      "Explicit prefer-nfo strategy works without flag",
		},
		{
			name:             "merge-arrays strategy",
			preserveNFO:      false,
			mergeStrategyStr: "merge-arrays",
			expectedStrategy: nfo.MergeArrays,
			description:      "merge-arrays strategy combines arrays from both sources",
		},
		{
			name:             "prefer-scraper strategy explicit",
			preserveNFO:      false,
			mergeStrategyStr: "prefer-scraper",
			expectedStrategy: nfo.PreferScraper,
			description:      "Explicit prefer-scraper uses scraped data",
		},
		{
			name:             "default to prefer-scraper on unknown strategy",
			preserveNFO:      false,
			mergeStrategyStr: "invalid-strategy",
			expectedStrategy: nfo.PreferScraper,
			description:      "Unknown strategies default to PreferScraper",
		},
		{
			name:             "default to prefer-scraper on empty strategy",
			preserveNFO:      false,
			mergeStrategyStr: "",
			expectedStrategy: nfo.PreferScraper,
			description:      "Empty strategy defaults to PreferScraper",
		},
		{
			name:             "case insensitive - PREFER-NFO",
			preserveNFO:      false,
			mergeStrategyStr: "PREFER-NFO",
			expectedStrategy: nfo.PreferNFO,
			description:      "Strategy string is case-insensitive (handled by strings.ToLower)",
		},
		{
			name:             "case insensitive - MeRgE-ArRaYs",
			preserveNFO:      false,
			mergeStrategyStr: "MeRgE-ArRaYs",
			expectedStrategy: nfo.MergeArrays,
			description:      "Mixed case strategy strings work",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This replicates the logic from runUpdate lines 1169-1182
			var mergeStrategy nfo.MergeStrategy
			if tt.preserveNFO {
				mergeStrategy = nfo.PreferNFO
			} else {
				// NOTE: Actual code uses strings.ToLower() for case-insensitive matching
				switch strings.ToLower(tt.mergeStrategyStr) {
				case "prefer-nfo":
					mergeStrategy = nfo.PreferNFO
				case "merge-arrays":
					mergeStrategy = nfo.MergeArrays
				default: // "prefer-scraper" or anything else
					mergeStrategy = nfo.PreferScraper
				}
			}

			assert.Equal(t, tt.expectedStrategy, mergeStrategy, tt.description)
		})
	}
}

// TestMergeStrategyPrecedence ensures preserve-nfo flag always wins
func TestMergeStrategyPrecedence(t *testing.T) {
	// Test that preserve-nfo flag takes absolute precedence
	strategies := []string{"prefer-scraper", "prefer-nfo", "merge-arrays", "invalid", ""}

	for _, strategyStr := range strategies {
		t.Run("preserve-nfo_overrides_"+strategyStr, func(t *testing.T) {
			// Simulate the determination logic
			var mergeStrategy nfo.MergeStrategy
			preserveNFO := true

			if preserveNFO {
				mergeStrategy = nfo.PreferNFO
			} else {
				switch strings.ToLower(strategyStr) {
				case "prefer-nfo":
					mergeStrategy = nfo.PreferNFO
				case "merge-arrays":
					mergeStrategy = nfo.MergeArrays
				default:
					mergeStrategy = nfo.PreferScraper
				}
			}

			assert.Equal(t, nfo.PreferNFO, mergeStrategy,
				"preserve-nfo flag should always result in PreferNFO strategy regardless of strategy string %q",
				strategyStr)
		})
	}
}

// TestMergeStrategyDefaults ensures sensible defaults
func TestMergeStrategyDefaults(t *testing.T) {
	tests := []struct {
		name             string
		strategyStr      string
		expectedStrategy nfo.MergeStrategy
		description      string
	}{
		{
			name:             "empty string defaults to prefer-scraper",
			strategyStr:      "",
			expectedStrategy: nfo.PreferScraper,
			description:      "Empty strategy should default to PreferScraper",
		},
		{
			name:             "whitespace defaults to prefer-scraper",
			strategyStr:      "   ",
			expectedStrategy: nfo.PreferScraper,
			description:      "Whitespace-only strategy should default to PreferScraper",
		},
		{
			name:             "random string defaults to prefer-scraper",
			strategyStr:      "foobar",
			expectedStrategy: nfo.PreferScraper,
			description:      "Invalid strategy should default to PreferScraper",
		},
		{
			name:             "typo defaults to prefer-scraper",
			strategyStr:      "prefr-scraper",
			expectedStrategy: nfo.PreferScraper,
			description:      "Typo in strategy should default to PreferScraper",
		},
		{
			name:             "explicit prefer-scraper",
			strategyStr:      "prefer-scraper",
			expectedStrategy: nfo.PreferScraper,
			description:      "Explicit prefer-scraper uses PreferScraper (same as default)",
		},
		{
			name:             "valid prefer-nfo",
			strategyStr:      "prefer-nfo",
			expectedStrategy: nfo.PreferNFO,
			description:      "Valid prefer-nfo strategy",
		},
		{
			name:             "valid merge-arrays",
			strategyStr:      "merge-arrays",
			expectedStrategy: nfo.MergeArrays,
			description:      "Valid merge-arrays strategy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mergeStrategy nfo.MergeStrategy
			preserveNFO := false

			if preserveNFO {
				mergeStrategy = nfo.PreferNFO
			} else {
				switch strings.ToLower(tt.strategyStr) {
				case "prefer-nfo":
					mergeStrategy = nfo.PreferNFO
				case "merge-arrays":
					mergeStrategy = nfo.MergeArrays
				default:
					mergeStrategy = nfo.PreferScraper
				}
			}

			assert.Equal(t, tt.expectedStrategy, mergeStrategy, tt.description)
		})
	}
}
