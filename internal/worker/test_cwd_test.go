package worker

import (
	"os"
	"sync"
	"testing"
)

var testCWDMutex sync.Mutex

// chdirToTempDir isolates tests that use relative paths like "data/...".
// go test runs packages with their package directory as CWD, so without this
// helper tests can leak artifacts into internal/worker/data.
func chdirToTempDir(t *testing.T) string {
	t.Helper()

	testCWDMutex.Lock()

	originalWD, err := os.Getwd()
	if err != nil {
		testCWDMutex.Unlock()
		t.Fatalf("failed to get working directory: %v", err)
	}

	tempDir := t.TempDir()
	if err := os.Chdir(tempDir); err != nil {
		testCWDMutex.Unlock()
		t.Fatalf("failed to change working directory: %v", err)
	}

	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
		testCWDMutex.Unlock()
	})

	return tempDir
}
