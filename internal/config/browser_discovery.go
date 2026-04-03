package config

import (
	"os"
	"os/exec"
	"runtime"
)

// AutoDiscoverBrowserBinary attempts to find Chrome/Chromium binary path.
// Returns empty string if not found.
func AutoDiscoverBrowserBinary() string {
	switch runtime.GOOS {
	case "darwin":
		return discoverBrowserDarwin()
	case "linux":
		return discoverBrowserLinux()
	case "windows":
		return discoverBrowserWindows()
	}
	return ""
}

func discoverBrowserDarwin() string {
	// macOS paths - check file existence
	paths := []string{
		"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		"/Applications/Chromium.app/Contents/MacOS/Chromium",
		"/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
	}
	for _, path := range paths {
		if fileExists(path) {
			return path
		}
	}

	// Try PATH lookup
	if path, err := exec.LookPath("google-chrome"); err == nil {
		return path
	}
	if path, err := exec.LookPath("chromium"); err == nil {
		return path
	}
	if path, err := exec.LookPath("chrome"); err == nil {
		return path
	}

	return ""
}

func discoverBrowserLinux() string {
	// Linux paths
	paths := []string{
		"/usr/bin/google-chrome",
		"/usr/bin/google-chrome-stable",
		"/usr/bin/chromium",
		"/usr/bin/chromium-browser",
		"/usr/bin/microsoft-edge",
	}
	for _, path := range paths {
		if fileExists(path) {
			return path
		}
	}

	// Try PATH lookup
	if path, err := exec.LookPath("google-chrome"); err == nil {
		return path
	}
	if path, err := exec.LookPath("chromium"); err == nil {
		return path
	}
	if path, err := exec.LookPath("chromium-browser"); err == nil {
		return path
	}

	return ""
}

func discoverBrowserWindows() string {
	// Windows paths
	paths := []string{
		`C:\Program Files\Google\Chrome\Application\chrome.exe`,
		`C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`,
		`C:\Program Files\Microsoft\Edge\Application\msedge.exe`,
		`C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`,
	}
	for _, path := range paths {
		if fileExists(path) {
			return path
		}
	}

	return ""
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// GetBrowserBinaryPath returns the browser binary path, auto-discovering if needed.
func GetBrowserBinaryPath(cfg BrowserConfig) string {
	if cfg.BinaryPath != "" {
		return cfg.BinaryPath
	}
	return AutoDiscoverBrowserBinary()
}
