package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/javinizer/javinizer-go/internal/config"
	"github.com/javinizer/javinizer-go/internal/database"
	"github.com/javinizer/javinizer-go/internal/models"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupScrapeTestDB creates a test database with proper structure
func setupScrapeTestDB(t *testing.T) (configPath string, testCfg *config.Config) {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "data", "test.db")

	// Ensure database directory exists
	err := os.MkdirAll(filepath.Dir(dbPath), 0755)
	require.NoError(t, err)

	configPath, testCfg = createTestConfig(t,
		WithDatabaseDSN(dbPath),
		WithScraperPriority([]string{"r18dev", "dmm"}),
		WithDownloadCover(false),
	)

	// Initialize database with migrations
	db, err := database.New(testCfg)
	require.NoError(t, err)
	err = db.AutoMigrate()
	require.NoError(t, err)
	db.Close()

	return configPath, testCfg
}

// TestRunScrape_CachedMovie tests that cached movies are retrieved from database
func TestRunScrape_CachedMovie(t *testing.T) {
	configPath, testCfg := setupScrapeTestDB(t)

	withTempConfigFile(t, configPath, func() {
		deps := createTestDependencies(t, testCfg)
		defer deps.Close()

		// Pre-populate database with cached movie
		repo := database.NewMovieRepository(deps.DB)
		cachedMovie := createTestMovie("IPX-535", "Cached Title")
		cachedMovie.ContentID = "ipx00535"
		cachedMovie.Runtime = 120
		err := repo.Upsert(cachedMovie)
		require.NoError(t, err)

		// Run scrape command
		cmd := &cobra.Command{}
		cmd.Flags().Bool("force", false, "force flag")

		stdout, _ := captureOutput(t, func() {
			resetLoggerForTests(t) // Reset logger INSIDE captureOutput so it uses the pipe
			err := runScrape(cmd, []string{"IPX-535"}, deps)
			require.NoError(t, err)
		})

		// Verify cache hit
		assert.Contains(t, stdout, "Found in cache", "Should find movie in cache")
		assert.Contains(t, stdout, "Cached Title", "Should display cached title")
		assert.Contains(t, stdout, "120 min", "Should display cached runtime")

		// Verify database still has the cached movie
		movie, err := repo.FindByID("IPX-535")
		require.NoError(t, err)
		assert.Equal(t, "Cached Title", movie.Title)
	})
}

// TestRunScrape_CachedWithTranslations tests cached movie with multiple translations
func TestRunScrape_CachedWithTranslations(t *testing.T) {
	configPath, testCfg := setupScrapeTestDB(t)

	withTempConfigFile(t, configPath, func() {
		deps := createTestDependencies(t, testCfg)
		defer deps.Close()

		// Pre-populate with movie having multiple translations
		repo := database.NewMovieRepository(deps.DB)
		movie := createTestMovie("ABC-123", "English Title")
		movie.Translations = []models.MovieTranslation{
			{
				Language:   "en",
				Title:      "English Title",
				SourceName: "r18dev",
			},
			{
				Language:   "ja",
				Title:      "日本語タイトル",
				SourceName: "dmm",
			},
		}
		err := repo.Upsert(movie)
		require.NoError(t, err)

		cmd := &cobra.Command{}
		cmd.Flags().Bool("force", false, "force flag")

		stdout, _ := captureOutput(t, func() {
			resetLoggerForTests(t) // Reset logger INSIDE captureOutput so it uses the pipe
			err := runScrape(cmd, []string{"ABC-123"}, deps)
			require.NoError(t, err)
		})

		// Should show cache hit and translations
		assert.Contains(t, stdout, "Found in cache")
		assert.Contains(t, stdout, "English Title")

		// Verify translations section if multiple translations exist
		if len(movie.Translations) > 1 {
			assert.Contains(t, stdout, "Translations")
		}
	})
}

// TestRunScrape_UpdateExistingMovie tests updating cached movie with --force
func TestRunScrape_UpdateExistingMovie(t *testing.T) {
	configPath, testCfg := setupScrapeTestDB(t)

	withTempConfigFile(t, configPath, func() {
		deps := createTestDependencies(t, testCfg)
		defer deps.Close()

		// Pre-populate with old data
		repo := database.NewMovieRepository(deps.DB)
		oldMovie := createTestMovie("TESTUPDATE-001", "Old Title")
		oldMovie.Runtime = 90
		err := repo.Upsert(oldMovie)
		require.NoError(t, err)

		// First scrape without force - should use cache
		cmd1 := &cobra.Command{}
		cmd1.Flags().Bool("force", false, "force flag")

		stdout1, _ := captureOutput(t, func() {
			resetLoggerForTests(t) // Reset logger INSIDE captureOutput so it uses the pipe
			err := runScrape(cmd1, []string{"TESTUPDATE-001"}, deps)
			require.NoError(t, err)
		})

		assert.Contains(t, stdout1, "Found in cache")
		assert.Contains(t, stdout1, "Old Title")

		// Second scrape with --force flag
		cmd2 := &cobra.Command{}
		cmd2.Flags().Bool("force", false, "force flag")
		cmd2.Flags().Set("force", "true")

		stdout2, _ := captureOutput(t, func() {
			resetLoggerForTests(t) // Reset logger INSIDE captureOutput so it uses the pipe
			err := runScrape(cmd2, []string{"TESTUPDATE-001"}, deps)
			// Expect error because real scrapers will fail without network/mocks
			assert.Error(t, err, "Should error when scrapers fail to find movie")
		})

		// Should show cache cleared even though scraping failed
		assert.Contains(t, stdout2, "Cache cleared", "Force flag should clear cache")

		// Note: Since we don't have real scrapers, the scrape will fail, but cache was cleared
		// This tests the --force flag behavior
	})
}

// NOTE: TestRunScrape_DatabaseSaveError intentionally removed
//
// This test was a no-op that provided false coverage. Database save errors
// cannot be reliably tested because runScrape() calls logging.Fatal() on
// database initialization failures, which exits the process. Testing this
// properly would require:
//
// 1. Refactoring runScrape() to return errors instead of calling logging.Fatal, OR
// 2. Implementing a test hook mechanism to intercept logging.Fatal calls
//
// Until such refactoring is done, this error path remains untested.

// NOTE: TestRunScrape_SuccessfulScrapeAndSave intentionally not implemented
//
// The full success path (cache miss → scraper returns data → aggregation → save to DB)
// cannot be tested without refactoring the production code. Currently, runScrape() creates
// its own scraper registry internally with no injection point for mocks. Testing this path
// would require:
//
// 1. Adding a package-level scraper registry variable that tests can override, OR
// 2. Refactoring runScrape() to accept a scraper registry as a parameter
//
// Current tests DO cover:
// - Cache hit path (TestRunScrape_CachedMovie, TestRunScrape_CachedWithTranslations)
// - Cache updates (TestRunScrape_UpdateExistingMovie)
// - Output formatting (TestRunScrape_OutputFormatting, TestRunScrape_MultipleActresses, TestRunScrape_MediaURLsSection)
// - Cache miss behavior (TestRunScrape_NoMovieInCache)
// - Scraper priority (TestRunScrape_CustomScraperPriority)
// - Content ID resolution (TestRunScrape_ContentIDResolution)
//
// The untested scrape path is a known gap that should be addressed in future refactoring.

// TestRunScrape_OutputFormatting tests that output is properly formatted
func TestRunScrape_OutputFormatting(t *testing.T) {
	configPath, testCfg := setupScrapeTestDB(t)

	withTempConfigFile(t, configPath, func() {
		deps := createTestDependencies(t, testCfg)
		defer deps.Close()

		// Create movie with comprehensive metadata
		repo := database.NewMovieRepository(deps.DB)
		releaseDate := time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC)
		movie := &models.Movie{
			ID:          "DISPLAY-001",
			ContentID:   "display00001",
			Title:       "Display Test Movie",
			Description: "This is a comprehensive test description with multiple lines of text to verify proper formatting and wrapping behavior in the output display.",
			ReleaseDate: &releaseDate,
			Runtime:     125,
			Director:    "Famous Director",
			Maker:       "Premium Studio",
			Label:       "Exclusive Label",
			Series:      "Popular Series",
			RatingScore: 8.5,
			RatingVotes: 250,
			Actresses: []models.Actress{
				{
					DMMID:        12345,
					FirstName:    "Star",
					LastName:     "Actress",
					JapaneseName: "スター女優",
				},
			},
			Genres: []models.Genre{
				{Name: "Drama"},
				{Name: "Romance"},
				{Name: "Action"},
			},
			CoverURL:  "http://example.com/cover.jpg",
			PosterURL: "http://example.com/poster.jpg",
		}
		err := repo.Upsert(movie)
		require.NoError(t, err)

		cmd := &cobra.Command{}
		cmd.Flags().Bool("force", false, "force flag")

		stdout, _ := captureOutput(t, func() {
			err := runScrape(cmd, []string{"DISPLAY-001"}, deps)
			require.NoError(t, err)
		})

		// Verify all key fields are displayed
		assert.Contains(t, stdout, "DISPLAY-001", "Should display ID")
		assert.Contains(t, stdout, "Display Test Movie", "Should display title")
		assert.Contains(t, stdout, "2023-05-15", "Should display release date")
		assert.Contains(t, stdout, "125 min", "Should display runtime")
		assert.Contains(t, stdout, "Famous Director", "Should display director")
		assert.Contains(t, stdout, "Premium Studio", "Should display maker")
		assert.Contains(t, stdout, "Exclusive Label", "Should display label")
		assert.Contains(t, stdout, "Popular Series", "Should display series")
		assert.Contains(t, stdout, "Actress Star", "Should display actress name (LastName FirstName format)")
		assert.Contains(t, stdout, "Drama", "Should display genre")
		assert.Contains(t, stdout, "Romance", "Should display genre")

		// Verify rating if displayed
		assert.Contains(t, stdout, "8.5", "Should display rating score")

		// Verify description section
		assert.Contains(t, stdout, "Description:", "Should have description section")
		assert.Contains(t, stdout, "comprehensive test description", "Should display description")
	})
}

// TestRunScrape_MultipleActresses tests display of multiple actresses
func TestRunScrape_MultipleActresses(t *testing.T) {
	configPath, testCfg := setupScrapeTestDB(t)

	withTempConfigFile(t, configPath, func() {
		deps := createTestDependencies(t, testCfg)
		defer deps.Close()

		repo := database.NewMovieRepository(deps.DB)
		movie := createTestMovie("MULTI-001", "Multi Actress Movie")
		movie.Actresses = []models.Actress{
			{DMMID: 1, FirstName: "First", LastName: "Actress", JapaneseName: "女優一"},
			{DMMID: 2, FirstName: "Second", LastName: "Actress", JapaneseName: "女優二"},
			{DMMID: 3, FirstName: "Third", LastName: "Actress", JapaneseName: "女優三"},
			{DMMID: 4, FirstName: "Fourth", LastName: "Actress", JapaneseName: "女優四"},
			{DMMID: 5, FirstName: "Fifth", LastName: "Actress", JapaneseName: "女優五"},
		}
		err := repo.Upsert(movie)
		require.NoError(t, err)

		cmd := &cobra.Command{}
		cmd.Flags().Bool("force", false, "force flag")

		stdout, _ := captureOutput(t, func() {
			err := runScrape(cmd, []string{"MULTI-001"}, deps)
			require.NoError(t, err)
		})

		assert.Contains(t, stdout, "Actresses")
		assert.Contains(t, stdout, "Actress First") // LastName FirstName format
		assert.Contains(t, stdout, "Actress Second")
		assert.Contains(t, stdout, "Actress Third")

		// All actresses should be displayed with detailed info
		assert.Contains(t, stdout, "DMM ID: 1")
		assert.Contains(t, stdout, "DMM ID: 2")
		assert.Contains(t, stdout, "DMM ID: 3")
		assert.Contains(t, stdout, "Actress Fourth")
		assert.Contains(t, stdout, "Actress Fifth")
	})
}

// TestRunScrape_MediaURLsSection tests media URLs are displayed
func TestRunScrape_MediaURLsSection(t *testing.T) {
	configPath, testCfg := setupScrapeTestDB(t)

	withTempConfigFile(t, configPath, func() {
		deps := createTestDependencies(t, testCfg)
		defer deps.Close()

		repo := database.NewMovieRepository(deps.DB)
		movie := createTestMovie("MEDIA-001", "Media Test")
		movie.CoverURL = "http://example.com/cover.jpg"
		movie.PosterURL = "http://example.com/poster.jpg"
		movie.TrailerURL = "http://example.com/trailer.mp4"
		movie.Screenshots = []string{
			"http://example.com/screenshot1.jpg",
			"http://example.com/screenshot2.jpg",
			"http://example.com/screenshot3.jpg",
		}
		err := repo.Upsert(movie)
		require.NoError(t, err)

		cmd := &cobra.Command{}
		cmd.Flags().Bool("force", false, "force flag")

		stdout, _ := captureOutput(t, func() {
			err := runScrape(cmd, []string{"MEDIA-001"}, deps)
			require.NoError(t, err)
		})

		// Verify media URLs section
		assert.Contains(t, stdout, "Media URLs:")
		assert.Contains(t, stdout, "Cover URL")
		assert.Contains(t, stdout, "cover.jpg")
		assert.Contains(t, stdout, "Poster URL")
		assert.Contains(t, stdout, "poster.jpg")
		assert.Contains(t, stdout, "Trailer URL")
		assert.Contains(t, stdout, "trailer.mp4")
		assert.Contains(t, stdout, "Screenshots")
		assert.Contains(t, stdout, "3 total")
	})
}

// TestRunScrape_NoMovieInCache tests behavior when movie not in cache
func TestRunScrape_NoMovieInCache(t *testing.T) {
	configPath, testCfg := setupScrapeTestDB(t)

	withTempConfigFile(t, configPath, func() {
		deps := createTestDependencies(t, testCfg)
		defer deps.Close()

		cmd := &cobra.Command{}
		cmd.Flags().Bool("force", false, "force flag")

		stdout, _ := captureOutput(t, func() {
			resetLoggerForTests(t) // Reset logger INSIDE captureOutput so it uses the pipe
			err := runScrape(cmd, []string{"NOTFOUND-999"}, deps)
			// Expect error because real scrapers will fail without network/mocks
			assert.Error(t, err, "Should error when scrapers fail to find movie")
		})

		// Should attempt to scrape (will fail with real scrapers that need network)
		// We're testing the flow, not success
		assert.Contains(t, stdout, "Scraping metadata for: NOTFOUND-999")

		// Since real scrapers need network access and we're not mocking,
		// we expect the command to attempt scraping and then error
		// The test verifies the command runs without panicking
	})
}

// TestRunScrape_CustomScraperPriority tests --scrapers flag
func TestRunScrape_CustomScraperPriority(t *testing.T) {
	configPath, testCfg := setupScrapeTestDB(t)

	withTempConfigFile(t, configPath, func() {
		deps := createTestDependencies(t, testCfg)
		defer deps.Close()

		// Set custom scrapers via global flag
		originalScrapersFlag := scrapersFlag
		scrapersFlag = []string{"dmm"}
		defer func() { scrapersFlag = originalScrapersFlag }()

		cmd := &cobra.Command{}
		cmd.Flags().Bool("force", false, "force flag")
		cmd.Flags().StringSlice("scrapers", []string{"dmm"}, "scrapers")

		stdout, _ := captureOutput(t, func() {
			resetLoggerForTests(t) // Reset logger INSIDE captureOutput so it uses the pipe
			err := runScrape(cmd, []string{"NOTFOUND-999"}, deps)
			// Expect error because real scrapers will fail without network/mocks
			assert.Error(t, err, "Should error when scrapers fail to find movie")
		})

		// Should show custom scraper usage
		assert.Contains(t, stdout, "Scraping metadata")

		// The command should execute and attempt to use custom scrapers
		// Actual scraping will fail due to network, but we're testing flag handling
	})
}

// TestRunScrape_ContentIDResolution tests ContentID resolution logic
func TestRunScrape_ContentIDResolution(t *testing.T) {
	configPath, testCfg := setupScrapeTestDB(t)

	withTempConfigFile(t, configPath, func() {
		deps := createTestDependencies(t, testCfg)
		defer deps.Close()

		// Test with a movie that has both ID and ContentID
		repo := database.NewMovieRepository(deps.DB)
		movie := createTestMovie("IPX-535", "Test Movie")
		movie.ContentID = "ipx00535" // Different from ID
		err := repo.Upsert(movie)
		require.NoError(t, err)

		cmd := &cobra.Command{}
		cmd.Flags().Bool("force", false, "force flag")

		stdout, _ := captureOutput(t, func() {
			err := runScrape(cmd, []string{"IPX-535"}, deps)
			require.NoError(t, err)
		})

		// Should display both ID and ContentID if different
		assert.Contains(t, stdout, "IPX-535")
		assert.Contains(t, stdout, "ipx00535")
	})
}
