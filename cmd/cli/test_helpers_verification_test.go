package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/javinizer/javinizer-go/internal/config"
	"github.com/javinizer/javinizer-go/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSetupTestDB verifies the database helper works correctly
func TestSetupTestDB(t *testing.T) {
	db := setupTestDB(t)
	require.NotNil(t, db)

	// Verify we can query the database
	var count int64
	err := db.Model(&models.Movie{}).Count(&count).Error
	require.NoError(t, err)
	assert.Equal(t, int64(0), count, "Expected empty database")
}

// TestSetupTestDB_WithSpecialCharacters verifies URL escaping works for subtests
func TestSetupTestDB_WithSpecialCharacters(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"test?with=query"},
		{"test#with-hash"},
		{"test/with/slashes"},
		{"test with spaces"},
		{"test&with&ampersands"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic or fail due to URI issues
			db := setupTestDB(t)
			require.NotNil(t, db)

			// Verify database is functional
			movie := &models.Movie{
				ContentID: "TEST-001",
				Title:     "Test",
			}
			err := db.Create(movie).Error
			require.NoError(t, err)
		})
	}
}

// TestMockScraperRegistry verifies the mock scraper registry works
func TestMockScraperRegistry(t *testing.T) {
	results := map[string]*models.ScraperResult{
		"r18dev": {
			Source: "r18dev",
			Title:  "Test Movie",
			ID:     "IPX-535",
		},
		"dmm": {
			Source: "dmm",
			Title:  "テストムービー",
			ID:     "IPX-535",
		},
	}

	registry := setupMockScraperRegistry(t, results)
	require.NotNil(t, registry)

	// Verify we can get scrapers
	r18, exists := registry.Get("r18dev")
	require.True(t, exists)
	require.NotNil(t, r18)
	assert.Equal(t, "r18dev", r18.Name())

	// Verify scraper returns expected results
	result, err := r18.Search("IPX-535")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "Test Movie", result.Title)
	assert.Equal(t, "r18dev", result.Source)
}

// TestMockScraperRegistryWithError verifies error handling
func TestMockScraperRegistryWithError(t *testing.T) {
	registry := setupMockScraperRegistryWithError(t, "r18dev", assert.AnError)
	require.NotNil(t, registry)

	scraper, exists := registry.Get("r18dev")
	require.True(t, exists)

	// Should return error
	result, err := scraper.Search("IPX-535")
	require.Error(t, err)
	assert.Nil(t, result)
}

// TestCreateTestConfig verifies config creation
func TestCreateTestConfig(t *testing.T) {
	configPath, cfg := createTestConfig(t,
		WithScraperPriority([]string{"r18dev"}),
		WithOutputFolder("<ID> - <TITLE>"),
		WithNFOEnabled(true),
	)

	require.NotNil(t, cfg)
	assertFileExists(t, configPath)

	// Verify config values
	assert.Equal(t, []string{"r18dev"}, cfg.Scrapers.Priority)
	assert.Equal(t, "<ID> - <TITLE>", cfg.Output.FolderFormat)
	assert.True(t, cfg.Metadata.NFO.Enabled)
}

// TestCreateTestVideoFile verifies file creation
func TestCreateTestVideoFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := createTestVideoFile(t, tmpDir, "IPX-535.mp4")

	assertFileExists(t, path)

	// Verify file content
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "dummy video content", string(content))
}

// TestCreateTestMovieFiles verifies multiple file creation
func TestCreateTestMovieFiles(t *testing.T) {
	tmpDir := t.TempDir()
	paths := createTestMovieFiles(t, tmpDir)

	require.Len(t, paths, 8, "Expected 8 test files")

	// Verify all files exist
	for _, path := range paths {
		assertFileExists(t, path)
	}
}

// TestCaptureOutput verifies output capturing
func TestCaptureOutput(t *testing.T) {
	stdout, stderr := captureOutput(t, func() {
		os.Stdout.WriteString("Hello stdout\n")
		os.Stderr.WriteString("Hello stderr\n")
	})

	assert.Contains(t, stdout, "Hello stdout")
	assert.Contains(t, stderr, "Hello stderr")
}

// TestAssertMovieInDB verifies database assertions
func TestAssertMovieInDB(t *testing.T) {
	db := setupTestDB(t)
	movieID := insertTestMovie(t, db, "IPX-535", "Test Movie")

	assert.Equal(t, "IPX-535", movieID)
	assertMovieInDB(t, db, "IPX-535")
}

// TestAssertMovieNotInDB verifies negative database assertions
func TestAssertMovieNotInDB(t *testing.T) {
	db := setupTestDB(t)

	assertMovieNotInDB(t, db, "NONEXISTENT-001")
}

// TestCountMoviesInDB verifies counting helper
func TestCountMoviesInDB(t *testing.T) {
	db := setupTestDB(t)

	// Initially empty
	count := countMoviesInDB(t, db)
	assert.Equal(t, 0, count)

	// Insert movies
	insertTestMovie(t, db, "IPX-535", "Movie 1")
	insertTestMovie(t, db, "ABC-123", "Movie 2")

	// Verify count
	count = countMoviesInDB(t, db)
	assert.Equal(t, 2, count)
}

// TestSetupTestDatabaseWithData verifies bulk data creation
func TestSetupTestDatabaseWithData(t *testing.T) {
	db, movieIDs := setupTestDatabaseWithData(t, 5)

	require.Len(t, movieIDs, 5)

	// Verify all movies exist
	for _, id := range movieIDs {
		assertMovieInDB(t, db, id)
	}

	// Verify count
	count := countMoviesInDB(t, db)
	assert.Equal(t, 5, count)
}

// TestCreateTestDirectoryStructure verifies directory creation
func TestCreateTestDirectoryStructure(t *testing.T) {
	rootDir, files := createTestDirectoryStructure(t)

	require.NotEmpty(t, rootDir)
	require.Len(t, files, 3)

	// Verify all files exist
	for name, path := range files {
		assertFileExists(t, path)
		t.Logf("Created %s at %s", name, path)
	}
}

// TestCreateTestMovie verifies movie creation helper
func TestCreateTestMovie(t *testing.T) {
	movie := createTestMovie("IPX-535", "Test Movie")

	require.NotNil(t, movie)
	assert.Equal(t, "IPX-535", movie.ID)
	assert.Equal(t, "Test Movie", movie.Title)
	assert.Equal(t, 120, movie.Runtime)
	assert.Len(t, movie.Actresses, 1)
	assert.Len(t, movie.Genres, 2)
}

// TestAssertFileHelpers verifies file assertion helpers
func TestAssertFileHelpers(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file
	existingFile := createTestVideoFile(t, tmpDir, "exists.mp4")

	// Test file exists assertion
	assertFileExists(t, existingFile)

	// Test file not exists assertion
	nonExistentFile := tmpDir + "/nonexistent.mp4"
	assertFileNotExists(t, nonExistentFile)
}

// TestCountTagsForMovie verifies tag counting
func TestCountTagsForMovie(t *testing.T) {
	db := setupTestDB(t)

	// Insert a movie
	movieID := insertTestMovie(t, db, "IPX-535", "Test Movie")

	// Initially no tags
	count := countTagsForMovie(t, db, movieID)
	assert.Equal(t, 0, count)

	// Add some tags
	err := db.Create(&models.MovieTag{MovieID: movieID, Tag: "favorite"}).Error
	require.NoError(t, err)
	err = db.Create(&models.MovieTag{MovieID: movieID, Tag: "watched"}).Error
	require.NoError(t, err)

	// Verify count
	count = countTagsForMovie(t, db, movieID)
	assert.Equal(t, 2, count)
}

func TestWithTempConfig(t *testing.T) {
	// Create a test config
	tmpDir := t.TempDir()
	testCfg := config.DefaultConfig()
	testCfg.Database.DSN = filepath.Join(tmpDir, "test.db")

	// Save original global config
	originalCfg := cfg

	// Verify config is temporarily replaced
	withTempConfig(t, testCfg, func() {
		// Inside the function, cfg should be the test config
		assert.Equal(t, filepath.Join(tmpDir, "test.db"), cfg.Database.DSN)
	})

	// After function returns, original config should be restored
	assert.Equal(t, originalCfg, cfg)
}

func TestWithTempConfig_RestoresOnFailure(t *testing.T) {
	// Create a test config
	tmpDir := t.TempDir()
	testCfg := config.DefaultConfig()
	testCfg.Database.DSN = filepath.Join(tmpDir, "test.db")

	// Save original global config
	originalCfg := cfg

	// Verify config is restored even when function panics
	func() {
		defer func() {
			recover() // Catch the panic
		}()
		withTempConfig(t, testCfg, func() {
			panic("test panic")
		})
	}()

	// Original config should be restored even after panic
	assert.Equal(t, originalCfg, cfg)
}

func TestWithTempConfigFile(t *testing.T) {
	// Create a test config file
	_, testCfg := createTestConfig(t, WithOutputFolder("test-output"))

	// Save to file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yaml")
	err := config.Save(testCfg, configPath)
	require.NoError(t, err)

	// Save original global config file path
	originalCfgFile := cfgFile

	// Verify config file path is temporarily replaced
	withTempConfigFile(t, configPath, func() {
		// Inside the function, cfgFile should be the test config path
		assert.Equal(t, configPath, cfgFile)
	})

	// After function returns, original config file path should be restored
	assert.Equal(t, originalCfgFile, cfgFile)
}

func TestWithTempConfigFile_RestoresOnFailure(t *testing.T) {
	// Create a test config file
	_, testCfg := createTestConfig(t, WithOutputFolder("test-output"))

	// Save to file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yaml")
	err := config.Save(testCfg, configPath)
	require.NoError(t, err)

	// Save original global config file path
	originalCfgFile := cfgFile

	// Verify config file path is restored even when function panics
	func() {
		defer func() {
			recover() // Catch the panic
		}()
		withTempConfigFile(t, configPath, func() {
			panic("test panic")
		})
	}()

	// Original config file path should be restored even after panic
	assert.Equal(t, originalCfgFile, cfgFile)
}
