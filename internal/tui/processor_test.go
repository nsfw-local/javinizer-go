package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSetGetCustomScrapers_DefensiveCopy verifies defensive copying behavior
func TestSetGetCustomScrapers_DefensiveCopy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setScrapers    []string
		modifyOriginal bool
		expectedGet    []string
	}{
		{
			name:           "nil scrapers",
			setScrapers:    nil,
			modifyOriginal: false,
			expectedGet:    nil,
		},
		{
			name:           "empty scrapers",
			setScrapers:    []string{},
			modifyOriginal: false,
			expectedGet:    nil, // Empty slice becomes nil after defensive copy
		},
		{
			name:           "single scraper",
			setScrapers:    []string{"r18dev"},
			modifyOriginal: false,
			expectedGet:    []string{"r18dev"},
		},
		{
			name:           "multiple scrapers",
			setScrapers:    []string{"r18dev", "dmm"},
			modifyOriginal: false,
			expectedGet:    []string{"r18dev", "dmm"},
		},
		{
			name:           "modify original after set",
			setScrapers:    []string{"r18dev", "dmm"},
			modifyOriginal: true,
			expectedGet:    []string{"r18dev", "dmm"}, // Should still be original
		},
	}

	for _, tt := range tests {
		tt := tt // Rebind for parallel subtest
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create minimal coordinator (only need to test Set/Get)
			pc := &ProcessingCoordinator{}

			// Set scrapers
			pc.SetCustomScrapers(tt.setScrapers)

			// Modify original if requested
			if tt.modifyOriginal && tt.setScrapers != nil && len(tt.setScrapers) > 0 {
				tt.setScrapers[0] = "modified"
			}

			// Get scrapers
			got := pc.GetCustomScrapers()

			// Verify
			assert.Equal(t, tt.expectedGet, got)

			// Verify modifying returned slice doesn't affect internal state
			if got != nil && len(got) > 0 {
				got[0] = "external-modification"
				gotAgain := pc.GetCustomScrapers()
				assert.Equal(t, tt.expectedGet, gotAgain, "Internal state should not be affected by external modification")
			}
		})
	}
}

// TestSetOptions verifies configuration methods update state correctly
func TestSetOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		scrape           bool
		download         bool
		organize         bool
		nfo              bool
		expectedScrape   bool
		expectedDownload bool
		expectedOrganize bool
		expectedNFO      bool
	}{
		{
			name:             "all enabled",
			scrape:           true,
			download:         true,
			organize:         true,
			nfo:              true,
			expectedScrape:   true,
			expectedDownload: true,
			expectedOrganize: true,
			expectedNFO:      true,
		},
		{
			name:             "all disabled",
			scrape:           false,
			download:         false,
			organize:         false,
			nfo:              false,
			expectedScrape:   false,
			expectedDownload: false,
			expectedOrganize: false,
			expectedNFO:      false,
		},
		{
			name:             "selective: scrape and organize only",
			scrape:           true,
			download:         false,
			organize:         true,
			nfo:              false,
			expectedScrape:   true,
			expectedDownload: false,
			expectedOrganize: true,
			expectedNFO:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pc := &ProcessingCoordinator{}

			pc.SetOptions(tt.scrape, tt.download, tt.organize, tt.nfo)

			assert.Equal(t, tt.expectedScrape, pc.scrapeEnabled)
			assert.Equal(t, tt.expectedDownload, pc.downloadEnabled)
			assert.Equal(t, tt.expectedOrganize, pc.organizeEnabled)
			assert.Equal(t, tt.expectedNFO, pc.nfoEnabled)
		})
	}
}

// TestSetDryRun verifies dry-run flag is set correctly
func TestSetDryRun(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		dryRun   bool
		expected bool
	}{
		{
			name:     "enable dry-run",
			dryRun:   true,
			expected: true,
		},
		{
			name:     "disable dry-run",
			dryRun:   false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pc := &ProcessingCoordinator{}

			pc.SetDryRun(tt.dryRun)

			assert.Equal(t, tt.expected, pc.dryRun)
		})
	}
}

// TestSetDestPath verifies destination path is set correctly
func TestSetDestPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		destPath string
		expected string
	}{
		{
			name:     "set destination path",
			destPath: "/videos/organized",
			expected: "/videos/organized",
		},
		{
			name:     "empty destination path",
			destPath: "",
			expected: "",
		},
		{
			name:     "relative destination path",
			destPath: "./organized",
			expected: "./organized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pc := &ProcessingCoordinator{}

			pc.SetDestPath(tt.destPath)

			assert.Equal(t, tt.expected, pc.destPath)
		})
	}
}

// TestSetMoveFiles verifies move files flag is set correctly
func TestSetMoveFiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		moveFiles bool
		expected  bool
	}{
		{
			name:      "enable move files",
			moveFiles: true,
			expected:  true,
		},
		{
			name:      "disable move files (copy mode)",
			moveFiles: false,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pc := &ProcessingCoordinator{}

			pc.SetMoveFiles(tt.moveFiles)

			assert.Equal(t, tt.expected, pc.moveFiles)
		})
	}
}

// TestSetScrapeEnabled verifies scrape enabled flag is set correctly
func TestSetScrapeEnabled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		enabled  bool
		expected bool
	}{
		{
			name:     "enable scraping",
			enabled:  true,
			expected: true,
		},
		{
			name:     "disable scraping",
			enabled:  false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pc := &ProcessingCoordinator{}

			pc.SetScrapeEnabled(tt.enabled)

			assert.Equal(t, tt.expected, pc.scrapeEnabled)
		})
	}
}

// TestSetDownloadEnabled verifies download enabled flag is set correctly
func TestSetDownloadEnabled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		enabled  bool
		expected bool
	}{
		{
			name:     "enable downloads",
			enabled:  true,
			expected: true,
		},
		{
			name:     "disable downloads",
			enabled:  false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pc := &ProcessingCoordinator{}

			pc.SetDownloadEnabled(tt.enabled)

			assert.Equal(t, tt.expected, pc.downloadEnabled)
		})
	}
}

// TestSetOrganizeEnabled verifies organize enabled flag is set correctly
func TestSetOrganizeEnabled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		enabled  bool
		expected bool
	}{
		{
			name:     "enable organize",
			enabled:  true,
			expected: true,
		},
		{
			name:     "disable organize",
			enabled:  false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pc := &ProcessingCoordinator{}

			pc.SetOrganizeEnabled(tt.enabled)

			assert.Equal(t, tt.expected, pc.organizeEnabled)
		})
	}
}

// TestSetNFOEnabled verifies NFO enabled flag is set correctly
func TestSetNFOEnabled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		enabled  bool
		expected bool
	}{
		{
			name:     "enable NFO generation",
			enabled:  true,
			expected: true,
		},
		{
			name:     "disable NFO generation",
			enabled:  false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pc := &ProcessingCoordinator{}

			pc.SetNFOEnabled(tt.enabled)

			assert.Equal(t, tt.expected, pc.nfoEnabled)
		})
	}
}

// TestSetForceUpdate verifies force update flag is set correctly
func TestSetForceUpdate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		forceUpdate bool
		expected    bool
	}{
		{
			name:        "enable force update",
			forceUpdate: true,
			expected:    true,
		},
		{
			name:        "disable force update",
			forceUpdate: false,
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pc := &ProcessingCoordinator{}

			pc.SetForceUpdate(tt.forceUpdate)

			assert.Equal(t, tt.expected, pc.forceUpdate)
		})
	}
}

// TestSetForceRefresh verifies force refresh flag is set correctly
func TestSetForceRefresh(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		forceRefresh bool
		expected     bool
	}{
		{
			name:         "enable force refresh",
			forceRefresh: true,
			expected:     true,
		},
		{
			name:         "disable force refresh",
			forceRefresh: false,
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pc := &ProcessingCoordinator{}

			pc.SetForceRefresh(tt.forceRefresh)

			assert.Equal(t, tt.expected, pc.forceRefresh)
		})
	}
}

// TestNewProcessingCoordinator verifies constructor initializes with correct defaults
func TestNewProcessingCoordinator(t *testing.T) {
	t.Parallel()

	// Create coordinator with nil dependencies (just testing initialization)
	pc := NewProcessingCoordinator(
		nil, // pool
		nil, // progressTracker
		nil, // movieRepo
		nil, // registry
		nil, // aggregator
		nil, // downloader
		nil, // organizer
		nil, // nfoGenerator
		"/dest/path",
		true, // moveFiles
	)

	// Verify initialization
	assert.NotNil(t, pc)
	assert.Equal(t, "/dest/path", pc.destPath)
	assert.Equal(t, true, pc.moveFiles)
	assert.Equal(t, true, pc.scrapeEnabled, "scrapeEnabled should default to true")
	assert.Equal(t, true, pc.downloadEnabled, "downloadEnabled should default to true")
	assert.Equal(t, true, pc.organizeEnabled, "organizeEnabled should default to true")
	assert.Equal(t, true, pc.nfoEnabled, "nfoEnabled should default to true")
	assert.Equal(t, false, pc.dryRun, "dryRun should default to false")
	assert.Equal(t, false, pc.forceUpdate, "forceUpdate should default to false")
	assert.Equal(t, false, pc.forceRefresh, "forceRefresh should default to false")
	assert.Nil(t, pc.customScraperPriority, "customScraperPriority should default to nil")
}
