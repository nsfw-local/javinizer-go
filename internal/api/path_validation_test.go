package api

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/javinizer/javinizer-go/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateScanPath(t *testing.T) {
	// Create temp directory for testing
	tempDir := t.TempDir()
	allowedDir := filepath.Join(tempDir, "allowed")
	require.NoError(t, os.Mkdir(allowedDir, 0755))

	tests := []struct {
		name          string
		inputPath     string
		securityCfg   *config.SecurityConfig
		expectedError bool
		errorContains string
	}{
		{
			name:      "valid path within allowed directory",
			inputPath: allowedDir,
			securityCfg: &config.SecurityConfig{
				AllowedDirectories: []string{tempDir},
				DeniedDirectories:  []string{},
				MaxFilesPerScan:    10000,
				ScanTimeoutSeconds: 30,
			},
			expectedError: false,
		},
		{
			name:      "valid path - no allowlist restriction",
			inputPath: tempDir,
			securityCfg: &config.SecurityConfig{
				AllowedDirectories: []string{},
				DeniedDirectories:  []string{},
				MaxFilesPerScan:    10000,
				ScanTimeoutSeconds: 30,
			},
			expectedError: false,
		},
		{
			name:      "path traversal attempt with ../",
			inputPath: filepath.Join(allowedDir, "../etc/passwd"),
			securityCfg: &config.SecurityConfig{
				AllowedDirectories: []string{allowedDir},
				DeniedDirectories:  []string{},
				MaxFilesPerScan:    10000,
				ScanTimeoutSeconds: 30,
			},
			expectedError: true,
			errorContains: "does not exist", // Path validation happens before allowlist check
		},
		{
			name:      "absolute path outside allowed directory",
			inputPath: tempDir,
			securityCfg: &config.SecurityConfig{
				AllowedDirectories: []string{allowedDir},
				DeniedDirectories:  []string{},
				MaxFilesPerScan:    10000,
				ScanTimeoutSeconds: 30,
			},
			expectedError: true,
			errorContains: "outside allowed directories",
		},
		{
			name:      "path with multiple ../ sequences",
			inputPath: filepath.Join(tempDir, "../../etc"),
			securityCfg: &config.SecurityConfig{
				AllowedDirectories: []string{tempDir},
				DeniedDirectories:  []string{},
				MaxFilesPerScan:    10000,
				ScanTimeoutSeconds: 30,
			},
			expectedError: true,
			errorContains: "does not exist", // Path validation happens before allowlist check
		},
		{
			name:      "nonexistent path",
			inputPath: "/nonexistent/path/12345",
			securityCfg: &config.SecurityConfig{
				AllowedDirectories: []string{},
				DeniedDirectories:  []string{},
				MaxFilesPerScan:    10000,
				ScanTimeoutSeconds: 30,
			},
			expectedError: true,
			errorContains: "does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validPath, err := validateScanPath(tt.inputPath, tt.securityCfg)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, validPath)
				// Verify returned path is absolute and clean
				assert.True(t, filepath.IsAbs(validPath))
			}
		})
	}
}

func TestValidateScanPath_SystemDirectories(t *testing.T) {
	// Test that system directories are blocked regardless of allowlist
	systemDirs := []string{
		"/etc",
		"/var/log",
		"/usr/bin",
	}

	// Add Windows-specific test paths if on Windows
	if runtime.GOOS == "windows" {
		systemDirs = append(systemDirs, "C:\\Windows")
	}

	// Add macOS-specific test paths if on macOS
	if runtime.GOOS == "darwin" {
		systemDirs = append(systemDirs, "/System")
	}

	securityCfg := &config.SecurityConfig{
		AllowedDirectories: []string{},
		DeniedDirectories:  []string{},
		MaxFilesPerScan:    10000,
		ScanTimeoutSeconds: 30,
	}

	for _, dir := range systemDirs {
		t.Run("blocks "+dir, func(t *testing.T) {
			// Skip if directory doesn't exist (won't be blocked if it doesn't exist)
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				t.Skip("System directory doesn't exist on this platform")
			}

			_, err := validateScanPath(dir, securityCfg)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "system directory")
		})
	}
}

func TestValidateScanPath_FileVsDirectory(t *testing.T) {
	// Create temp file
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "testfile.txt")
	require.NoError(t, os.WriteFile(tempFile, []byte("test"), 0644))

	securityCfg := &config.SecurityConfig{
		AllowedDirectories: []string{},
		DeniedDirectories:  []string{},
		MaxFilesPerScan:    10000,
		ScanTimeoutSeconds: 30,
	}

	t.Run("rejects file path", func(t *testing.T) {
		_, err := validateScanPath(tempFile, securityCfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not a directory")
	})

	t.Run("accepts directory path", func(t *testing.T) {
		validPath, err := validateScanPath(tempDir, securityCfg)
		assert.NoError(t, err)
		// Compare canonical paths since validateScanPath returns canonical path
		expectedPath, _ := filepath.EvalSymlinks(tempDir)
		assert.Equal(t, expectedPath, validPath)
	})
}

func TestGetDeniedDirectories(t *testing.T) {
	denied := getDeniedDirectories()

	// Should always include these cross-platform directories
	assert.Contains(t, denied, "/etc")
	assert.Contains(t, denied, "/var/log")

	// Platform-specific checks
	if runtime.GOOS == "windows" {
		assert.Contains(t, denied, "C:\\Windows")
	}

	if runtime.GOOS == "darwin" {
		assert.Contains(t, denied, "/System")
	}
}

func BenchmarkValidateScanPath(b *testing.B) {
	tempDir := b.TempDir()
	testPath := tempDir
	securityCfg := &config.SecurityConfig{
		AllowedDirectories: []string{tempDir},
		DeniedDirectories:  []string{},
		MaxFilesPerScan:    10000,
		ScanTimeoutSeconds: 30,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = validateScanPath(testPath, securityCfg)
	}
}

func TestValidateScanPath_EdgeCases(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name          string
		inputPath     string
		securityCfg   *config.SecurityConfig
		expectedError bool
		errorContains string
	}{
		{
			name:      "empty path defaults to current directory",
			inputPath: "",
			securityCfg: &config.SecurityConfig{
				AllowedDirectories: []string{},
				DeniedDirectories:  []string{},
				MaxFilesPerScan:    10000,
				ScanTimeoutSeconds: 30,
			},
			expectedError: false, // Empty path becomes "." which resolves to current directory
		},
		{
			name:      "path with trailing slash",
			inputPath: tempDir + "/",
			securityCfg: &config.SecurityConfig{
				AllowedDirectories: []string{},
				DeniedDirectories:  []string{},
				MaxFilesPerScan:    10000,
				ScanTimeoutSeconds: 30,
			},
			expectedError: false,
		},
		{
			name:      "path with ./ prefix",
			inputPath: "./",
			securityCfg: &config.SecurityConfig{
				AllowedDirectories: []string{},
				DeniedDirectories:  []string{},
				MaxFilesPerScan:    10000,
				ScanTimeoutSeconds: 30,
			},
			expectedError: false, // ./ resolves to current directory
		},
		{
			name:      "relative path cleaned to absolute",
			inputPath: tempDir,
			securityCfg: &config.SecurityConfig{
				AllowedDirectories: []string{},
				DeniedDirectories:  []string{},
				MaxFilesPerScan:    10000,
				ScanTimeoutSeconds: 30,
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validPath, err := validateScanPath(tt.inputPath, tt.securityCfg)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.True(t, filepath.IsAbs(validPath))
			}
		})
	}
}

func TestNormalizeWindowsPath(t *testing.T) {
	// Only run on Windows or skip if testing Windows-specific behavior
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows-specific tests on non-Windows platform")
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Win32 namespace (\\?\) tests
		{
			name:     "Win32 extended path - C drive",
			input:    `\\?\C:\Windows\System32`,
			expected: `C:\Windows\System32`,
		},
		{
			name:     "Win32 extended path - D drive",
			input:    `\\?\D:\Data\Files`,
			expected: `D:\Data\Files`,
		},
		{
			name:     "Win32 UNC path",
			input:    `\\?\UNC\server\share\folder`,
			expected: `\\server\share\folder`,
		},
		{
			name:     "Win32 extended path - lowercase",
			input:    `\\?\c:\windows`,
			expected: `c:\windows`,
		},
		{
			name:     "Win32 UNC - lowercase",
			input:    `\\?\unc\server\share`,
			expected: `\\server\share`,
		},
		// Mixed case variants
		{
			name:     "Win32 extended path - mixed case prefix",
			input:    `\\?\C:\Windows`,
			expected: `C:\Windows`,
		},
		{
			name:     "Win32 UNC - Unc",
			input:    `\\?\Unc\server\share`,
			expected: `\\server\share`,
		},
		{
			name:     "Win32 UNC - uNc",
			input:    `\\?\uNc\server\share`,
			expected: `\\server\share`,
		},
		{
			name:     "Win32 UNC - UnC",
			input:    `\\?\UnC\server\share`,
			expected: `\\server\share`,
		},

		// NT namespace (\??\) tests
		{
			name:     "NT namespace path - C drive",
			input:    `\??\C:\Windows\System32`,
			expected: `C:\Windows\System32`,
		},
		{
			name:     "NT namespace path - D drive",
			input:    `\??\D:\Data`,
			expected: `D:\Data`,
		},
		{
			name:     "NT namespace UNC",
			input:    `\??\UNC\server\share\folder`,
			expected: `\\server\share\folder`,
		},
		{
			name:     "NT namespace - lowercase",
			input:    `\??\c:\windows`,
			expected: `c:\windows`,
		},
		{
			name:     "NT namespace UNC - lowercase",
			input:    `\??\unc\server\share`,
			expected: `\\server\share`,
		},
		// Mixed case variants
		{
			name:     "NT namespace UNC - Unc",
			input:    `\??\Unc\server\share`,
			expected: `\\server\share`,
		},
		{
			name:     "NT namespace UNC - uNc",
			input:    `\??\uNc\server\share`,
			expected: `\\server\share`,
		},

		// Device namespace (\\.\) tests
		{
			name:     "Device namespace - C drive",
			input:    `\\.\C:\Windows\System32`,
			expected: `C:\Windows\System32`,
		},
		{
			name:     "Device namespace - D drive",
			input:    `\\.\D:\Data`,
			expected: `D:\Data`,
		},
		{
			name:     "Device namespace UNC",
			input:    `\\.\UNC\server\share\folder`,
			expected: `\\server\share\folder`,
		},
		{
			name:     "Device namespace - lowercase",
			input:    `\\.\c:\windows`,
			expected: `c:\windows`,
		},
		{
			name:     "Device namespace UNC - lowercase",
			input:    `\\.\unc\server\share`,
			expected: `\\server\share`,
		},
		// Mixed case variants
		{
			name:     "Device namespace UNC - Unc",
			input:    `\\.\Unc\server\share`,
			expected: `\\server\share`,
		},
		{
			name:     "Device namespace UNC - uNc",
			input:    `\\.\uNc\server\share`,
			expected: `\\server\share`,
		},

		// Regular paths (no normalization needed)
		{
			name:     "Regular path - absolute",
			input:    `C:\Windows\System32`,
			expected: `C:\Windows\System32`,
		},
		{
			name:     "Regular UNC path",
			input:    `\\server\share\folder`,
			expected: `\\server\share\folder`,
		},
		{
			name:     "Relative path",
			input:    `folder\file.txt`,
			expected: `folder\file.txt`,
		},

		// Edge cases
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Just prefix - Win32",
			input:    `\\?\`,
			expected: ``,
		},
		{
			name:     "Just prefix - NT",
			input:    `\??\`,
			expected: ``,
		},
		{
			name:     "Just prefix - Device",
			input:    `\\.\`,
			expected: ``,
		},
		{
			name:     "Short path with prefix",
			input:    `\\?\C:`,
			expected: `C:`,
		},
		{
			name:     "UNC with only server",
			input:    `\\?\UNC\server`,
			expected: `\\server`,
		},

		// Malformed inputs (should pass through unchanged or strip prefix)
		{
			name:     "Prefix without path",
			input:    `\\?\UNC`,
			expected: `\\?\UNC`, // Too short, no match
		},
		{
			name:     "Prefix with slash only",
			input:    `\\?\UNC\`,
			expected: `\\`, // Strips prefix, keeps UNC \\ prefix
		},
		{
			name:     "Mixed prefix styles",
			input:    `\\?\??\C:\Windows`, // Malformed, but should strip first prefix
			expected: `??\C:\Windows`,
		},

		// Special characters in path
		{
			name:     "Path with spaces",
			input:    `\\?\C:\Program Files\App`,
			expected: `C:\Program Files\App`,
		},
		{
			name:     "Path with special chars",
			input:    `\\?\C:\Users\name@domain.com\Documents`,
			expected: `C:\Users\name@domain.com\Documents`,
		},

		// Volume GUIDs (should strip prefix)
		{
			name:     "Volume GUID path - Win32",
			input:    `\\?\Volume{12345678-1234-1234-1234-123456789012}\Windows`,
			expected: `Volume{12345678-1234-1234-1234-123456789012}\Windows`,
		},
		{
			name:     "Volume GUID path - NT namespace",
			input:    `\??\Volume{12345678-1234-1234-1234-123456789012}\Windows`,
			expected: `Volume{12345678-1234-1234-1234-123456789012}\Windows`,
		},
		{
			name:     "Volume GUID path - Device namespace",
			input:    `\\.\Volume{12345678-1234-1234-1234-123456789012}\Windows`,
			expected: `Volume{12345678-1234-1234-1234-123456789012}\Windows`,
		},

		// GLOBALROOT and shadow copy volumes (common attack vectors)
		{
			name:     "GLOBALROOT shadow copy - Win32",
			input:    `\\?\GLOBALROOT\Device\HarddiskVolumeShadowCopy1\Windows`,
			expected: `GLOBALROOT\Device\HarddiskVolumeShadowCopy1\Windows`,
		},
		{
			name:     "GLOBALROOT shadow copy - NT namespace",
			input:    `\??\GLOBALROOT\Device\HarddiskVolumeShadowCopy1\Windows`,
			expected: `GLOBALROOT\Device\HarddiskVolumeShadowCopy1\Windows`,
		},
		{
			name:     "GLOBALROOT shadow copy - Device namespace",
			input:    `\\.\GLOBALROOT\Device\HarddiskVolumeShadowCopy1\Windows`,
			expected: `GLOBALROOT\Device\HarddiskVolumeShadowCopy1\Windows`,
		},
		{
			name:     "GLOBALROOT with mixed case - Win32",
			input:    `\\?\GlobalRoot\Device\HarddiskVolumeShadowCopy2\System`,
			expected: `GlobalRoot\Device\HarddiskVolumeShadowCopy2\System`,
		},

		// Device namespace - named pipes, mailslots, reserved devices
		// Test all three namespace prefixes for comprehensive coverage
		{
			name:     "Named pipe - Device namespace",
			input:    `\\.\pipe\mypipe`,
			expected: `pipe\mypipe`,
		},
		{
			name:     "Named pipe - Win32 namespace",
			input:    `\\?\pipe\mypipe`,
			expected: `pipe\mypipe`,
		},
		{
			name:     "Named pipe - NT namespace",
			input:    `\??\pipe\mypipe`,
			expected: `pipe\mypipe`,
		},
		{
			name:     "Mailslot - Device namespace",
			input:    `\\.\mailslot\myslot`,
			expected: `mailslot\myslot`,
		},
		{
			name:     "Mailslot - Win32 namespace",
			input:    `\\?\mailslot\myslot`,
			expected: `mailslot\myslot`,
		},
		{
			name:     "Mailslot - NT namespace",
			input:    `\??\mailslot\myslot`,
			expected: `mailslot\myslot`,
		},
		{
			name:     "COM port device - Device namespace",
			input:    `\\.\COM1`,
			expected: `COM1`,
		},
		{
			name:     "COM port device - Win32 namespace",
			input:    `\\?\COM1`,
			expected: `COM1`,
		},
		{
			name:     "COM port device - NT namespace",
			input:    `\??\COM1`,
			expected: `COM1`,
		},
		{
			name:     "Physical drive device - Device namespace",
			input:    `\\.\PhysicalDrive0`,
			expected: `PhysicalDrive0`,
		},
		{
			name:     "Physical drive device - Win32 namespace",
			input:    `\\?\PhysicalDrive0`,
			expected: `PhysicalDrive0`,
		},
		{
			name:     "Physical drive device - NT namespace",
			input:    `\??\PhysicalDrive0`,
			expected: `PhysicalDrive0`,
		},
		{
			name:     "CDROM device - Device namespace",
			input:    `\\.\CdRom0`,
			expected: `CdRom0`,
		},
		{
			name:     "CDROM device - Win32 namespace",
			input:    `\\?\CdRom0`,
			expected: `CdRom0`,
		},
		{
			name:     "CDROM device - NT namespace",
			input:    `\??\CdRom0`,
			expected: `CdRom0`,
		},
		{
			name:     "Harddisk device - Device namespace",
			input:    `\\.\Harddisk0\Partition1`,
			expected: `Harddisk0\Partition1`,
		},
		{
			name:     "Harddisk device - Win32 namespace",
			input:    `\\?\Harddisk0\Partition1`,
			expected: `Harddisk0\Partition1`,
		},
		{
			name:     "Harddisk device - NT namespace",
			input:    `\??\Harddisk0\Partition1`,
			expected: `Harddisk0\Partition1`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeWindowsPath(tt.input)
			assert.Equal(t, tt.expected, result, "normalizeWindowsPath(%q) = %q, expected %q", tt.input, result, tt.expected)
		})
	}
}

func TestNormalizeWindowsPath_NonWindows(t *testing.T) {
	// Test that normalization is a no-op on non-Windows platforms
	if runtime.GOOS == "windows" {
		t.Skip("Skipping non-Windows test on Windows platform")
	}

	tests := []string{
		`\\?\C:\Windows`,
		`\??\C:\Windows`,
		`\\.\C:\Windows`,
		`/etc/passwd`,
		`/var/log`,
		`~/Documents`,
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			result := normalizeWindowsPath(input)
			assert.Equal(t, input, result, "On non-Windows, normalizeWindowsPath should return input unchanged")
		})
	}
}

func BenchmarkNormalizeWindowsPath(b *testing.B) {
	if runtime.GOOS != "windows" {
		b.Skip("Skipping Windows benchmark on non-Windows platform")
	}

	testCases := []string{
		`\\?\C:\Windows\System32`,
		`\??\C:\Windows\System32`,
		`\\.\C:\Windows\System32`,
		`\\?\UNC\server\share`,
		`C:\Windows\System32`,
	}

	for _, tc := range testCases {
		b.Run(tc, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = normalizeWindowsPath(tc)
			}
		})
	}
}
