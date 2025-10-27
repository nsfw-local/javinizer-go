package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/javinizer/javinizer-go/internal/config"
	"github.com/javinizer/javinizer-go/internal/models"
	"github.com/javinizer/javinizer-go/internal/template"
)

// Downloader handles media file downloads
type Downloader struct {
	config     *config.OutputConfig
	httpClient *http.Client
	userAgent  string
}

// DownloadResult represents the result of a download operation
type DownloadResult struct {
	URL          string
	LocalPath    string
	Size         int64
	Downloaded   bool
	Error        error
	Type         MediaType
	Duration     time.Duration
}

// MediaType represents the type of media being downloaded
type MediaType string

const (
	MediaTypeCover       MediaType = "cover"
	MediaTypePoster      MediaType = "poster"
	MediaTypeExtrafanart MediaType = "extrafanart"
	MediaTypeTrailer     MediaType = "trailer"
	MediaTypeActress     MediaType = "actress"
)

// NewDownloader creates a new media downloader
func NewDownloader(cfg *config.OutputConfig, userAgent string) *Downloader {
	return &Downloader{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				IdleConnTimeout:     30 * time.Second,
				DisableCompression:  false,
				MaxIdleConnsPerHost: 2,
			},
		},
		userAgent: userAgent,
	}
}

// SetDownloadExtrafanart sets whether extrafanart downloads are enabled
func (d *Downloader) SetDownloadExtrafanart(enabled bool) {
	d.config.DownloadExtrafanart = enabled
}

// DownloadCover downloads the movie cover image
func (d *Downloader) DownloadCover(movie *models.Movie, destDir string) (*DownloadResult, error) {
	if !d.config.DownloadCover || movie.CoverURL == "" {
		return &DownloadResult{Type: MediaTypeCover, Downloaded: false}, nil
	}

	ctx := template.NewContextFromMovie(movie)
	filename := fmt.Sprintf("%s-poster.jpg", ctx.ID)
	destPath := filepath.Join(destDir, filename)

	return d.download(movie.CoverURL, destPath, MediaTypeCover)
}

// DownloadPoster downloads the movie poster (vertical cover)
func (d *Downloader) DownloadPoster(movie *models.Movie, destDir string) (*DownloadResult, error) {
	if !d.config.DownloadPoster || movie.CoverURL == "" {
		return &DownloadResult{Type: MediaTypePoster, Downloaded: false}, nil
	}

	// For now, poster is the same as cover
	// In the future, we could add poster-specific URL field
	ctx := template.NewContextFromMovie(movie)
	filename := fmt.Sprintf("%s-fanart.jpg", ctx.ID)
	destPath := filepath.Join(destDir, filename)

	return d.download(movie.CoverURL, destPath, MediaTypePoster)
}

// DownloadExtrafanart downloads screenshots to the extrafanart subdirectory
// Extrafanart is used by media centers like Kodi/Plex for background images
// Note: In the original Javinizer, screenshots and extrafanart are the same thing
func (d *Downloader) DownloadExtrafanart(movie *models.Movie, destDir string) ([]DownloadResult, error) {
	if !d.config.DownloadExtrafanart || len(movie.Screenshots) == 0 {
		return []DownloadResult{}, nil
	}

	// Create extrafanart subdirectory
	extrafanartDir := filepath.Join(destDir, "extrafanart")

	results := make([]DownloadResult, 0, len(movie.Screenshots))

	for i, url := range movie.Screenshots {
		// Extrafanart images are typically numbered: fanart1.jpg, fanart2.jpg, etc.
		filename := fmt.Sprintf("fanart%d.jpg", i+1)
		destPath := filepath.Join(extrafanartDir, filename)

		result, err := d.download(url, destPath, MediaTypeExtrafanart)
		if err != nil {
			result = &DownloadResult{
				URL:   url,
				Type:  MediaTypeExtrafanart,
				Error: err,
			}
		}
		results = append(results, *result)
	}

	return results, nil
}

// DownloadTrailer downloads the movie trailer
func (d *Downloader) DownloadTrailer(movie *models.Movie, destDir string) (*DownloadResult, error) {
	if !d.config.DownloadTrailer || movie.TrailerURL == "" {
		return &DownloadResult{Type: MediaTypeTrailer, Downloaded: false}, nil
	}

	ctx := template.NewContextFromMovie(movie)

	// Determine extension from URL
	ext := filepath.Ext(movie.TrailerURL)
	if ext == "" {
		ext = ".mp4" // Default to mp4
	}

	filename := fmt.Sprintf("%s-trailer%s", ctx.ID, ext)
	destPath := filepath.Join(destDir, filename)

	return d.download(movie.TrailerURL, destPath, MediaTypeTrailer)
}

// DownloadActressImages downloads actress thumbnail images
func (d *Downloader) DownloadActressImages(movie *models.Movie, destDir string) ([]DownloadResult, error) {
	if !d.config.DownloadActress || len(movie.Actresses) == 0 {
		return []DownloadResult{}, nil
	}

	results := make([]DownloadResult, 0)

	for _, actress := range movie.Actresses {
		if actress.ThumbURL == "" {
			continue
		}

		// Sanitize actress name for filename
		name := template.SanitizeFilename(actress.FullName())
		filename := fmt.Sprintf("actress-%s.jpg", name)
		destPath := filepath.Join(destDir, ".actors", filename)

		result, err := d.download(actress.ThumbURL, destPath, MediaTypeActress)
		if err != nil {
			result = &DownloadResult{
				URL:   actress.ThumbURL,
				Type:  MediaTypeActress,
				Error: err,
			}
		}
		results = append(results, *result)
	}

	return results, nil
}

// DownloadAll downloads all enabled media types for a movie
func (d *Downloader) DownloadAll(movie *models.Movie, destDir string) ([]DownloadResult, error) {
	results := make([]DownloadResult, 0)

	// Download cover
	if coverResult, err := d.DownloadCover(movie, destDir); err == nil && coverResult != nil {
		results = append(results, *coverResult)
	}

	// Download poster
	if posterResult, err := d.DownloadPoster(movie, destDir); err == nil && posterResult != nil {
		results = append(results, *posterResult)
	}

	// Download extrafanart (screenshots)
	if extrafanart, err := d.DownloadExtrafanart(movie, destDir); err == nil {
		results = append(results, extrafanart...)
	}

	// Download trailer
	if trailerResult, err := d.DownloadTrailer(movie, destDir); err == nil && trailerResult != nil {
		results = append(results, *trailerResult)
	}

	// Download actress images
	if actresses, err := d.DownloadActressImages(movie, destDir); err == nil {
		results = append(results, actresses...)
	}

	return results, nil
}

// download performs the actual HTTP download
func (d *Downloader) download(url, destPath string, mediaType MediaType) (*DownloadResult, error) {
	startTime := time.Now()

	result := &DownloadResult{
		URL:        url,
		LocalPath:  destPath,
		Type:       mediaType,
		Downloaded: false,
	}

	// Check if file already exists
	if info, err := os.Stat(destPath); err == nil {
		result.Size = info.Size()
		result.Downloaded = false // Already exists, not downloaded
		result.Duration = time.Since(startTime)
		return result, nil
	}

	// Create destination directory
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		result.Error = fmt.Errorf("failed to create directory: %w", err)
		result.Duration = time.Since(startTime)
		return result, result.Error
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		result.Error = fmt.Errorf("failed to create request: %w", err)
		result.Duration = time.Since(startTime)
		return result, result.Error
	}

	// Set user agent
	if d.userAgent != "" {
		req.Header.Set("User-Agent", d.userAgent)
	}

	// Execute request
	resp, err := d.httpClient.Do(req)
	if err != nil {
		result.Error = fmt.Errorf("failed to download: %w", err)
		result.Duration = time.Since(startTime)
		return result, result.Error
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		result.Error = fmt.Errorf("bad status code: %d", resp.StatusCode)
		result.Duration = time.Since(startTime)
		return result, result.Error
	}

	// Create temporary file
	tempPath := destPath + ".tmp"
	outFile, err := os.Create(tempPath)
	if err != nil {
		result.Error = fmt.Errorf("failed to create file: %w", err)
		result.Duration = time.Since(startTime)
		return result, result.Error
	}

	// Download to temp file
	written, err := io.Copy(outFile, resp.Body)
	outFile.Close()

	if err != nil {
		os.Remove(tempPath)
		result.Error = fmt.Errorf("failed to write file: %w", err)
		result.Duration = time.Since(startTime)
		return result, result.Error
	}

	// Rename temp file to final destination
	if err := os.Rename(tempPath, destPath); err != nil {
		os.Remove(tempPath)
		result.Error = fmt.Errorf("failed to rename file: %w", err)
		result.Duration = time.Since(startTime)
		return result, result.Error
	}

	result.Size = written
	result.Downloaded = true
	result.Duration = time.Since(startTime)

	return result, nil
}

// GetImageExtension determines the image extension from a URL
func GetImageExtension(url string) string {
	url = strings.ToLower(url)

	// Check common image extensions
	for _, ext := range []string{".jpg", ".jpeg", ".png", ".gif", ".webp"} {
		if strings.Contains(url, ext) {
			return ext
		}
	}

	// Default to jpg
	return ".jpg"
}

// CleanupPartialDownloads removes .tmp files from a directory
func CleanupPartialDownloads(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if strings.HasSuffix(entry.Name(), ".tmp") {
			path := filepath.Join(dir, entry.Name())
			os.Remove(path) // Ignore errors
		}
	}

	return nil
}
