package batch

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/javinizer/javinizer-go/internal/config"
	"github.com/javinizer/javinizer-go/internal/models"
	"github.com/javinizer/javinizer-go/internal/worker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMultipartPreviewEndToEnd(t *testing.T) {
	// Initialize WebSocket hub
	initTestWebSocket(t)

	// Create config with multipart templates
	cfg := &config.Config{
		Output: config.OutputConfig{
			FolderFormat:     "<ID>",
			FileFormat:       "<ID>",
			PosterFormat:     "<ID><IF:MULTIPART>-pt<PART></IF>-poster.jpg",
			FanartFormat:     "<ID><IF:MULTIPART>-pt<PART></IF>-fanart.jpg",
			ScreenshotFolder: "extrafanart",
			// Enable media downloads for preview testing
			DownloadCover:       true,
			DownloadPoster:      true,
			DownloadExtrafanart: true,
		},
		API: config.APIConfig{
			Security: config.SecurityConfig{
				AllowedDirectories: []string{"/path", "/output"},
			},
		},
	}

	deps := createTestDeps(t, cfg, "")

	// Create job with multipart files
	job := deps.JobQueue.CreateJob([]string{
		"/path/to/STSK-074-pt1.mp4",
		"/path/to/STSK-074-pt2.mp4",
	})

	movie := &models.Movie{
		ID:    "STSK-074",
		Title: "Multipart Test Movie",
	}

	// Simulate what RunBatchScrapeOnce does - add pt1 first
	result1 := &worker.FileResult{
		FilePath:    "/path/to/STSK-074-pt1.mp4",
		MovieID:     "STSK-074",
		Status:      worker.JobStatusCompleted,
		Data:        movie,
		IsMultiPart: true,
		PartNumber:  1,
		PartSuffix:  "-pt1",
		StartedAt:   time.Now(),
	}
	job.UpdateFileResult("/path/to/STSK-074-pt1.mp4", result1)

	result2 := &worker.FileResult{
		FilePath:    "/path/to/STSK-074-pt2.mp4",
		MovieID:     "STSK-074",
		Status:      worker.JobStatusCompleted,
		Data:        movie,
		IsMultiPart: true,
		PartNumber:  2,
		PartSuffix:  "-pt2",
		StartedAt:   time.Now(),
	}
	job.UpdateFileResult("/path/to/STSK-074-pt2.mp4", result2)

	// Verify what's in the job
	status := job.GetStatus()
	t.Logf("Job has %d results", len(status.Results))
	for path, res := range status.Results {
		t.Logf("  %s: IsMultiPart=%v, PartNumber=%d, PartSuffix=%q",
			path, res.IsMultiPart, res.PartNumber, res.PartSuffix)
	}

	// Now call the preview endpoint
	router := gin.New()
	router.POST("/batch/:id/movies/:movieId/preview", previewOrganize(deps))

	reqBody := OrganizePreviewRequest{
		Destination: "/output",
		CopyOnly:    false,
	}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/batch/"+job.ID+"/movies/STSK-074/preview", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	t.Logf("Response status: %d", w.Code)
	t.Logf("Response body: %s", w.Body.String())

	assert.Equal(t, 200, w.Code)

	var response OrganizePreviewResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	t.Logf("PosterPath: %s", response.PosterPath)
	t.Logf("FanartPath: %s", response.FanartPath)

	// These are the key assertions - poster and fanart should have -pt1 suffix
	assert.Contains(t, response.PosterPath, "-pt1-poster", "poster should have pt1 suffix")
	assert.Contains(t, response.FanartPath, "-pt1-fanart", "fanart should have pt1 suffix")
}

func TestMultipartPreviewLetterPatternFiles(t *testing.T) {
	// Test case: Letter-pattern multipart files (cemd-349-a.mp4, cemd-349-b.mp4)
	// These should NOT cause conflicts because each part gets a unique filename

	initTestWebSocket(t)

	cfg := &config.Config{
		Output: config.OutputConfig{
			FolderFormat:     "<ID>",
			FileFormat:       "<ID><IF:MULTIPART>-pt<PART></IF>", // Uses IsMultiPart conditional
			PosterFormat:     "<ID><IF:MULTIPART>-pt<PART></IF>-poster.jpg",
			FanartFormat:     "<ID><IF:MULTIPART>-pt<PART></IF>-fanart.jpg",
			ScreenshotFolder: "extrafanart",
			DownloadCover:    true,
			DownloadPoster:   true,
		},
		API: config.APIConfig{
			Security: config.SecurityConfig{
				AllowedDirectories: []string{"/path", "/output"},
			},
		},
	}

	deps := createTestDeps(t, cfg, "")

	// Create job with letter-pattern multipart files
	job := deps.JobQueue.CreateJob([]string{
		"/path/to/cemd-349-a.mp4",
		"/path/to/cemd-349-b.mp4",
	})

	movie := &models.Movie{
		ID:    "CEMD-349",
		Title: "Test Movie",
	}

	// Simulate discovery phase results with IsMultiPart=true for letter patterns
	result1 := &worker.FileResult{
		FilePath:    "/path/to/cemd-349-a.mp4",
		MovieID:     "CEMD-349",
		Status:      worker.JobStatusCompleted,
		Data:        movie,
		IsMultiPart: true, // Set by ValidateMultipartInDirectory during discovery
		PartNumber:  1,    // A = 1
		PartSuffix:  "-A", // Letter suffix
		StartedAt:   time.Now(),
	}
	job.UpdateFileResult("/path/to/cemd-349-a.mp4", result1)

	result2 := &worker.FileResult{
		FilePath:    "/path/to/cemd-349-b.mp4",
		MovieID:     "CEMD-349",
		Status:      worker.JobStatusCompleted,
		Data:        movie,
		IsMultiPart: true, // Set by ValidateMultipartInDirectory during discovery
		PartNumber:  2,    // B = 2
		PartSuffix:  "-B", // Letter suffix
		StartedAt:   time.Now(),
	}
	job.UpdateFileResult("/path/to/cemd-349-b.mp4", result2)

	// Verify job has the correct multipart metadata
	status := job.GetStatus()
	for path, res := range status.Results {
		t.Logf("  %s: IsMultiPart=%v, PartNumber=%d, PartSuffix=%q",
			path, res.IsMultiPart, res.PartNumber, res.PartSuffix)
		assert.True(t, res.IsMultiPart, "file should be marked as multipart")
	}

	// Test preview for first file
	router := gin.New()
	router.POST("/batch/:id/movies/:movieId/preview", previewOrganize(deps))

	reqBody := OrganizePreviewRequest{Destination: "/output"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/batch/"+job.ID+"/movies/CEMD-349/preview", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	t.Logf("Response status: %d", w.Code)
	t.Logf("Response body: %s", w.Body.String())

	assert.Equal(t, 200, w.Code)

	var response OrganizePreviewResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	t.Logf("PosterPath: %s", response.PosterPath)

	// Poster should have -pt1 suffix (part number from discovery phase)
	assert.Contains(t, response.PosterPath, "-pt1-poster", "poster should have pt1 suffix from PartNumber")

	// Verify the file paths in response have unique part suffixes (no conflicts)
	assert.Contains(t, response.FullPath, "CEMD-349-pt1.mp4", "full path should have pt1 suffix")
}

func TestMultipartPreviewSingleFile(t *testing.T) {
	// Test case: User submits only ONE multipart file (e.g., just pt1)
	// The poster should still use the multipart template

	initTestWebSocket(t)

	cfg := &config.Config{
		Output: config.OutputConfig{
			FolderFormat:     "<ID>",
			FileFormat:       "<ID>",
			PosterFormat:     "<ID><IF:MULTIPART>-pt<PART></IF>-poster.jpg",
			FanartFormat:     "<ID><IF:MULTIPART>-pt<PART></IF>-fanart.jpg",
			ScreenshotFolder: "extrafanart",
			// Enable media downloads for preview testing
			DownloadCover:       true,
			DownloadPoster:      true,
			DownloadExtrafanart: true,
		},
		API: config.APIConfig{
			Security: config.SecurityConfig{
				AllowedDirectories: []string{"/path", "/output"},
			},
		},
	}

	deps := createTestDeps(t, cfg, "")

	// Create job with ONLY ONE multipart file
	job := deps.JobQueue.CreateJob([]string{"/path/to/STSK-074-pt1.mp4"})

	movie := &models.Movie{
		ID:    "STSK-074",
		Title: "Multipart Test Movie",
	}

	result := &worker.FileResult{
		FilePath:    "/path/to/STSK-074-pt1.mp4",
		MovieID:     "STSK-074",
		Status:      worker.JobStatusCompleted,
		Data:        movie,
		IsMultiPart: true,
		PartNumber:  1,
		PartSuffix:  "-pt1",
		StartedAt:   time.Now(),
	}
	job.UpdateFileResult("/path/to/STSK-074-pt1.mp4", result)

	router := gin.New()
	router.POST("/batch/:id/movies/:movieId/preview", previewOrganize(deps))

	reqBody := OrganizePreviewRequest{Destination: "/output"}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/batch/"+job.ID+"/movies/STSK-074/preview", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	t.Logf("Response: %s", w.Body.String())

	assert.Equal(t, 200, w.Code)

	var response OrganizePreviewResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	t.Logf("PosterPath: %s", response.PosterPath)
	t.Logf("FanartPath: %s", response.FanartPath)

	// Even with single file, if it's multipart, poster should have -pt1
	assert.Contains(t, response.PosterPath, "-pt1-poster", "poster should have pt1 suffix")
	assert.Contains(t, response.FanartPath, "-pt1-fanart", "fanart should have pt1 suffix")
}
