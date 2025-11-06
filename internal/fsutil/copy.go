package fsutil

import (
	"fmt"
	"io"
	"os"
)

// CopyFileAtomic performs an atomic streaming copy from src to dst.
// It writes to a temporary file first, then renames it to the destination.
// This ensures the destination file is never in a partially written state.
//
// Features:
//   - Streaming copy (memory-safe for large files)
//   - Atomic rename (most filesystems)
//   - Automatic cleanup of temp files on error
//
// Returns an error if any operation fails (open, copy, close, rename).
func CopyFileAtomic(src, dst string) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Write to temporary file first for atomic operation
	tmpDst := dst + ".tmp"
	dstFile, err := os.Create(tmpDst)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	// Stream copy (memory-safe for large files)
	_, err = io.Copy(dstFile, srcFile)
	closeErr := dstFile.Close()

	if err != nil {
		os.Remove(tmpDst) // Clean up temp file on copy error
		return fmt.Errorf("failed to copy data: %w", err)
	}

	if closeErr != nil {
		os.Remove(tmpDst) // Clean up temp file on close error
		return fmt.Errorf("failed to close temp file: %w", closeErr)
	}

	// Rename temp file to final destination (atomic on most filesystems)
	if err := os.Rename(tmpDst, dst); err != nil {
		os.Remove(tmpDst) // Clean up temp file on rename error
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}
