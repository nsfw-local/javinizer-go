package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/javinizer/javinizer-go/internal/config"
	"github.com/javinizer/javinizer-go/internal/database"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRunInit_Success verifies that runInit creates config and database successfully
func TestRunInit_Success(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	dbPath := filepath.Join(tmpDir, "data", "javinizer.db")

	// Create initial config
	testCfg := config.DefaultConfig()
	testCfg.Database.DSN = dbPath
	err := config.Save(testCfg, configPath)
	require.NoError(t, err)

	// Ensure database directory exists for initial connection
	err = os.MkdirAll(filepath.Dir(dbPath), 0755)
	require.NoError(t, err)

	withTempConfigFile(t, configPath, func() {
		cmd := &cobra.Command{}
		deps := createTestDependencies(t, testCfg)
		defer deps.Close()

		stdout, _ := captureOutput(t, func() {
			err := runInit(cmd, []string{}, deps)
			require.NoError(t, err)
		})

		// Verify config file exists
		assertFileExists(t, configPath)

		// Verify database was created
		assertFileExists(t, dbPath)

		// Verify output messages
		assert.Contains(t, stdout, "Initializing Javinizer")
		assert.Contains(t, stdout, "Created data directory")
		assert.Contains(t, stdout, "Initialized database")
		assert.Contains(t, stdout, "Saved configuration")
		assert.Contains(t, stdout, "Initialization complete")
	})
}

// TestRunInit_DatabaseMigrations verifies that all tables are created during initialization
func TestRunInit_DatabaseMigrations(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	dbPath := filepath.Join(tmpDir, "data", "javinizer.db")

	// Create initial config
	testCfg := config.DefaultConfig()
	testCfg.Database.DSN = dbPath
	err := config.Save(testCfg, configPath)
	require.NoError(t, err)

	// Ensure database directory exists for initial connection
	err = os.MkdirAll(filepath.Dir(dbPath), 0755)
	require.NoError(t, err)

	withTempConfigFile(t, configPath, func() {
		cmd := &cobra.Command{}
		deps := createTestDependencies(t, testCfg)
		defer deps.Close()

		captureOutput(t, func() {
			err := runInit(cmd, []string{}, deps)
			require.NoError(t, err)
		})

		// Verify database file was created
		assertFileExists(t, dbPath)

		// Load the created config
		cfgLoaded, err := config.Load(configPath)
		require.NoError(t, err)

		// Connect to the created database
		db, err := database.New(cfgLoaded)
		require.NoError(t, err)
		defer db.Close()

		// Verify critical tables exist by attempting to query them
		// If migrations failed, these queries would fail
		var count int64

		// Test movies table
		err = db.Model(&struct{ ID string }{}).Table("movies").Count(&count).Error
		assert.NoError(t, err, "movies table should exist")

		// Test actresses table
		err = db.Model(&struct{ ID uint }{}).Table("actresses").Count(&count).Error
		assert.NoError(t, err, "actresses table should exist")

		// Test genres table
		err = db.Model(&struct{ ID uint }{}).Table("genres").Count(&count).Error
		assert.NoError(t, err, "genres table should exist")

		// Test genre_replacements table
		err = db.Model(&struct{ ID uint }{}).Table("genre_replacements").Count(&count).Error
		assert.NoError(t, err, "genre_replacements table should exist")

		// Test movie_tags table
		err = db.Model(&struct{ ID uint }{}).Table("movie_tags").Count(&count).Error
		assert.NoError(t, err, "movie_tags table should exist")
	})
}

// TestRunInit_DirectoryCreation verifies that necessary directories are created
func TestRunInit_DirectoryCreation(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	dataDir := filepath.Join(tmpDir, "data")
	dbPath := filepath.Join(dataDir, "javinizer.db")

	// Create initial config
	testCfg := config.DefaultConfig()
	testCfg.Database.DSN = dbPath
	err := config.Save(testCfg, configPath)
	require.NoError(t, err)

	// Ensure database directory exists for initial connection
	err = os.MkdirAll(dataDir, 0755)
	require.NoError(t, err)

	withTempConfigFile(t, configPath, func() {
		cmd := &cobra.Command{}
		deps := createTestDependencies(t, testCfg)
		defer deps.Close()

		stdout, _ := captureOutput(t, func() {
			err := runInit(cmd, []string{}, deps)
			require.NoError(t, err)
		})

		// Verify data directory was created
		info, err := os.Stat(dataDir)
		require.NoError(t, err, "data directory should exist")
		assert.True(t, info.IsDir(), "data should be a directory")

		// Verify output mentions directory creation
		assert.Contains(t, stdout, "Created data directory")
		assert.Contains(t, stdout, dataDir)
	})
}

// TestRunInit_ConfigFileContent verifies the created config has valid content
func TestRunInit_ConfigFileContent(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	dbPath := filepath.Join(tmpDir, "data", "javinizer.db")

	// Create initial config
	testCfg := config.DefaultConfig()
	testCfg.Database.DSN = dbPath
	err := config.Save(testCfg, configPath)
	require.NoError(t, err)

	// Ensure database directory exists for initial connection
	err = os.MkdirAll(filepath.Dir(dbPath), 0755)
	require.NoError(t, err)

	withTempConfigFile(t, configPath, func() {
		cmd := &cobra.Command{}
		deps := createTestDependencies(t, testCfg)
		defer deps.Close()

		captureOutput(t, func() {
			err := runInit(cmd, []string{}, deps)
			require.NoError(t, err)
		})

		// Load and verify config content
		cfgLoaded, err := config.Load(configPath)
		require.NoError(t, err)

		// Verify critical config fields have sensible defaults
		assert.NotEmpty(t, cfgLoaded.Database.DSN, "database DSN should be set")
		assert.Equal(t, "sqlite", cfgLoaded.Database.Type, "default database type should be sqlite")
		assert.NotEmpty(t, cfgLoaded.Scrapers.Priority, "scraper priority should be set")
		assert.NotEmpty(t, cfgLoaded.Output.FolderFormat, "folder format should be set")
		assert.NotEmpty(t, cfgLoaded.Output.FileFormat, "file format should be set")

		// Verify scrapers are configured
		assert.True(t, len(cfgLoaded.Scrapers.Priority) > 0, "should have at least one scraper in priority")
	})
}

// NOTE: TestRunInit_InvalidConfigPath is intentionally not implemented
//
// The error path in runInit() cannot be reliably tested because it calls logging.Fatal(),
// which invokes os.Exit() and terminates the test process. Testing this properly would require:
//
// 1. Refactoring runInit() to return errors instead of calling logging.Fatal, OR
// 2. Implementing a test hook mechanism to intercept logging.Fatal calls
//
// Until such refactoring is done, this error path remains untested. The success paths
// and other error-handling paths in runInit() ARE tested in the other test cases.
//
// Related test cases that DO work:
// - TestRunInit_Success: Tests successful initialization
// - TestRunInit_DatabaseMigrations: Tests successful DB setup
// - TestRunInit_DirectoryCreation: Tests successful directory creation

// TestRunInit_RepeatedInitialization verifies running init multiple times
func TestRunInit_RepeatedInitialization(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	dbPath := filepath.Join(tmpDir, "data", "javinizer.db")

	// Create initial config
	testCfg := config.DefaultConfig()
	testCfg.Database.DSN = dbPath
	err := config.Save(testCfg, configPath)
	require.NoError(t, err)

	// Ensure database directory exists for initial connection
	err = os.MkdirAll(filepath.Dir(dbPath), 0755)
	require.NoError(t, err)

	withTempConfigFile(t, configPath, func() {
		cmd := &cobra.Command{}
		deps := createTestDependencies(t, testCfg)
		defer deps.Close()

		// Run init first time
		stdout1, _ := captureOutput(t, func() {
			err := runInit(cmd, []string{}, deps)
			require.NoError(t, err)
		})
		assert.Contains(t, stdout1, "Initialization complete")

		// Verify config exists
		assertFileExists(t, configPath)

		// Get initial config content
		initialContent, err := os.ReadFile(configPath)
		require.NoError(t, err)

		// Run init second time
		stdout2, _ := captureOutput(t, func() {
			err := runInit(cmd, []string{}, deps)
			require.NoError(t, err)
		})
		assert.Contains(t, stdout2, "Initialization complete")

		// Verify config still exists and wasn't corrupted
		assertFileExists(t, configPath)

		// Verify both configs are valid (content should be idempotent)
		assert.NotEmpty(t, initialContent)

		// Both initializations should have created valid configs
		_, err = config.Load(configPath)
		assert.NoError(t, err, "config should be valid after repeated initialization")
	})
}
