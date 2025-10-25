package r18dev

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/javinizer/javinizer-go/internal/config"
	"github.com/javinizer/javinizer-go/internal/models"
)

const (
	baseURL = "https://r18.dev"
	apiURL  = baseURL + "/videos/vod/movies/detail/-/dvd_id=%s/json"
)

// Scraper implements the R18.dev scraper
type Scraper struct {
	client  *resty.Client
	enabled bool
}

// New creates a new R18.dev scraper
func New(cfg *config.Config) *Scraper {
	client := resty.New()
	client.SetTimeout(30 * time.Second)
	client.SetRetryCount(3)

	// Set user agent
	userAgent := cfg.Scrapers.UserAgent
	if userAgent == "" {
		userAgent = "Javinizer (+https://github.com/javinizer/Javinizer)"
	}
	client.SetHeader("User-Agent", userAgent)

	// Add browser-like headers to help bypass protection
	client.SetHeader("Accept", "application/json, text/html, */*")
	client.SetHeader("Accept-Language", "en-US,en;q=0.9,ja;q=0.8")
	client.SetHeader("Accept-Encoding", "gzip, deflate, br")
	client.SetHeader("Connection", "keep-alive")
	client.SetHeader("Referer", "https://r18.dev/")

	return &Scraper{
		client:  client,
		enabled: cfg.Scrapers.R18Dev.Enabled,
	}
}

// Name returns the scraper identifier
func (s *Scraper) Name() string {
	return "r18dev"
}

// IsEnabled returns whether the scraper is enabled
func (s *Scraper) IsEnabled() bool {
	return s.enabled
}

// GetURL constructs the URL for a given movie ID
func (s *Scraper) GetURL(id string) (string, error) {
	normalized := normalizeID(id)
	return fmt.Sprintf(apiURL, normalized), nil
}

// Search searches for and scrapes metadata for a given movie ID
func (s *Scraper) Search(id string) (*models.ScraperResult, error) {
	url, err := s.GetURL(id)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.R().
		SetHeader("Accept-Encoding", ""). // Disable compression to avoid issues
		Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data from R18.dev: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("R18.dev returned status code %d", resp.StatusCode())
	}

	// Check if response is HTML (404 or error page)
	contentType := resp.Header().Get("Content-Type")
	if strings.Contains(contentType, "text/html") {
		return nil, fmt.Errorf("movie not found on R18.dev (returned HTML)")
	}

	var data R18Response
	if err := json.Unmarshal(resp.Body(), &data); err != nil {
		// Log first 200 chars for debugging
		bodyPreview := string(resp.Body())
		if len(bodyPreview) > 200 {
			bodyPreview = bodyPreview[:200]
		}
		return nil, fmt.Errorf("failed to parse R18.dev response (preview: %s): %w", bodyPreview, err)
	}

	return s.parseResponse(&data, url)
}

// parseResponse converts R18 API response to ScraperResult
func (s *Scraper) parseResponse(data *R18Response, sourceURL string) (*models.ScraperResult, error) {
	// Use DVDID if available, otherwise convert ContentID to ID format
	movieID := data.DVDID
	if movieID == "" && data.ContentID != "" {
		movieID = contentIDToID(data.ContentID)
	}

	result := &models.ScraperResult{
		Source:      s.Name(),
		SourceURL:   sourceURL,
		Language:    "en", // R18.dev provides English metadata
		ID:          movieID,
		ContentID:   data.ContentID,
		Title:       cleanString(data.Title),
		Description: cleanString(data.Description),
		Runtime:     data.Runtime,
	}

	// Parse release date (now in YYYY-MM-DD format)
	if data.ReleaseDate != "" {
		t, err := time.Parse("2006-01-02", data.ReleaseDate)
		if err == nil {
			result.ReleaseDate = &t
		}
	}

	// Parse director (now a simple string)
	result.Director = cleanString(data.Director)

	// Parse maker/studio (now nested objects)
	result.Maker = cleanString(data.Maker.Name)
	result.Label = cleanString(data.Label.Name)

	// Parse series
	if data.Series.Name != "" {
		result.Series = cleanString(data.Series.Name)
	} else if data.SeriesName != "" {
		result.Series = cleanString(data.SeriesName)
	}

	// Parse actresses (now simpler structure)
	result.Actresses = make([]models.ActressInfo, 0, len(data.Actresses))
	for _, actress := range data.Actresses {
		thumbURL := actress.Image
		if thumbURL != "" && !strings.HasPrefix(thumbURL, "http") {
			thumbURL = "https://pics.dmm.co.jp/mono/actjpgs/" + thumbURL
		}

		actressName := cleanString(actress.Name)
		parts := strings.Fields(actressName)
		firstName := ""
		lastName := ""

		if len(parts) > 0 {
			firstName = parts[0]
		}
		if len(parts) > 1 {
			lastName = parts[1]
		}

		result.Actresses = append(result.Actresses, models.ActressInfo{
			FirstName:    firstName,
			LastName:     lastName,
			JapaneseName: actressName, // Use full name as Japanese name
			ThumbURL:     thumbURL,
		})
	}

	// Parse genres (now simple name field)
	result.Genres = make([]string, 0, len(data.Categories))
	for _, category := range data.Categories {
		if category.Name != "" {
			result.Genres = append(result.Genres, cleanString(category.Name))
		}
	}

	// Parse cover image (now nested in Images.JacketImage)
	if data.Images.JacketImage.Large2 != "" {
		result.CoverURL = data.Images.JacketImage.Large2
	} else if data.Images.JacketImage.Large != "" {
		result.CoverURL = strings.TrimSpace(data.Images.JacketImage.Large)
	}

	// Parse screenshots (now in Images.SampleImages)
	if len(data.Images.SampleImages) > 0 {
		result.ScreenshotURL = data.Images.SampleImages
	}

	// Parse trailer (now nested in Sample)
	if data.Sample.High != "" {
		result.TrailerURL = data.Sample.High
	} else if data.Sample.Low != "" {
		result.TrailerURL = data.Sample.Low
	}

	return result, nil
}

// normalizeID normalizes the movie ID for R18.dev API
func normalizeID(id string) string {
	// R18.dev expects IDs in format like "ipx00535" or "ABP00420"
	// Convert "IPX-535" to "ipx00535"
	id = strings.ToLower(id)
	id = strings.ReplaceAll(id, "-", "")

	return id
}

// contentIDToID converts content ID to standard ID format
// Example: "118abw00001" -> "ABW-001", "ipx00535" -> "IPX-535"
func contentIDToID(contentID string) string {
	// Remove any leading digits (e.g., "118abw00001" -> "abw00001")
	re := regexp.MustCompile(`^(\d*)([a-z]+)(\d+)(.*)$`)
	matches := re.FindStringSubmatch(strings.ToLower(contentID))

	if len(matches) > 3 {
		prefix := strings.ToUpper(matches[2])
		number := matches[3]
		suffix := ""
		if len(matches) > 4 {
			suffix = strings.ToUpper(matches[4])
		}

		// Remove leading zeros from number, but format to 3 digits
		numberInt, err := strconv.Atoi(number)
		if err == nil {
			number = fmt.Sprintf("%03d", numberInt)
		}

		return prefix + "-" + number + suffix
	}

	return strings.ToUpper(contentID)
}

// cleanString removes extra whitespace and newlines
func cleanString(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", "")
	// Replace multiple spaces with single space
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	return s
}

// R18Response represents the JSON response from R18.dev API (current format)
type R18Response struct {
	DVDID       string `json:"dvd_id"`
	ContentID   string `json:"content_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ReleaseDate string `json:"release_date"`
	Runtime     int    `json:"runtime"`

	// Director is now a simple string
	Director string `json:"director"`

	// Maker is now a nested object
	Maker struct {
		Name string `json:"name"`
	} `json:"maker"`

	// Label is now a nested object
	Label struct {
		Name string `json:"name"`
	} `json:"label"`

	// Series can be nested object or string
	Series struct {
		Name string `json:"name"`
	} `json:"series"`
	SeriesName string `json:"series_name"` // Fallback

	// Categories with simple name field
	Categories []struct {
		Name string `json:"name"`
	} `json:"categories"`

	// Actresses with simple name field
	Actresses []struct {
		Name  string `json:"name"`
		Image string `json:"image"`
	} `json:"actresses"`

	// Images are now nested differently
	Images struct {
		JacketImage struct {
			Large  string `json:"large"`
			Large2 string `json:"large2"`
		} `json:"jacket_image"`
		SampleImages []string `json:"sample_images"`
	} `json:"images"`

	// Sample/trailer
	Sample struct {
		High string `json:"high"`
		Low  string `json:"low"`
	} `json:"sample"`
}
