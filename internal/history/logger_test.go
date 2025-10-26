package history

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/javinizer/javinizer-go/internal/config"
	"github.com/javinizer/javinizer-go/internal/database"
	"github.com/javinizer/javinizer-go/internal/models"
)

func setupTestDB(t *testing.T) (*database.DB, func()) {
	t.Helper()

	// Create temp directory for test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type: "sqlite",
			DSN:  dbPath,
		},
	}

	db, err := database.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	if err := db.AutoMigrate(); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.RemoveAll(tmpDir)
	}

	return db, cleanup
}

func TestNewLogger(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	logger := NewLogger(db)
	if logger == nil {
		t.Fatal("NewLogger returned nil")
	}

	if logger.repo == nil {
		t.Fatal("Logger repository is nil")
	}
}

func TestLogOrganize_Success(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	logger := NewLogger(db)

	err := logger.LogOrganize("IPX-535", "/old/path.mp4", "/new/path.mp4", false, nil)
	if err != nil {
		t.Fatalf("LogOrganize failed: %v", err)
	}

	// Verify the record was created
	records, err := logger.GetByMovieID("IPX-535")
	if err != nil {
		t.Fatalf("GetByMovieID failed: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(records))
	}

	record := records[0]
	if record.MovieID != "IPX-535" {
		t.Errorf("Expected MovieID 'IPX-535', got '%s'", record.MovieID)
	}
	if record.Operation != "organize" {
		t.Errorf("Expected operation 'organize', got '%s'", record.Operation)
	}
	if record.OriginalPath != "/old/path.mp4" {
		t.Errorf("Expected OriginalPath '/old/path.mp4', got '%s'", record.OriginalPath)
	}
	if record.NewPath != "/new/path.mp4" {
		t.Errorf("Expected NewPath '/new/path.mp4', got '%s'", record.NewPath)
	}
	if record.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", record.Status)
	}
	if record.DryRun != false {
		t.Error("Expected DryRun to be false")
	}
}

func TestLogOrganize_Failed(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	logger := NewLogger(db)

	testErr := &os.PathError{Op: "move", Path: "/test/path", Err: os.ErrNotExist}
	err := logger.LogOrganize("IPX-535", "/old/path.mp4", "/new/path.mp4", false, testErr)
	if err != nil {
		t.Fatalf("LogOrganize failed: %v", err)
	}

	records, err := logger.GetByMovieID("IPX-535")
	if err != nil {
		t.Fatalf("GetByMovieID failed: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(records))
	}

	record := records[0]
	if record.Status != "failed" {
		t.Errorf("Expected status 'failed', got '%s'", record.Status)
	}
	if record.ErrorMessage == "" {
		t.Error("Expected error message to be set")
	}
}

func TestLogOrganize_DryRun(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	logger := NewLogger(db)

	err := logger.LogOrganize("IPX-535", "/old/path.mp4", "/new/path.mp4", true, nil)
	if err != nil {
		t.Fatalf("LogOrganize failed: %v", err)
	}

	records, err := logger.GetByMovieID("IPX-535")
	if err != nil {
		t.Fatalf("GetByMovieID failed: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(records))
	}

	record := records[0]
	if record.DryRun != true {
		t.Error("Expected DryRun to be true")
	}
}

func TestLogScrape(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	logger := NewLogger(db)

	metadata := map[string]string{
		"title":  "Test Movie",
		"studio": "Test Studio",
	}

	err := logger.LogScrape("IPX-535", "https://r18.dev/videos/vod/movies/detail/-/id=ipx00535", metadata, nil)
	if err != nil {
		t.Fatalf("LogScrape failed: %v", err)
	}

	records, err := logger.GetByMovieID("IPX-535")
	if err != nil {
		t.Fatalf("GetByMovieID failed: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(records))
	}

	record := records[0]
	if record.Operation != "scrape" {
		t.Errorf("Expected operation 'scrape', got '%s'", record.Operation)
	}
	if record.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", record.Status)
	}
	if record.Metadata == "" {
		t.Error("Expected metadata to be set")
	}
}

func TestLogDownload(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	logger := NewLogger(db)

	err := logger.LogDownload("IPX-535", "https://example.com/cover.jpg", "/local/cover.jpg", "cover", nil)
	if err != nil {
		t.Fatalf("LogDownload failed: %v", err)
	}

	records, err := logger.GetByMovieID("IPX-535")
	if err != nil {
		t.Fatalf("GetByMovieID failed: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(records))
	}

	record := records[0]
	if record.Operation != "download" {
		t.Errorf("Expected operation 'download', got '%s'", record.Operation)
	}
	if record.OriginalPath != "https://example.com/cover.jpg" {
		t.Errorf("Expected OriginalPath to be URL, got '%s'", record.OriginalPath)
	}
	if record.NewPath != "/local/cover.jpg" {
		t.Errorf("Expected NewPath '/local/cover.jpg', got '%s'", record.NewPath)
	}
}

func TestLogNFO(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	logger := NewLogger(db)

	err := logger.LogNFO("IPX-535", "/path/to/IPX-535.nfo", nil)
	if err != nil {
		t.Fatalf("LogNFO failed: %v", err)
	}

	records, err := logger.GetByMovieID("IPX-535")
	if err != nil {
		t.Fatalf("GetByMovieID failed: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(records))
	}

	record := records[0]
	if record.Operation != "nfo" {
		t.Errorf("Expected operation 'nfo', got '%s'", record.Operation)
	}
	if record.NewPath != "/path/to/IPX-535.nfo" {
		t.Errorf("Expected NewPath '/path/to/IPX-535.nfo', got '%s'", record.NewPath)
	}
}

func TestLogRevert(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	logger := NewLogger(db)

	err := logger.LogRevert("IPX-535", "/original/path.mp4", "/organized/path.mp4", nil)
	if err != nil {
		t.Fatalf("LogRevert failed: %v", err)
	}

	records, err := logger.GetByMovieID("IPX-535")
	if err != nil {
		t.Fatalf("GetByMovieID failed: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(records))
	}

	record := records[0]
	if record.Status != "reverted" {
		t.Errorf("Expected status 'reverted', got '%s'", record.Status)
	}
	if record.Operation != "organize" {
		t.Errorf("Expected operation 'organize', got '%s'", record.Operation)
	}
}

func TestGetRecent(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	logger := NewLogger(db)

	// Create multiple records
	for i := 0; i < 5; i++ {
		err := logger.LogOrganize("IPX-535", "/old/path.mp4", "/new/path.mp4", false, nil)
		if err != nil {
			t.Fatalf("LogOrganize failed: %v", err)
		}
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	records, err := logger.GetRecent(3)
	if err != nil {
		t.Fatalf("GetRecent failed: %v", err)
	}

	if len(records) != 3 {
		t.Errorf("Expected 3 records, got %d", len(records))
	}

	// Verify they are in reverse chronological order
	for i := 0; i < len(records)-1; i++ {
		if records[i].CreatedAt.Before(records[i+1].CreatedAt) {
			t.Error("Records are not in reverse chronological order")
		}
	}
}

func TestGetByOperation(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	logger := NewLogger(db)

	// Create different operation types
	logger.LogOrganize("IPX-535", "/old/path.mp4", "/new/path.mp4", false, nil)
	logger.LogScrape("IPX-535", "https://example.com", nil, nil)
	logger.LogDownload("IPX-535", "https://example.com/cover.jpg", "/local/cover.jpg", "cover", nil)
	logger.LogNFO("IPX-535", "/path/to/nfo", nil)

	// Get only scrape operations
	scrapeRecords, err := logger.GetByOperation("scrape", 10)
	if err != nil {
		t.Fatalf("GetByOperation failed: %v", err)
	}

	if len(scrapeRecords) != 1 {
		t.Errorf("Expected 1 scrape record, got %d", len(scrapeRecords))
	}

	if scrapeRecords[0].Operation != "scrape" {
		t.Errorf("Expected operation 'scrape', got '%s'", scrapeRecords[0].Operation)
	}

	// Get only organize operations
	organizeRecords, err := logger.GetByOperation("organize", 10)
	if err != nil {
		t.Fatalf("GetByOperation failed: %v", err)
	}

	if len(organizeRecords) != 1 {
		t.Errorf("Expected 1 organize record, got %d", len(organizeRecords))
	}
}

func TestGetByStatus(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	logger := NewLogger(db)

	// Create records with different statuses
	logger.LogOrganize("IPX-535", "/old/path.mp4", "/new/path.mp4", false, nil)
	logger.LogOrganize("IPX-536", "/old/path.mp4", "/new/path.mp4", false, os.ErrNotExist)

	// Get only successful operations
	successRecords, err := logger.GetByStatus("success", 10)
	if err != nil {
		t.Fatalf("GetByStatus failed: %v", err)
	}

	if len(successRecords) != 1 {
		t.Errorf("Expected 1 success record, got %d", len(successRecords))
	}

	// Get only failed operations
	failedRecords, err := logger.GetByStatus("failed", 10)
	if err != nil {
		t.Fatalf("GetByStatus failed: %v", err)
	}

	if len(failedRecords) != 1 {
		t.Errorf("Expected 1 failed record, got %d", len(failedRecords))
	}
}

func TestGetStats(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	logger := NewLogger(db)

	// Create various records
	logger.LogOrganize("IPX-535", "/old/path.mp4", "/new/path.mp4", false, nil)
	logger.LogOrganize("IPX-536", "/old/path.mp4", "/new/path.mp4", false, os.ErrNotExist)
	logger.LogScrape("IPX-535", "https://example.com", nil, nil)
	logger.LogDownload("IPX-535", "https://example.com/cover.jpg", "/local/cover.jpg", "cover", nil)
	logger.LogNFO("IPX-535", "/path/to/nfo", nil)
	logger.LogRevert("IPX-535", "/original/path.mp4", "/organized/path.mp4", nil)

	stats, err := logger.GetStats()
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}

	if stats.Total != 6 {
		t.Errorf("Expected Total 6, got %d", stats.Total)
	}

	if stats.Success != 4 {
		t.Errorf("Expected Success 4, got %d", stats.Success)
	}

	if stats.Failed != 1 {
		t.Errorf("Expected Failed 1, got %d", stats.Failed)
	}

	if stats.Reverted != 1 {
		t.Errorf("Expected Reverted 1, got %d", stats.Reverted)
	}

	if stats.Scrape != 1 {
		t.Errorf("Expected Scrape 1, got %d", stats.Scrape)
	}

	if stats.Organize != 3 {
		t.Errorf("Expected Organize 3 (2 organize + 1 revert), got %d", stats.Organize)
	}

	if stats.Download != 1 {
		t.Errorf("Expected Download 1, got %d", stats.Download)
	}

	if stats.NFO != 1 {
		t.Errorf("Expected NFO 1, got %d", stats.NFO)
	}
}

func TestCleanupOldRecords(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	logger := NewLogger(db)
	repo := database.NewHistoryRepository(db)

	// Create old record
	oldRecord := &models.History{
		MovieID:      "IPX-535",
		Operation:    "organize",
		OriginalPath: "/old/path.mp4",
		NewPath:      "/new/path.mp4",
		Status:       "success",
		CreatedAt:    time.Now().Add(-60 * 24 * time.Hour), // 60 days ago
	}
	repo.Create(oldRecord)

	// Create recent record
	logger.LogOrganize("IPX-536", "/old/path.mp4", "/new/path.mp4", false, nil)

	// Verify 2 records exist
	allRecords, _ := logger.GetRecent(10)
	if len(allRecords) != 2 {
		t.Fatalf("Expected 2 records before cleanup, got %d", len(allRecords))
	}

	// Cleanup records older than 30 days
	err := logger.CleanupOldRecords(30 * 24 * time.Hour)
	if err != nil {
		t.Fatalf("CleanupOldRecords failed: %v", err)
	}

	// Verify only 1 record remains
	remainingRecords, _ := logger.GetRecent(10)
	if len(remainingRecords) != 1 {
		t.Errorf("Expected 1 record after cleanup, got %d", len(remainingRecords))
	}

	if remainingRecords[0].MovieID != "IPX-536" {
		t.Errorf("Expected remaining record to be IPX-536, got %s", remainingRecords[0].MovieID)
	}
}

func TestMultipleMovies(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	logger := NewLogger(db)

	// Create records for multiple movies
	logger.LogOrganize("IPX-535", "/old/path1.mp4", "/new/path1.mp4", false, nil)
	logger.LogOrganize("IPX-536", "/old/path2.mp4", "/new/path2.mp4", false, nil)
	logger.LogOrganize("IPX-535", "/old/path3.mp4", "/new/path3.mp4", false, nil)

	// Get records for specific movie
	ipx535Records, err := logger.GetByMovieID("IPX-535")
	if err != nil {
		t.Fatalf("GetByMovieID failed: %v", err)
	}

	if len(ipx535Records) != 2 {
		t.Errorf("Expected 2 records for IPX-535, got %d", len(ipx535Records))
	}

	ipx536Records, err := logger.GetByMovieID("IPX-536")
	if err != nil {
		t.Fatalf("GetByMovieID failed: %v", err)
	}

	if len(ipx536Records) != 1 {
		t.Errorf("Expected 1 record for IPX-536, got %d", len(ipx536Records))
	}
}
