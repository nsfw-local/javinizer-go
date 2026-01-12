package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServeTempPoster(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create temp directory structure
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tempDir))
	defer os.Chdir(originalWd)

	// Create data/temp/posters/test-job-id directory
	jobID := "test-job-id"
	posterDir := filepath.Join("data", "temp", "posters", jobID)
	require.NoError(t, os.MkdirAll(posterDir, 0755))

	// Create a test poster file
	posterPath := filepath.Join(posterDir, "test-poster.jpg")
	require.NoError(t, os.WriteFile(posterPath, []byte("fake jpeg data"), 0644))

	tests := []struct {
		name           string
		jobID          string
		filename       string
		expectedStatus int
	}{
		{
			name:           "valid request",
			jobID:          jobID,
			filename:       "test-poster.jpg",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "path traversal in jobID",
			jobID:          "../../../etc",
			filename:       "passwd",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "path traversal in filename",
			jobID:          jobID,
			filename:       "../../../etc/passwd",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "non-jpg extension",
			jobID:          jobID,
			filename:       "test.png",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "non-existent file",
			jobID:          jobID,
			filename:       "nonexistent.jpg",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "jobID with path separator",
			jobID:          "job/id",
			filename:       "test.jpg",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "filename with backslash",
			jobID:          jobID,
			filename:       "test\\poster.jpg",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/temp/posters/:jobId/:filename", serveTempPoster())

			req := httptest.NewRequest(http.MethodGet, "/temp/posters/"+tt.jobID+"/"+tt.filename, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestServeTempPoster_PathTraversalDefenseInDepth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create temp directory structure
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tempDir))
	defer os.Chdir(originalWd)

	// Create the directory structure
	jobID := "test-job"
	posterDir := filepath.Join("data", "temp", "posters", jobID)
	require.NoError(t, os.MkdirAll(posterDir, 0755))

	// Create a sensitive file outside the poster directory
	sensitiveDir := filepath.Join("data", "temp")
	sensitiveFile := filepath.Join(sensitiveDir, "sensitive.jpg")
	require.NoError(t, os.WriteFile(sensitiveFile, []byte("sensitive data"), 0644))

	router := gin.New()
	router.GET("/temp/posters/:jobId/:filename", serveTempPoster())

	// Try to access sensitive.jpg via path traversal
	req := httptest.NewRequest(http.MethodGet, "/temp/posters/"+jobID+"/../sensitive.jpg", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 404 not found
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestServeCroppedPoster(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create temp directory structure
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tempDir))
	defer os.Chdir(originalWd)

	// Create data/posters directory
	posterDir := filepath.Join("data", "posters")
	require.NoError(t, os.MkdirAll(posterDir, 0755))

	// Create a test poster file
	posterPath := filepath.Join(posterDir, "test-cropped.jpg")
	require.NoError(t, os.WriteFile(posterPath, []byte("fake jpeg data"), 0644))

	tests := []struct {
		name           string
		filename       string
		expectedStatus int
		checkHeaders   bool
	}{
		{
			name:           "valid request",
			filename:       "test-cropped.jpg",
			expectedStatus: http.StatusOK,
			checkHeaders:   true,
		},
		{
			name:           "path traversal attempt",
			filename:       "../../../etc/passwd",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "non-jpg extension",
			filename:       "test.png",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "uppercase JPG extension",
			filename:       "nonexistent.JPG",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "non-existent file",
			filename:       "nonexistent.jpg",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "filename with path separator",
			filename:       "subdir/test.jpg",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/posters/:filename", serveCroppedPoster())

			req := httptest.NewRequest(http.MethodGet, "/posters/"+tt.filename, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkHeaders && w.Code == http.StatusOK {
				assert.Equal(t, "public, max-age=86400", w.Header().Get("Cache-Control"))
			}
		})
	}
}

func TestServeCroppedPoster_PathTraversalDefenseInDepth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create temp directory structure
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tempDir))
	defer os.Chdir(originalWd)

	// Create data/posters directory
	posterDir := filepath.Join("data", "posters")
	require.NoError(t, os.MkdirAll(posterDir, 0755))

	// Create a sensitive file outside the poster directory
	sensitiveFile := filepath.Join("data", "sensitive.jpg")
	require.NoError(t, os.WriteFile(sensitiveFile, []byte("sensitive data"), 0644))

	router := gin.New()
	router.GET("/posters/:filename", serveCroppedPoster())

	// Try to access sensitive.jpg via path traversal
	req := httptest.NewRequest(http.MethodGet, "/posters/../sensitive.jpg", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 404 not found
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestServeTempPoster_ValidJpgExtensions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create temp directory structure
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tempDir))
	defer os.Chdir(originalWd)

	jobID := "test-job"
	posterDir := filepath.Join("data", "temp", "posters", jobID)
	require.NoError(t, os.MkdirAll(posterDir, 0755))

	// Create test files with different extensions
	require.NoError(t, os.WriteFile(filepath.Join(posterDir, "test.jpg"), []byte("jpeg"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(posterDir, "test.JPG"), []byte("jpeg"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(posterDir, "test.Jpg"), []byte("jpeg"), 0644))

	router := gin.New()
	router.GET("/temp/posters/:jobId/:filename", serveTempPoster())

	tests := []struct {
		filename       string
		expectedStatus int
	}{
		{"test.jpg", http.StatusOK},
		{"test.JPG", http.StatusOK},
		{"test.Jpg", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/temp/posters/"+jobID+"/"+tt.filename, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
