package main

import (
	"path/filepath"
	"testing"

	"github.com/javinizer/javinizer-go/internal/matcher"
	"github.com/javinizer/javinizer-go/internal/models"
	"github.com/javinizer/javinizer-go/internal/scanner"
	"github.com/stretchr/testify/assert"
)

func TestConstructNFOPath(t *testing.T) {
	tests := []struct {
		name         string
		match        matcher.MatchResult
		movie        *models.Movie
		perFile      bool
		expectedName string // Just the filename part
	}{
		{
			name: "basic single-part file",
			match: matcher.MatchResult{
				File: scanner.FileInfo{
					Dir: "/test/videos",
				},
				IsMultiPart: false,
			},
			movie: &models.Movie{
				ID: "IPX-535",
			},
			perFile:      false,
			expectedName: "IPX-535.nfo",
		},
		{
			name: "multi-part with perFile enabled",
			match: matcher.MatchResult{
				File: scanner.FileInfo{
					Dir: "/test/videos",
				},
				IsMultiPart: true,
				PartSuffix:  "-pt1",
			},
			movie: &models.Movie{
				ID: "IPX-535",
			},
			perFile:      true,
			expectedName: "IPX-535-pt1.nfo",
		},
		{
			name: "multi-part with perFile disabled",
			match: matcher.MatchResult{
				File: scanner.FileInfo{
					Dir: "/test/videos",
				},
				IsMultiPart: true,
				PartSuffix:  "-pt2",
			},
			movie: &models.Movie{
				ID: "IPX-535",
			},
			perFile:      false,
			expectedName: "IPX-535.nfo",
		},
		{
			name: "sanitization of special characters",
			match: matcher.MatchResult{
				File: scanner.FileInfo{
					Dir: "/test/videos",
				},
				IsMultiPart: false,
			},
			movie: &models.Movie{
				ID: "TEST:123*456?",
			},
			perFile:      false,
			expectedName: "TEST -123456.nfo", // : replaced with " -", * and ? removed
		},
		{
			name: "path separators replaced with hyphens",
			match: matcher.MatchResult{
				File: scanner.FileInfo{
					Dir: "/test/videos",
				},
				IsMultiPart: false,
			},
			movie: &models.Movie{
				ID: "../../../etc/passwd",
			},
			perFile:      false,
			expectedName: "-..-..-etc-passwd.nfo", // Slashes replaced with hyphens, dots preserved
		},
		{
			name: "backslashes replaced with hyphens (Windows)",
			match: matcher.MatchResult{
				File: scanner.FileInfo{
					Dir: "/test/videos",
				},
				IsMultiPart: false,
			},
			movie: &models.Movie{
				ID: "..\\..\\..\\windows\\system32",
			},
			perFile:      false,
			expectedName: "-..-..-windows-system32.nfo", // Backslashes replaced with hyphens
		},
		{
			name: "null bytes removed",
			match: matcher.MatchResult{
				File: scanner.FileInfo{
					Dir: "/test/videos",
				},
				IsMultiPart: false,
			},
			movie: &models.Movie{
				ID: "TEST\x00123",
			},
			perFile:      false,
			expectedName: "TEST123.nfo",
		},
		{
			name: "only non-printable characters - fallback to metadata",
			match: matcher.MatchResult{
				File: scanner.FileInfo{
					Dir: "/test/videos",
				},
				IsMultiPart: false,
			},
			movie: &models.Movie{
				ID: "\x00\x01\x02\x03",
			},
			perFile:      false,
			expectedName: "metadata.nfo", // Falls back when sanitization results in empty
		},
		{
			name: "unicode characters preserved",
			match: matcher.MatchResult{
				File: scanner.FileInfo{
					Dir: "/test/videos",
				},
				IsMultiPart: false,
			},
			movie: &models.Movie{
				ID: "IPX-535-美女",
			},
			perFile:      false,
			expectedName: "IPX-535-美女.nfo", // Unicode should be preserved
		},
		{
			name: "leading/trailing spaces trimmed",
			match: matcher.MatchResult{
				File: scanner.FileInfo{
					Dir: "/test/videos",
				},
				IsMultiPart: false,
			},
			movie: &models.Movie{
				ID: "  IPX-535  ",
			},
			perFile:      false,
			expectedName: "IPX-535.nfo", // Spaces trimmed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call function
			result := constructNFOPath(tt.match, tt.movie, tt.perFile)

			// Verify directory is preserved
			assert.Equal(t, tt.match.File.Dir, filepath.Dir(result), "directory should be preserved")

			// Verify filename matches expected
			actualFilename := filepath.Base(result)
			assert.Equal(t, tt.expectedName, actualFilename, "filename should match expected")

			// Verify extension is always .nfo
			assert.Equal(t, ".nfo", filepath.Ext(result), "extension should always be .nfo")
		})
	}
}

func TestConstructNFOPath_SecurityInvariants(t *testing.T) {
	tests := []struct {
		name      string
		movieID   string
		expectDir string
	}{
		{
			name:      "absolute path injection attempt",
			movieID:   "/etc/passwd",
			expectDir: "/test/videos",
		},
		{
			name:      "relative path traversal",
			movieID:   "../../secret",
			expectDir: "/test/videos",
		},
		{
			name:      "windows path separator",
			movieID:   "..\\..\\secret",
			expectDir: "/test/videos",
		},
		{
			name:      "mixed separators",
			movieID:   "../..\\/secret",
			expectDir: "/test/videos",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match := matcher.MatchResult{
				File: scanner.FileInfo{
					Dir: tt.expectDir,
				},
				IsMultiPart: false,
			}

			movie := &models.Movie{
				ID: tt.movieID,
			}

			result := constructNFOPath(match, movie, false)

			// CRITICAL: Result must be within the expected directory
			resultDir := filepath.Dir(result)
			assert.Equal(t, tt.expectDir, resultDir, "NFO path must stay within source directory")

			// CRITICAL: Result must not contain path separators in filename
			// Note: Sanitization replaces / and \ with -, but dots are preserved
			// Security comes from filepath.Join() normalizing the path, not from removing dots
			filename := filepath.Base(result)
			assert.NotContains(t, filename, "/", "filename must not contain forward slash")
			assert.NotContains(t, filename, "\\", "filename must not contain backslash")
		})
	}
}

func TestConstructNFOPath_MultiPartScenarios(t *testing.T) {
	baseMatch := matcher.MatchResult{
		File: scanner.FileInfo{
			Dir: "/test/videos",
		},
	}

	movie := &models.Movie{
		ID: "IPX-535",
	}

	tests := []struct {
		name        string
		isMultiPart bool
		partSuffix  string
		perFile     bool
		expected    string
	}{
		{
			name:        "single part - perFile true",
			isMultiPart: false,
			partSuffix:  "",
			perFile:     true,
			expected:    "IPX-535.nfo",
		},
		{
			name:        "single part - perFile false",
			isMultiPart: false,
			partSuffix:  "",
			perFile:     false,
			expected:    "IPX-535.nfo",
		},
		{
			name:        "multi-part pt1 - perFile true",
			isMultiPart: true,
			partSuffix:  "-pt1",
			perFile:     true,
			expected:    "IPX-535-pt1.nfo",
		},
		{
			name:        "multi-part pt1 - perFile false",
			isMultiPart: true,
			partSuffix:  "-pt1",
			perFile:     false,
			expected:    "IPX-535.nfo",
		},
		{
			name:        "multi-part pt2 - perFile true",
			isMultiPart: true,
			partSuffix:  "-pt2",
			perFile:     true,
			expected:    "IPX-535-pt2.nfo",
		},
		{
			name:        "multi-part CD1 - perFile true",
			isMultiPart: true,
			partSuffix:  "-CD1",
			perFile:     true,
			expected:    "IPX-535-CD1.nfo",
		},
		{
			name:        "multi-part CD1 - perFile false",
			isMultiPart: true,
			partSuffix:  "-CD1",
			perFile:     false,
			expected:    "IPX-535.nfo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match := baseMatch
			match.IsMultiPart = tt.isMultiPart
			match.PartSuffix = tt.partSuffix

			result := constructNFOPath(match, movie, tt.perFile)
			actualFilename := filepath.Base(result)

			assert.Equal(t, tt.expected, actualFilename, "filename should match expected for multi-part scenario")
		})
	}
}
