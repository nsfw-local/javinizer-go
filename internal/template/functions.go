package template

import (
	"strings"
	"unicode"
)

// SanitizeFilename removes or replaces characters that are invalid in filenames
// This is a standalone function for use by callers
func SanitizeFilename(s string) string {
	// Replace invalid characters with safe alternatives
	replacements := map[rune]string{
		'/':  "-",
		'\\': "-",
		':':  " -",
		'*':  "",
		'?':  "",
		'"':  "'",
		'<':  "(",
		'>':  ")",
		'|':  "-",
	}

	var result strings.Builder
	for _, r := range s {
		if replacement, exists := replacements[r]; exists {
			result.WriteString(replacement)
		} else if unicode.IsPrint(r) {
			result.WriteRune(r)
		}
	}

	// Trim spaces and dots from ends (Windows doesn't like these)
	trimmed := strings.Trim(result.String(), " .")

	// Collapse multiple spaces
	for strings.Contains(trimmed, "  ") {
		trimmed = strings.ReplaceAll(trimmed, "  ", " ")
	}

	return trimmed
}

// SanitizeFolderPath sanitizes a folder path preserving slashes
func SanitizeFolderPath(s string) string {
	// Replace invalid characters but preserve forward slashes
	replacements := map[rune]string{
		'\\': "/", // Convert backslash to forward slash
		':':  " -",
		'*':  "",
		'?':  "",
		'"':  "'",
		'<':  "(",
		'>':  ")",
		'|':  "-",
	}

	var result strings.Builder
	for _, r := range s {
		if replacement, exists := replacements[r]; exists {
			result.WriteString(replacement)
		} else if unicode.IsPrint(r) {
			result.WriteRune(r)
		}
	}

	return result.String()
}
