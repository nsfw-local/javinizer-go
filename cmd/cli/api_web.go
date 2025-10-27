package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/javinizer/javinizer-go/internal/aggregator"
	"github.com/javinizer/javinizer-go/internal/database"
	"github.com/javinizer/javinizer-go/internal/downloader"
	"github.com/javinizer/javinizer-go/internal/logging"
	"github.com/javinizer/javinizer-go/internal/matcher"
	"github.com/javinizer/javinizer-go/internal/models"
	"github.com/javinizer/javinizer-go/internal/nfo"
	"github.com/javinizer/javinizer-go/internal/organizer"
	"github.com/javinizer/javinizer-go/internal/scanner"
	"github.com/javinizer/javinizer-go/internal/template"
	ws "github.com/javinizer/javinizer-go/internal/websocket"
	"github.com/javinizer/javinizer-go/internal/worker"
)

var (
	jobQueue   *worker.JobQueue
	wsHub      *ws.Hub
	wsUpgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for development
		},
	}
)

// ScanRequest represents a directory scan request
type ScanRequest struct {
	Path      string `json:"path" binding:"required" example:"/path/to/videos"`
	Recursive bool   `json:"recursive" example:"true"`
}

// ScanResponse represents scan results
type ScanResponse struct {
	Files   []FileInfo `json:"files"`
	Count   int        `json:"count" example:"10"`
	Skipped []string   `json:"skipped,omitempty"`
}

// FileInfo represents file or directory information
type FileInfo struct {
	Name     string `json:"name" example:"video.mp4"`
	Path     string `json:"path" example:"/path/to/video.mp4"`
	IsDir    bool   `json:"is_dir" example:"false"`
	Size     int64  `json:"size" example:"1024000000"`
	ModTime  string `json:"mod_time" example:"2024-01-15T10:30:00Z"`
	MovieID  string `json:"movie_id,omitempty" example:"IPX-535"`
	Matched  bool   `json:"matched" example:"true"`
}

// BatchScrapeRequest represents a batch scrape request
type BatchScrapeRequest struct {
	Files       []string `json:"files" binding:"required"`
	Strict      bool     `json:"strict" example:"false"`
	Force       bool     `json:"force" example:"false"`
	Destination string   `json:"destination,omitempty" example:"/path/to/output"`
	Update      bool     `json:"update" example:"false"` // Update mode: only create/update metadata files without moving video files
}

// BatchScrapeResponse represents batch scrape response
type BatchScrapeResponse struct {
	JobID string `json:"job_id" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// OrganizeRequest represents an organize request
type OrganizeRequest struct {
	Destination string `json:"destination" binding:"required" example:"/path/to/output"`
	CopyOnly    bool   `json:"copy_only" example:"false"`
}

// OrganizePreviewRequest represents a preview request
type OrganizePreviewRequest struct {
	Destination string `json:"destination" binding:"required" example:"/path/to/output"`
	CopyOnly    bool   `json:"copy_only" example:"false"`
}

// OrganizePreviewResponse represents the expected output structure
type OrganizePreviewResponse struct {
	FolderName      string   `json:"folder_name" example:"IPX-535 [IdeaPocket] - Beautiful Woman (2021)"`
	FileName        string   `json:"file_name" example:"IPX-535"`
	FullPath        string   `json:"full_path" example:"/path/to/output/IPX-535 [IdeaPocket] - Beautiful Woman (2021)/IPX-535.mp4"`
	NFOPath         string   `json:"nfo_path" example:"/path/to/output/IPX-535 [IdeaPocket] - Beautiful Woman (2021)/IPX-535.nfo"`
	PosterPath      string   `json:"poster_path" example:"/path/to/output/IPX-535 [IdeaPocket] - Beautiful Woman (2021)/IPX-535-poster.jpg"`
	FanartPath      string   `json:"fanart_path" example:"/path/to/output/IPX-535 [IdeaPocket] - Beautiful Woman (2021)/IPX-535-fanart.jpg"`
	ExtrafanartPath string   `json:"extrafanart_path" example:"/path/to/output/IPX-535 [IdeaPocket] - Beautiful Woman (2021)/extrafanart"`
	Screenshots     []string `json:"screenshots" example:"fanart1.jpg,fanart2.jpg,fanart3.jpg"`
}

// BatchJobResponse represents a batch job status
type BatchJobResponse struct {
	ID          string                    `json:"id"`
	Status      string                    `json:"status"`
	TotalFiles  int                       `json:"total_files"`
	Completed   int                       `json:"completed"`
	Failed      int                       `json:"failed"`
	Progress    float64                   `json:"progress"`
	Results     map[string]*worker.FileResult `json:"results"`
	StartedAt   string                    `json:"started_at"`
	CompletedAt *string                   `json:"completed_at,omitempty"`
}

// BrowseRequest represents a browse request
type BrowseRequest struct {
	Path string `json:"path" example:"/path/to/directory"`
}

// BrowseResponse represents browse results
type BrowseResponse struct {
	CurrentPath string     `json:"current_path" example:"/path/to/directory"`
	ParentPath  string     `json:"parent_path,omitempty" example:"/path/to"`
	Items       []FileInfo `json:"items"`
}

// scanDirectory godoc
// @Summary Scan directory for video files
// @Description Scan a directory for video files and match JAV IDs
// @Tags web
// @Accept json
// @Produce json
// @Param request body ScanRequest true "Scan parameters"
// @Success 200 {object} ScanResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/scan [post]
func scanDirectory(mat *matcher.Matcher) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ScanRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, ErrorResponse{Error: err.Error()})
			return
		}

		// Verify path exists
		if _, err := os.Stat(req.Path); os.IsNotExist(err) {
			c.JSON(400, ErrorResponse{Error: fmt.Sprintf("Path does not exist: %s", req.Path)})
			return
		}

		// Scan directory
		scan := scanner.NewScanner(&cfg.Matching)
		result, err := scan.Scan(req.Path)
		if err != nil {
			c.JSON(500, ErrorResponse{Error: err.Error()})
			return
		}

		// Match IDs
		matchResults := mat.Match(result.Files)

		// Build response
		files := make([]FileInfo, 0, len(result.Files))
		matchMap := make(map[string]*matcher.MatchResult)
		for i, match := range matchResults {
			matchMap[match.File.Path] = &matchResults[i]
		}

		for _, fileInfo := range result.Files {
			match, found := matchMap[fileInfo.Path]
			info, _ := os.Stat(fileInfo.Path)

			apiFileInfo := FileInfo{
				Name:    fileInfo.Name,
				Path:    fileInfo.Path,
				IsDir:   false,
				Size:    info.Size(),
				ModTime: info.ModTime().Format("2006-01-02T15:04:05Z07:00"),
				Matched: found,
			}
			if found {
				apiFileInfo.MovieID = match.ID
			}
			files = append(files, apiFileInfo)
		}

		c.JSON(200, ScanResponse{
			Files:   files,
			Count:   len(files),
			Skipped: result.Skipped,
		})
	}
}

// getCurrentWorkingDirectory godoc
// @Summary Get current working directory
// @Description Returns the server's current working directory
// @Tags web
// @Produce json
// @Success 200 {object} map[string]string
// @Router /api/v1/cwd [get]
func getCurrentWorkingDirectory() gin.HandlerFunc {
	return func(c *gin.Context) {
		cwd, err := os.Getwd()
		if err != nil {
			c.JSON(500, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(200, gin.H{"path": cwd})
	}
}

// browseDirectory godoc
// @Summary Browse directory
// @Description Browse a directory and list its contents
// @Tags web
// @Accept json
// @Produce json
// @Param request body BrowseRequest true "Browse parameters"
// @Success 200 {object} BrowseResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/browse [post]
func browseDirectory() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req BrowseRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, ErrorResponse{Error: err.Error()})
			return
		}

		// Default to current directory if not specified
		if req.Path == "" {
			req.Path, _ = os.Getwd()
		}

		// Verify path exists and is a directory
		info, err := os.Stat(req.Path)
		if os.IsNotExist(err) {
			c.JSON(400, ErrorResponse{Error: fmt.Sprintf("Path does not exist: %s", req.Path)})
			return
		}
		if !info.IsDir() {
			c.JSON(400, ErrorResponse{Error: fmt.Sprintf("Path is not a directory: %s", req.Path)})
			return
		}

		// Read directory contents
		entries, err := os.ReadDir(req.Path)
		if err != nil {
			c.JSON(500, ErrorResponse{Error: err.Error()})
			return
		}

		// Build response
		items := make([]FileInfo, 0, len(entries))
		for _, entry := range entries {
			fullPath := filepath.Join(req.Path, entry.Name())
			info, err := entry.Info()
			if err != nil {
				continue
			}

			items = append(items, FileInfo{
				Name:    entry.Name(),
				Path:    fullPath,
				IsDir:   entry.IsDir(),
				Size:    info.Size(),
				ModTime: info.ModTime().Format("2006-01-02T15:04:05Z07:00"),
			})
		}

		// Get parent path
		parentPath := filepath.Dir(req.Path)
		if parentPath == req.Path {
			parentPath = "" // Root directory
		}

		c.JSON(200, BrowseResponse{
			CurrentPath: req.Path,
			ParentPath:  parentPath,
			Items:       items,
		})
	}
}

// batchScrape godoc
// @Summary Batch scrape movies
// @Description Scrape metadata for multiple movies in batch
// @Tags web
// @Accept json
// @Produce json
// @Param request body BatchScrapeRequest true "Batch scrape parameters"
// @Success 200 {object} BatchScrapeResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/batch/scrape [post]
func batchScrape(registry *models.ScraperRegistry, agg *aggregator.Aggregator, movieRepo *database.MovieRepository, mat *matcher.Matcher) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req BatchScrapeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, ErrorResponse{Error: err.Error()})
			return
		}

		// Create job
		job := jobQueue.CreateJob(req.Files)

		// Start processing in background
		go processBatchJob(job, registry, agg, movieRepo, mat, req.Strict, req.Force, req.Destination)

		c.JSON(200, BatchScrapeResponse{
			JobID: job.ID,
		})
	}
}

// processBatchJob processes a batch scraping job (metadata only, no file organization)
func processBatchJob(job *worker.BatchJob, registry *models.ScraperRegistry, agg *aggregator.Aggregator, movieRepo *database.MovieRepository, mat *matcher.Matcher, strict, force bool, destination string) {
	ctx, cancel := context.WithCancel(context.Background())
	job.CancelFunc = cancel
	defer cancel()

	job.MarkStarted()

	for i, filePath := range job.Files {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			job.MarkCancelled()
			return
		default:
		}

		// Get movie ID from filename
		fileInfo := scanner.FileInfo{
			Path:      filePath,
			Name:      filepath.Base(filePath),
			Extension: filepath.Ext(filePath),
			Dir:       filepath.Dir(filePath),
		}
		matchResults := mat.Match([]scanner.FileInfo{fileInfo})
		if len(matchResults) == 0 {
			result := &worker.FileResult{
				FilePath:  filePath,
				Status:    worker.JobStatusFailed,
				Error:     "Could not extract movie ID from filename",
				StartedAt: time.Now(),
			}
			now := time.Now()
			result.EndedAt = &now
			job.UpdateFileResult(filePath, result)

			// Broadcast progress
			wsHub.BroadcastProgress(&ws.ProgressMessage{
				JobID:     job.ID,
				FileIndex: i,
				FilePath:  filePath,
				Status:    string(worker.JobStatusFailed),
				Progress:  job.Progress,
				Error:     result.Error,
			})
			continue
		}

		movieID := matchResults[0].ID
		result := &worker.FileResult{
			FilePath:  filePath,
			MovieID:   movieID,
			Status:    worker.JobStatusRunning,
			StartedAt: time.Now(),
		}
		job.UpdateFileResult(filePath, result)

		// Broadcast start
		wsHub.BroadcastProgress(&ws.ProgressMessage{
			JobID:     job.ID,
			FileIndex: i,
			FilePath:  filePath,
			Status:    "running",
			Progress:  job.Progress,
			Message:   fmt.Sprintf("Scraping %s", movieID),
		})

		// Check cache first
		if !force {
			existing, err := movieRepo.FindByID(movieID)
			if err == nil && existing != nil {
				result.Status = worker.JobStatusCompleted
				result.Data = existing
				now := time.Now()
				result.EndedAt = &now
				job.UpdateFileResult(filePath, result)

				wsHub.BroadcastProgress(&ws.ProgressMessage{
					JobID:     job.ID,
					FileIndex: i,
					FilePath:  filePath,
					Status:    "completed",
					Progress:  job.Progress,
					Message:   "Found in cache",
				})
				continue
			}
		}

		// Scrape from sources
		results := []*models.ScraperResult{}
		errors := []string{}

		for _, scraper := range registry.GetByPriority(cfg.Scrapers.Priority) {
			scraperResult, err := scraper.Search(movieID)
			if err != nil {
				errors = append(errors, fmt.Sprintf("%s: %v", scraper.Name(), err))
				continue
			}
			results = append(results, scraperResult)
		}

		if len(results) == 0 {
			result.Status = worker.JobStatusFailed
			result.Error = fmt.Sprintf("Movie not found: %s", strings.Join(errors, "; "))
			now := time.Now()
			result.EndedAt = &now
			job.UpdateFileResult(filePath, result)

			wsHub.BroadcastProgress(&ws.ProgressMessage{
				JobID:     job.ID,
				FileIndex: i,
				FilePath:  filePath,
				Status:    "failed",
				Progress:  job.Progress,
				Error:     result.Error,
			})
			continue
		}

		// Aggregate results
		movie, err := agg.Aggregate(results)
		if err != nil {
			result.Status = worker.JobStatusFailed
			result.Error = err.Error()
			now := time.Now()
			result.EndedAt = &now
			job.UpdateFileResult(filePath, result)

			wsHub.BroadcastProgress(&ws.ProgressMessage{
				JobID:     job.ID,
				FileIndex: i,
				FilePath:  filePath,
				Status:    "failed",
				Progress:  job.Progress,
				Error:     err.Error(),
			})
			continue
		}

		movie.OriginalFileName = filepath.Base(filePath)

		// Save to database
		if err := movieRepo.Upsert(movie); err != nil {
			logging.Errorf("Failed to save movie to database: %v", err)
		}

		// Reload movie from database to get associations (actresses, genres)
		reloadedMovie, err := movieRepo.FindByID(movie.ID)
		if err != nil {
			logging.Errorf("Failed to reload movie from database: %v", err)
			reloadedMovie = movie // Fallback to original if reload fails
		}

		result.Status = worker.JobStatusCompleted
		result.Data = reloadedMovie
		now := time.Now()
		result.EndedAt = &now
		job.UpdateFileResult(filePath, result)

		wsHub.BroadcastProgress(&ws.ProgressMessage{
			JobID:     job.ID,
			FileIndex: i,
			FilePath:  filePath,
			Status:    "completed",
			Progress:  job.Progress,
			Message:   fmt.Sprintf("Scraped %s successfully", movieID),
		})
	}

	job.MarkCompleted()

	// Broadcast final completion
	wsHub.BroadcastProgress(&ws.ProgressMessage{
		JobID:    job.ID,
		Status:   "completed",
		Progress: 100,
		Message:  fmt.Sprintf("Completed %d of %d files", job.Completed, job.TotalFiles),
	})
}

// processOrganizeJob processes file organization for a completed scrape job
func processOrganizeJob(job *worker.BatchJob, mat *matcher.Matcher, destination string, copyOnly bool) {
	// Initialize organizer, downloader, and NFO generator
	org := organizer.NewOrganizer(&cfg.Output)
	dl := downloader.NewDownloader(&cfg.Output, "Javinizer/1.0")
	nfoGen := nfo.NewGenerator(nil) // Use default config

	// Broadcast organization started
	wsHub.BroadcastProgress(&ws.ProgressMessage{
		JobID:    job.ID,
		Status:   "organizing",
		Progress: 0,
		Message:  "Starting file organization",
	})

	status := job.GetStatus()
	organized := 0
	failed := 0

	for filePath, fileResult := range status.Results {
		// Skip files that failed during scraping
		if fileResult.Status != worker.JobStatusCompleted || fileResult.Data == nil {
			continue
		}

		movie, ok := fileResult.Data.(*models.Movie)
		if !ok {
			logging.Errorf("Invalid movie data type for file: %s", filePath)
			failed++
			continue
		}

		// Create match result for organizer
		fileInfo := scanner.FileInfo{
			Path:      filePath,
			Name:      filepath.Base(filePath),
			Extension: filepath.Ext(filePath),
			Dir:       filepath.Dir(filePath),
		}
		matchResults := mat.Match([]scanner.FileInfo{fileInfo})
		if len(matchResults) == 0 {
			logging.Errorf("Could not match file: %s", filePath)
			failed++
			continue
		}

		match := matchResults[0]

		// Organize file
		result, err := org.Organize(match, movie, destination, false, false, copyOnly)
		if err != nil {
			logging.Errorf("Failed to organize %s: %v", filePath, err)
			failed++

			wsHub.BroadcastProgress(&ws.ProgressMessage{
				JobID:    job.ID,
				FilePath: filePath,
				Status:   "failed",
				Progress: float64(organized+failed) / float64(len(status.Results)) * 100,
				Error:    err.Error(),
			})
			continue
		}

		// Download artwork if file was moved
		if result.Moved && cfg.Output.DownloadPoster {
			if _, err := dl.DownloadPoster(movie, result.FolderPath); err != nil {
				logging.Errorf("Failed to download poster for %s: %v", movie.ID, err)
			}
		}

		if result.Moved && cfg.Output.DownloadCover {
			if _, err := dl.DownloadCover(movie, result.FolderPath); err != nil {
				logging.Errorf("Failed to download cover for %s: %v", movie.ID, err)
			}
		}

		// Download screenshots if enabled
		if result.Moved && cfg.Output.DownloadExtrafanart {
			if _, err := dl.DownloadExtrafanart(movie, result.FolderPath); err != nil {
				logging.Errorf("Failed to download screenshots for %s: %v", movie.ID, err)
			}
		}

		// Generate NFO file
		if result.Moved {
			nfoPath := filepath.Join(result.FolderPath, strings.TrimSuffix(result.FileName, filepath.Ext(result.FileName))+".nfo")
			if err := nfoGen.Generate(movie, nfoPath); err != nil {
				logging.Errorf("Failed to generate NFO for %s: %v", movie.ID, err)
			}
		}

		organized++

		wsHub.BroadcastProgress(&ws.ProgressMessage{
			JobID:    job.ID,
			FilePath: filePath,
			Status:   "organized",
			Progress: float64(organized+failed) / float64(len(status.Results)) * 100,
			Message:  fmt.Sprintf("Organized %s", movie.ID),
		})
	}

	// Broadcast final completion
	wsHub.BroadcastProgress(&ws.ProgressMessage{
		JobID:    job.ID,
		Status:   "organization_completed",
		Progress: 100,
		Message:  fmt.Sprintf("Organized %d files, %d failed", organized, failed),
	})
}

// getBatchJob godoc
// @Summary Get batch job status
// @Description Retrieve the status of a batch scraping job
// @Tags web
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} BatchJobResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/batch/{id} [get]
func getBatchJob() gin.HandlerFunc {
	return func(c *gin.Context) {
		jobID := c.Param("id")

		job, ok := jobQueue.GetJob(jobID)
		if !ok {
			c.JSON(404, ErrorResponse{Error: "Job not found"})
			return
		}

		status := job.GetStatus()
		var completedAt *string
		if status.CompletedAt != nil {
			str := status.CompletedAt.Format("2006-01-02T15:04:05Z07:00")
			completedAt = &str
		}

		c.JSON(200, BatchJobResponse{
			ID:          status.ID,
			Status:      string(status.Status),
			TotalFiles:  status.TotalFiles,
			Completed:   status.Completed,
			Failed:      status.Failed,
			Progress:    status.Progress,
			Results:     status.Results,
			StartedAt:   status.StartedAt.Format("2006-01-02T15:04:05Z07:00"),
			CompletedAt: completedAt,
		})
	}
}

// cancelBatchJob godoc
// @Summary Cancel batch job
// @Description Cancel a running batch scraping job
// @Tags web
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/batch/{id}/cancel [post]
func cancelBatchJob() gin.HandlerFunc {
	return func(c *gin.Context) {
		jobID := c.Param("id")

		job, ok := jobQueue.GetJob(jobID)
		if !ok {
			c.JSON(404, ErrorResponse{Error: "Job not found"})
			return
		}

		job.Cancel()

		c.JSON(200, gin.H{"message": "Job cancelled successfully"})
	}
}

// UpdateMovieRequest represents the update movie request payload
type UpdateMovieRequest struct {
	Movie *models.Movie `json:"movie" binding:"required"`
}

// updateBatchMovie godoc
// @Summary Update movie in batch job
// @Description Update a movie's metadata within a batch job's results
// @Tags web
// @Accept json
// @Produce json
// @Param id path string true "Job ID"
// @Param movieId path string true "Movie ID"
// @Param request body UpdateMovieRequest true "Updated movie data"
// @Success 200 {object} MovieResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/batch/{id}/movies/{movieId} [patch]
func updateBatchMovie(movieRepo *database.MovieRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		jobID := c.Param("id")
		movieID := c.Param("movieId")

		var req UpdateMovieRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, ErrorResponse{Error: err.Error()})
			return
		}

		// Get the batch job
		job, ok := jobQueue.GetJob(jobID)
		if !ok {
			c.JSON(404, ErrorResponse{Error: "Job not found"})
			return
		}

		// Find the file result with this movie ID
		status := job.GetStatus()
		var foundFilePath string
		var foundResult *worker.FileResult
		for filePath, result := range status.Results {
			if result.MovieID == movieID {
				foundFilePath = filePath
				foundResult = result
				break
			}
		}

		if foundResult == nil {
			c.JSON(404, ErrorResponse{Error: fmt.Sprintf("Movie %s not found in job", movieID)})
			return
		}

		// Update the movie data in the file result
		foundResult.Data = req.Movie
		job.UpdateFileResult(foundFilePath, foundResult)

		// Also update in database if it exists
		if err := movieRepo.Upsert(req.Movie); err != nil {
			logging.Errorf("Failed to update movie in database: %v", err)
			// Don't fail the request if DB update fails
		}

		c.JSON(200, MovieResponse{Movie: req.Movie})
	}
}

// organizeJob godoc
// @Summary Organize batch job files
// @Description Organize files from a completed scrape job (move files, download artwork, create NFO)
// @Tags web
// @Accept json
// @Produce json
// @Param id path string true "Job ID"
// @Param request body OrganizeRequest true "Organization parameters"
// @Success 200 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/batch/{id}/organize [post]
func organizeJob(mat *matcher.Matcher) gin.HandlerFunc {
	return func(c *gin.Context) {
		jobID := c.Param("id")

		var req OrganizeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, ErrorResponse{Error: err.Error()})
			return
		}

		job, ok := jobQueue.GetJob(jobID)
		if !ok {
			c.JSON(404, ErrorResponse{Error: "Job not found"})
			return
		}

		// Check if job is in correct state (completed scraping)
		status := job.GetStatus()
		if status.Status != worker.JobStatusCompleted {
			c.JSON(400, ErrorResponse{Error: "Job must be completed before organizing"})
			return
		}

		// Start organization in background
		go processOrganizeJob(job, mat, req.Destination, req.CopyOnly)

		c.JSON(200, gin.H{"message": "Organization started"})
	}
}

// handleWebSocket handles WebSocket connections
func handleWebSocket() gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			logging.Errorf("Failed to upgrade to websocket: %v", err)
			return
		}

		client := ws.NewClient(conn)
		wsHub.Register(client)

		// Start pumps
		go client.WritePump()
		go client.ReadPump(wsHub)
	}
}

// previewOrganize godoc
// @Summary Preview organize output
// @Description Generate a preview of the expected output structure for a movie
// @Tags web
// @Accept json
// @Produce json
// @Param id path string true "Job ID"
// @Param movieId path string true "Movie ID"
// @Param request body OrganizePreviewRequest true "Preview parameters"
// @Success 200 {object} OrganizePreviewResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/batch/{id}/movies/{movieId}/preview [post]
func previewOrganize() gin.HandlerFunc {
	return func(c *gin.Context) {
		jobID := c.Param("id")
		movieID := c.Param("movieId")

		var req OrganizePreviewRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, ErrorResponse{Error: err.Error()})
			return
		}

		// Get the batch job
		job, ok := jobQueue.GetJob(jobID)
		if !ok {
			c.JSON(404, ErrorResponse{Error: "Job not found"})
			return
		}

		// Find the movie in the job results
		status := job.GetStatus()
		var movie *models.Movie
		for _, result := range status.Results {
			if result.MovieID == movieID {
				if result.Data != nil {
					var ok bool
					movie, ok = result.Data.(*models.Movie)
					if !ok {
						c.JSON(500, ErrorResponse{Error: "Invalid movie data type"})
						return
					}
				}
				break
			}
		}

		if movie == nil {
			c.JSON(404, ErrorResponse{Error: fmt.Sprintf("Movie %s not found in job", movieID)})
			return
		}

		// Create template context from movie
		ctx := template.NewContextFromMovie(movie)
		templateEngine := template.NewEngine()

		// Generate folder name
		folderName, err := templateEngine.Execute(cfg.Output.FolderFormat, ctx)
		if err != nil {
			c.JSON(500, ErrorResponse{Error: fmt.Sprintf("Failed to generate folder name: %v", err)})
			return
		}
		folderName = template.SanitizeFolderPath(folderName)

		// Generate file name
		fileName, err := templateEngine.Execute(cfg.Output.FileFormat, ctx)
		if err != nil {
			c.JSON(500, ErrorResponse{Error: fmt.Sprintf("Failed to generate file name: %v", err)})
			return
		}
		fileName = template.SanitizeFilename(fileName)

		// Build paths
		folderPath := filepath.Join(req.Destination, folderName)
		videoPath := filepath.Join(folderPath, fileName+".mp4") // Using .mp4 as placeholder
		nfoPath := filepath.Join(folderPath, fileName+".nfo")
		posterPath := filepath.Join(folderPath, fileName+"-poster.jpg")
		fanartPath := filepath.Join(folderPath, fileName+"-fanart.jpg")
		extrafanartPath := filepath.Join(folderPath, "extrafanart")

		// Generate screenshot names
		screenshots := []string{}
		if movie.Screenshots != nil && len(movie.Screenshots) > 0 {
			for i := range movie.Screenshots {
				screenshots = append(screenshots, fmt.Sprintf("fanart%d.jpg", i+1))
			}
		}

		c.JSON(200, OrganizePreviewResponse{
			FolderName:      folderName,
			FileName:        fileName,
			FullPath:        videoPath,
			NFOPath:         nfoPath,
			PosterPath:      posterPath,
			FanartPath:      fanartPath,
			ExtrafanartPath: extrafanartPath,
			Screenshots:     screenshots,
		})
	}
}
