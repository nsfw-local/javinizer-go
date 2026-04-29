package fsutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestCopyFileFs_Basic(t *testing.T) {
	fs := afero.NewMemMapFs()

	err := afero.WriteFile(fs, "/source.txt", []byte("copy fs test content"), 0644)
	assert.NoError(t, err)

	err = CopyFileFs(fs, "/source.txt", "/dest.txt")
	assert.NoError(t, err)

	content, err := afero.ReadFile(fs, "/dest.txt")
	assert.NoError(t, err)
	assert.Equal(t, []byte("copy fs test content"), content)
}

func TestCopyFileFs_SamePath(t *testing.T) {
	fs := afero.NewMemMapFs()

	err := afero.WriteFile(fs, "/source.txt", []byte("same path"), 0644)
	assert.NoError(t, err)

	err = CopyFileFs(fs, "/source.txt", "/source.txt")
	assert.NoError(t, err)
}

func TestCopyFileFs_CreatesDir(t *testing.T) {
	fs := afero.NewMemMapFs()

	err := afero.WriteFile(fs, "/source.txt", []byte("nested"), 0644)
	assert.NoError(t, err)

	err = CopyFileFs(fs, "/source.txt", "/a/b/c/dest.txt")
	assert.NoError(t, err)

	content, err := afero.ReadFile(fs, "/a/b/c/dest.txt")
	assert.NoError(t, err)
	assert.Equal(t, []byte("nested"), content)
}

func TestCopyFileFs_SourceNotFound(t *testing.T) {
	fs := afero.NewMemMapFs()

	err := CopyFileFs(fs, "/nonexistent.txt", "/dest.txt")
	assert.Error(t, err)
}

func TestMoveFileFs_MemMapFs(t *testing.T) {
	fs := afero.NewMemMapFs()

	err := afero.WriteFile(fs, "/source.txt", []byte("move fs test"), 0644)
	assert.NoError(t, err)

	err = MoveFileFs(fs, "/source.txt", "/subdir/dest.txt")
	assert.NoError(t, err)

	content, err := afero.ReadFile(fs, "/subdir/dest.txt")
	assert.NoError(t, err)
	assert.Equal(t, []byte("move fs test"), content)
}

func TestMoveFileFs_SamePath(t *testing.T) {
	fs := afero.NewMemMapFs()

	err := afero.WriteFile(fs, "/source.txt", []byte("same"), 0644)
	assert.NoError(t, err)

	err = MoveFileFs(fs, "/source.txt", "/source.txt")
	assert.NoError(t, err)
}

func TestCrossDeviceMoveFs(t *testing.T) {
	fs := afero.NewMemMapFs()

	err := afero.WriteFile(fs, "/source.txt", []byte("cross device fs"), 0644)
	assert.NoError(t, err)

	err = crossDeviceMoveFs(fs, "/source.txt", "/dest.txt")
	assert.NoError(t, err)

	content, err := afero.ReadFile(fs, "/dest.txt")
	assert.NoError(t, err)
	assert.Equal(t, []byte("cross device fs"), content)

	_, err = fs.Stat("/source.txt")
	assert.True(t, os.IsNotExist(err))
}

func TestCrossDeviceMoveFs_SourceRemoveFailure(t *testing.T) {
	fs := afero.NewMemMapFs()

	err := afero.WriteFile(fs, "/source.txt", []byte("data"), 0644)
	assert.NoError(t, err)

	readonlyFs := afero.NewReadOnlyFs(fs)
	err = crossDeviceMoveFs(readonlyFs, "/source.txt", "/dest.txt")
	assert.Error(t, err)
}

func TestCopyFileData_Basic(t *testing.T) {
	tmpDir := t.TempDir()

	srcPath := filepath.Join(tmpDir, "source.txt")
	err := os.WriteFile(srcPath, []byte("copy data test"), 0644)
	assert.NoError(t, err)

	dstPath := filepath.Join(tmpDir, "dest.txt")
	err = copyFileData(srcPath, dstPath)
	assert.NoError(t, err)

	content, err := os.ReadFile(dstPath)
	assert.NoError(t, err)
	assert.Equal(t, []byte("copy data test"), content)
}

func TestCopyFileData_SourceNotFound(t *testing.T) {
	tmpDir := t.TempDir()

	err := copyFileData(filepath.Join(tmpDir, "nonexistent.txt"), filepath.Join(tmpDir, "dest.txt"))
	assert.Error(t, err)
}

func TestCopyFileDataFs_BasicMemMap(t *testing.T) {
	fs := afero.NewMemMapFs()

	err := afero.WriteFile(fs, "/source.txt", []byte("memmap copy"), 0644)
	assert.NoError(t, err)

	err = copyFileDataFs(fs, "/source.txt", "/dest.txt")
	assert.NoError(t, err)

	content, err := afero.ReadFile(fs, "/dest.txt")
	assert.NoError(t, err)
	assert.Equal(t, []byte("memmap copy"), content)
}

func TestCopyFileDataFs_SourceNotFound(t *testing.T) {
	fs := afero.NewMemMapFs()

	err := copyFileDataFs(fs, "/nonexistent.txt", "/dest.txt")
	assert.Error(t, err)
}

func TestMoveFile_SamePath(t *testing.T) {
	tmpDir := t.TempDir()

	srcPath := filepath.Join(tmpDir, "source.txt")
	err := os.WriteFile(srcPath, []byte("same path"), 0644)
	assert.NoError(t, err)

	err = MoveFile(srcPath, srcPath)
	assert.NoError(t, err)
}
