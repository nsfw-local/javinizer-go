package dmm

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-resty/resty/v2"
	"github.com/javinizer/javinizer-go/internal/config"
	"github.com/javinizer/javinizer-go/internal/models"
)

const (
	baseURL     = "https://www.dmm.co.jp"
	newBaseURL  = "https://video.dmm.co.jp"
	searchURL   = baseURL + "/search/=/searchstr=%s/"
	digitalURL  = baseURL + "/digital/videoa/-/detail/=/cid=%s/"
	physicalURL = baseURL + "/mono/dvd/-/detail/=/cid=%s/"
	// New URL format (video.dmm.co.jp)
	newDigitalURL = newBaseURL + "/av/content/?id=%s"
)

// Scraper implements the DMM/Fanza scraper
type Scraper struct {
	client         *resty.Client
	enabled        bool
	scrapeActress  bool
}

// New creates a new DMM scraper
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

	// Add browser-like headers to help with scraping
	client.SetHeader("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	client.SetHeader("Accept-Language", "en-US,en;q=0.9,ja;q=0.8")
	client.SetHeader("Accept-Encoding", "gzip, deflate, br")
	client.SetHeader("Connection", "keep-alive")
	client.SetHeader("Upgrade-Insecure-Requests", "1")

	// Set age verification cookie
	client.SetCookie(&http.Cookie{
		Name:   "age_check_done",
		Value:  "1",
		Domain: ".dmm.co.jp",
	})

	// Additional cookies that might help
	client.SetCookie(&http.Cookie{
		Name:   "cklg",
		Value:  "ja",
		Domain: ".dmm.co.jp",
	})

	return &Scraper{
		client:        client,
		enabled:       cfg.Scrapers.DMM.Enabled,
		scrapeActress: cfg.Scrapers.DMM.ScrapeActress,
	}
}

// Name returns the scraper identifier
func (s *Scraper) Name() string {
	return "dmm"
}

// IsEnabled returns whether the scraper is enabled
func (s *Scraper) IsEnabled() bool {
	return s.enabled
}

// GetURL attempts to find the URL for a given movie ID using DMM search
func (s *Scraper) GetURL(id string) (string, error) {
	contentID := normalizeContentID(id)

	// Search DMM using the contentID (e.g., "ipx00535")
	searchURLFormatted := fmt.Sprintf(searchURL, contentID)
	resp, err := s.client.R().Get(searchURLFormatted)
	if err != nil {
		return "", fmt.Errorf("failed to search DMM: %w", err)
	}

	if resp.StatusCode() != 200 {
		return "", fmt.Errorf("DMM search returned status code %d", resp.StatusCode())
	}

	// Parse search results
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(resp.String()))
	if err != nil {
		return "", fmt.Errorf("failed to parse search results: %w", err)
	}

	// Extract all product links from search results
	var matchedURL string

	// Also compute a "clean" version without leading zeros for matching
	// Example: ipx00535 -> ipx535
	cleanSearchID := regexp.MustCompile(`^([a-z]+)0*(\d+.*)$`).ReplaceAllString(contentID, "$1$2")

	// Look for DVD links (/mono/dvd/) first as they're still scrapeable
	doc.Find("a").Each(func(i int, sel *goquery.Selection) {
		href, exists := sel.Attr("href")
		if !exists {
			return
		}

		// Check if it's a DVD product link
		if strings.Contains(href, "/mono/dvd/-/detail/=/cid=") {
			// Extract the CID from the URL
			cidRegex := regexp.MustCompile(`cid=([^/?&]+)`)
			matches := cidRegex.FindStringSubmatch(href)

			if len(matches) > 1 {
				urlCID := matches[1]

				// Clean the CID from URL (remove prepended numbers like "9ipx535" -> "ipx535")
				cleanURLCID := regexp.MustCompile(`^\d*([a-z]+\d+.*)$`).ReplaceAllString(urlCID, "$1")

				// Match against our search ID (both with and without leading zeros)
				if cleanURLCID == cleanSearchID || cleanURLCID == contentID {
					matchedURL = href
					return
				}
			}
		}
	})

	if matchedURL == "" {
		return "", fmt.Errorf("movie not found on DMM search results")
	}

	// Ensure the URL is absolute
	if !strings.HasPrefix(matchedURL, "http") {
		matchedURL = baseURL + matchedURL
	}

	return matchedURL, nil
}

// Search searches for and scrapes metadata for a given movie ID
func (s *Scraper) Search(id string) (*models.ScraperResult, error) {
	url, err := s.GetURL(id)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.R().Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data from DMM: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("DMM returned status code %d", resp.StatusCode())
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(resp.String()))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	return s.parseHTML(doc, url)
}

// parseHTML extracts metadata from DMM HTML
func (s *Scraper) parseHTML(doc *goquery.Document, sourceURL string) (*models.ScraperResult, error) {
	result := &models.ScraperResult{
		Source:    s.Name(),
		SourceURL: sourceURL,
		Language:  "ja", // DMM provides Japanese metadata
	}

	// Extract Content ID from URL
	if cid := extractContentIDFromURL(sourceURL); cid != "" {
		result.ContentID = cid
		result.ID = normalizeID(cid)
	}

	// Extract title
	result.Title = cleanString(doc.Find("h1#title.item").Text())

	// Extract description
	result.Description = s.extractDescription(doc)

	// Extract release date
	if date := s.extractReleaseDate(doc); date != nil {
		result.ReleaseDate = date
	}

	// Extract runtime
	result.Runtime = s.extractRuntime(doc)

	// Extract director
	result.Director = s.extractDirector(doc)

	// Extract maker/studio
	result.Maker = s.extractMaker(doc)

	// Extract label
	result.Label = s.extractLabel(doc)

	// Extract series
	result.Series = s.extractSeries(doc)

	// Extract rating
	result.Rating = s.extractRating(doc)

	// Extract genres
	result.Genres = s.extractGenres(doc)

	// Extract actresses
	result.Actresses = s.extractActresses(doc)

	// Extract cover URL
	result.CoverURL = s.extractCoverURL(doc)

	// Extract screenshots
	result.ScreenshotURL = s.extractScreenshots(doc)

	// Extract trailer URL
	result.TrailerURL = s.extractTrailerURL(doc, sourceURL)

	return result, nil
}

// extractDescription extracts the plot description
func (s *Scraper) extractDescription(doc *goquery.Document) string {
	desc := doc.Find("div.mg-b20.lh4 p.mg-b20").Text()
	if desc == "" {
		desc = doc.Find("div.mg-b20.lh4").Text()
	}
	return cleanString(desc)
}

// extractReleaseDate extracts the release date
func (s *Scraper) extractReleaseDate(doc *goquery.Document) *time.Time {
	dateRegex := regexp.MustCompile(`\d{4}/\d{2}/\d{2}`)
	dateStr := dateRegex.FindString(doc.Text())

	if dateStr != "" {
		t, err := time.Parse("2006/01/02", dateStr)
		if err == nil {
			return &t
		}
	}
	return nil
}

// extractRuntime extracts the runtime in minutes
func (s *Scraper) extractRuntime(doc *goquery.Document) int {
	runtimeRegex := regexp.MustCompile(`(\d{2,3})\s?(?:minutes|分)`)
	matches := runtimeRegex.FindStringSubmatch(doc.Text())

	if len(matches) > 1 {
		runtime, _ := strconv.Atoi(matches[1])
		return runtime
	}
	return 0
}

// extractDirector extracts the director name
func (s *Scraper) extractDirector(doc *goquery.Document) string {
	directorRegex := regexp.MustCompile(`<a.*?href="[^"]*\?director=(\d+)".*?>([^<]+)</a>`)
	html, _ := doc.Html()
	matches := directorRegex.FindStringSubmatch(html)

	if len(matches) > 2 {
		return cleanString(matches[2])
	}
	return ""
}

// extractMaker extracts the studio/maker name
func (s *Scraper) extractMaker(doc *goquery.Document) string {
	// Updated pattern to match PowerShell: supports both ?maker= and /article=maker/id= formats
	makerRegex := regexp.MustCompile(`<a[^>]*href="[^"]*(?:\?maker=|/article=maker/id=)(\d+)[^"]*"[^>]*>([\s\S]*?)</a>`)
	html, _ := doc.Html()
	matches := makerRegex.FindStringSubmatch(html)

	if len(matches) > 2 {
		return cleanString(matches[2])
	}
	return ""
}

// extractLabel extracts the label name
func (s *Scraper) extractLabel(doc *goquery.Document) string {
	// Updated pattern to match PowerShell: supports both ?label= and /article=label/id= formats
	labelRegex := regexp.MustCompile(`<a[^>]*href="[^"]*(?:\?label=|/article=label/id=)(\d+)[^"]*"[^>]*>([\s\S]*?)</a>`)
	html, _ := doc.Html()
	matches := labelRegex.FindStringSubmatch(html)

	if len(matches) > 2 {
		return cleanString(matches[2])
	}
	return ""
}

// extractSeries extracts the series name
func (s *Scraper) extractSeries(doc *goquery.Document) string {
	seriesRegex := regexp.MustCompile(`<a href="(?:/digital/videoa/|(?:/en)?/mono/dvd/)-/list/=/article=series/id=\d*/"[^>]*?>(.*)</a></td>`)
	html, _ := doc.Html()
	matches := seriesRegex.FindStringSubmatch(html)

	if len(matches) > 1 {
		return cleanString(matches[1])
	}
	return ""
}

// extractRating extracts the rating information
func (s *Scraper) extractRating(doc *goquery.Document) *models.Rating {
	ratingRegex := regexp.MustCompile(`<strong>(.*)\s?(?:points|点)</strong>`)
	html, _ := doc.Html()
	matches := ratingRegex.FindStringSubmatch(html)

	if len(matches) > 1 {
		ratingStr := matches[1]

		// Convert word ratings to numbers
		ratingMap := map[string]float64{
			"One":   1.0,
			"Two":   2.0,
			"Three": 3.0,
			"Four":  4.0,
			"Five":  5.0,
		}

		rating := 0.0
		if val, exists := ratingMap[ratingStr]; exists {
			rating = val
		} else {
			rating, _ = strconv.ParseFloat(ratingStr, 64)
		}

		// Multiply by 2 to conform to 1-10 scale
		rating = rating * 2

		// Extract vote count
		votesRegex := regexp.MustCompile(`<p class="d-review__evaluates">.*?<strong>(\d+)</strong>`)
		votesMatches := votesRegex.FindStringSubmatch(html)
		votes := 0
		if len(votesMatches) > 1 {
			votes, _ = strconv.Atoi(votesMatches[1])
		}

		if rating > 0 || votes > 0 {
			return &models.Rating{
				Score: rating,
				Votes: votes,
			}
		}
	}
	return nil
}

// extractGenres extracts genre tags
func (s *Scraper) extractGenres(doc *goquery.Document) []string {
	genres := make([]string, 0)
	html, _ := doc.Html()

	if strings.Contains(html, "Genre:") || strings.Contains(html, "ジャンル：") {
		genreSection := ""
		parts := strings.Split(html, "Genre:")
		if len(parts) < 2 {
			parts = strings.Split(html, "ジャンル：")
		}

		if len(parts) >= 2 {
			endParts := strings.Split(parts[1], "</tr>")
			if len(endParts) > 0 {
				genreSection = endParts[0]
			}
		}

		// Extract genre names from anchor tags
		genreNameRegex := regexp.MustCompile(`>([^<]+)</a>`)
		matches := genreNameRegex.FindAllStringSubmatch(genreSection, -1)

		for _, match := range matches {
			if len(match) > 1 {
				genre := cleanString(match[1])
				if genre != "" {
					genres = append(genres, genre)
				}
			}
		}
	}

	return genres
}

// extractActresses extracts actress information
func (s *Scraper) extractActresses(doc *goquery.Document) []models.ActressInfo {
	actresses := make([]models.ActressInfo, 0)

	// Look for performer block
	actressRegex := regexp.MustCompile(`<a.*?href="[^"]*\?actress=(\d+)".*?>([^<]+)</a>`)
	html, _ := doc.Html()
	matches := actressRegex.FindAllStringSubmatch(html, -1)

	for _, match := range matches {
		if len(match) > 2 {
			actressName := cleanString(match[2])

			// Remove parenthetical content
			actressName = regexp.MustCompile(`\(.*\)|（.*）`).ReplaceAllString(actressName, "")
			actressName = strings.TrimSpace(actressName)

			// Determine if name is Japanese
			isJapanese := regexp.MustCompile(`[\u3040-\u309f]|[\u30a0-\u30ff]|[\uff66-\uff9f]|[\u4e00-\u9faf]`).MatchString(actressName)

			actress := models.ActressInfo{}

			if isJapanese {
				actress.JapaneseName = actressName
			} else {
				// Split English name
				parts := strings.Fields(actressName)
				if len(parts) == 1 {
					actress.FirstName = parts[0]
				} else if len(parts) >= 2 {
					actress.LastName = parts[0]
					actress.FirstName = parts[1]
				}
			}

			actresses = append(actresses, actress)
		}
	}

	return actresses
}

// extractCoverURL extracts the cover image URL
func (s *Scraper) extractCoverURL(doc *goquery.Document) string {
	coverRegex := regexp.MustCompile(`(https://pics\.dmm\.co\.jp/(?:mono/movie/adult|digital/(?:video|amateur))/(.*)/(.*).jpg)`)
	html, _ := doc.Html()
	matches := coverRegex.FindStringSubmatch(html)

	if len(matches) > 1 {
		// Replace 'ps.jpg' with 'pl.jpg' for larger image
		return strings.Replace(matches[1], "ps.jpg", "pl.jpg", 1)
	}
	return ""
}

// extractScreenshots extracts screenshot URLs
func (s *Scraper) extractScreenshots(doc *goquery.Document) []string {
	screenshots := make([]string, 0)

	doc.Find("a[name='sample-image']").Each(func(i int, sel *goquery.Selection) {
		if imgSrc, exists := sel.Find("img").Attr("data-lazy"); exists {
			// Add 'jp-' prefix
			imgSrc = strings.Replace(imgSrc, "-", "jp-", 1)
			screenshots = append(screenshots, imgSrc)
		}
	})

	return screenshots
}

// extractTrailerURL extracts the trailer video URL
func (s *Scraper) extractTrailerURL(doc *goquery.Document, sourceURL string) string {
	// This would require additional requests to iframe URLs
	// Simplified implementation - returns empty for now
	// Full implementation would need to parse the sample player iframe
	return ""
}

// normalizeContentID converts movie ID to DMM content ID format
// Example: "ABP-420" -> "abp00420"
func normalizeContentID(id string) string {
	// Convert to lowercase
	id = strings.ToLower(id)

	// Remove hyphens
	id = strings.ReplaceAll(id, "-", "")

	// Extract prefix and number
	re := regexp.MustCompile(`^(\d*)([a-z]+)(\d+)(.*)$`)
	matches := re.FindStringSubmatch(id)

	if len(matches) > 3 {
		prefix := matches[2]
		number := matches[3]
		suffix := ""
		if len(matches) > 4 {
			suffix = matches[4]
		}

		// Pad number with zeros (DMM uses 5-digit padding)
		paddedNumber := fmt.Sprintf("%05s", number)

		return prefix + paddedNumber + suffix
	}

	return id
}

// normalizeID converts content ID back to standard ID format
// Example: "abp00420" -> "ABP-420"
func normalizeID(contentID string) string {
	re := regexp.MustCompile(`^(\d*)([a-z]+)(\d+)(.*)$`)
	matches := re.FindStringSubmatch(contentID)

	if len(matches) > 3 {
		prefix := strings.ToUpper(matches[2])
		number := matches[3]
		suffix := ""
		if len(matches) > 4 {
			suffix = strings.ToUpper(matches[4])
		}

		// Remove leading zeros
		numberInt, _ := strconv.Atoi(number)
		paddedNumber := fmt.Sprintf("%03d", numberInt)

		return prefix + "-" + paddedNumber + suffix
	}

	return strings.ToUpper(contentID)
}

// extractContentIDFromURL extracts content ID from DMM URL
func extractContentIDFromURL(url string) string {
	re := regexp.MustCompile(`cid=([^/]+)`)
	matches := re.FindStringSubmatch(url)

	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// cleanString removes extra whitespace and newlines
func cleanString(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\t", " ")
	// Replace multiple spaces with single space
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	return s
}
