package nfo

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/javinizer/javinizer-go/internal/models"
)

// ParseResult contains the parsed NFO data and any warnings
type ParseResult struct {
	Movie    *models.Movie
	Warnings []string // Non-fatal parsing issues
	Source   string   // File path for debugging
}

// Maximum NFO file size (1 MB) - prevents memory exhaustion attacks
const maxNFOSize = 1 << 20 // 1 MiB

// ParseNFO parses a Kodi-compatible NFO file into a models.Movie struct
// Uses streaming XML parsing with a size limit to prevent memory exhaustion.
func ParseNFO(filePath string) (*ParseResult, error) {
	// Open file
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read NFO file: %w", err)
	}
	defer f.Close()

	// Limit reader to prevent memory exhaustion on large files
	limited := io.LimitReader(f, maxNFOSize)

	// Parse XML using streaming decoder
	decoder := xml.NewDecoder(limited)
	var nfoMovie Movie
	if err := decoder.Decode(&nfoMovie); err != nil {
		// Check if error is due to size limit
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return nil, fmt.Errorf("NFO file exceeds maximum size of %d bytes", maxNFOSize)
		}
		return nil, fmt.Errorf("failed to parse NFO XML: %w", err)
	}

	// Convert to models.Movie
	movie, warnings := NFOToMovie(&nfoMovie)

	return &ParseResult{
		Movie:    movie,
		Warnings: warnings,
		Source:   filePath,
	}, nil
}

// NFOToMovie converts an NFO Movie struct to a models.Movie
func NFOToMovie(nfo *Movie) (*models.Movie, []string) {
	var warnings []string

	movie := &models.Movie{
		Title:         nfo.Title,
		OriginalTitle: nfo.OriginalTitle,
		Description:   nfo.Plot,
		Director:      nfo.Director,
		Maker:         nfo.Maker,
		Label:         nfo.Label,
		Series:        nfo.Set,
		Runtime:       nfo.Runtime,
		ReleaseYear:   nfo.Year,
	}

	// Extract ID from various sources
	if nfo.ID != "" {
		movie.ID = nfo.ID
	}

	// Extract ContentID from uniqueid elements
	for _, uid := range nfo.UniqueID {
		if uid.Type == "contentid" && uid.Value != "" {
			movie.ContentID = uid.Value
			break
		}
	}

	// If ContentID is still empty, use ID as fallback
	if movie.ContentID == "" && movie.ID != "" {
		movie.ContentID = movie.ID
	}

	// Parse release date (prefer ReleaseDate over Premiered)
	dateStr := nfo.ReleaseDate
	if dateStr == "" {
		dateStr = nfo.Premiered
	}
	if dateStr != "" {
		if parsedDate, err := parseDate(dateStr); err == nil {
			movie.ReleaseDate = &parsedDate
			// Update ReleaseYear from parsed date if not set
			if movie.ReleaseYear == 0 {
				movie.ReleaseYear = parsedDate.Year()
			}
		} else {
			warnings = append(warnings, fmt.Sprintf("failed to parse date %q: %v", dateStr, err))
		}
	}

	// Extract rating
	if len(nfo.Ratings.Rating) > 0 {
		// Use first rating or find default
		var rating *Rating
		for i := range nfo.Ratings.Rating {
			if nfo.Ratings.Rating[i].Default {
				rating = &nfo.Ratings.Rating[i]
				break
			}
		}
		if rating == nil {
			rating = &nfo.Ratings.Rating[0]
		}

		movie.RatingScore = rating.Value
		movie.RatingVotes = rating.Votes
	}

	// Convert actors to actresses
	if len(nfo.Actors) > 0 {
		movie.Actresses = make([]models.Actress, 0, len(nfo.Actors))
		for _, actor := range nfo.Actors {
			actress := parseActorToActress(actor)
			movie.Actresses = append(movie.Actresses, actress)
		}
	}

	// Convert genres
	if len(nfo.Genres) > 0 {
		movie.Genres = make([]models.Genre, 0, len(nfo.Genres))
		genreMap := make(map[string]bool) // For deduplication
		for _, genreName := range nfo.Genres {
			genreName = strings.TrimSpace(genreName)
			if genreName != "" && !genreMap[genreName] {
				movie.Genres = append(movie.Genres, models.Genre{Name: genreName})
				genreMap[genreName] = true
			}
		}
	}

	// Extract cover URL from thumbs
	for _, thumb := range nfo.Thumb {
		if thumb.Aspect == "poster" && thumb.Value != "" {
			movie.CoverURL = thumb.Value
			break
		}
	}
	// Fallback to first thumb if no poster aspect found
	if movie.CoverURL == "" && len(nfo.Thumb) > 0 {
		movie.CoverURL = nfo.Thumb[0].Value
	}

	// Extract screenshot URLs from fanart
	if nfo.Fanart != nil && len(nfo.Fanart.Thumbs) > 0 {
		movie.Screenshots = make([]string, 0, len(nfo.Fanart.Thumbs))
		for _, thumb := range nfo.Fanart.Thumbs {
			if thumb.Value != "" {
				movie.Screenshots = append(movie.Screenshots, thumb.Value)
			}
		}
	}

	// Extract trailer URL
	if nfo.Trailer != "" {
		movie.TrailerURL = nfo.Trailer
	}

	// Extract original filename
	if nfo.OriginalPath != "" {
		movie.OriginalFileName = nfo.OriginalPath
	}

	// Set source info
	movie.SourceName = "nfo"

	return movie, warnings
}

// parseActorToActress converts an NFO Actor to a models.Actress
func parseActorToActress(actor Actor) models.Actress {
	firstName, lastName := splitActorName(actor.Name)

	actress := models.Actress{
		FirstName: firstName,
		LastName:  lastName,
		ThumbURL:  actor.Thumb,
	}

	// Check both Role and Name for Japanese characters
	// Prefer Role if it contains Japanese, otherwise fall back to Name
	if actor.Role != "" && containsJapanese(actor.Role) {
		actress.JapaneseName = actor.Role
	} else if containsJapanese(actor.Name) {
		actress.JapaneseName = actor.Name
	}

	return actress
}

// splitActorName attempts to split a full name into first and last names
// Handles both "FirstName LastName" and "LastName FirstName" formats
func splitActorName(fullName string) (firstName, lastName string) {
	fullName = strings.TrimSpace(fullName)
	if fullName == "" {
		return "", ""
	}

	parts := strings.Fields(fullName)
	if len(parts) == 0 {
		return "", ""
	} else if len(parts) == 1 {
		return parts[0], ""
	} else if len(parts) == 2 {
		// Assume FirstName LastName format (most common in NFO files)
		return parts[0], parts[1]
	} else {
		// Multiple parts: take first as firstName, rest as lastName
		return parts[0], strings.Join(parts[1:], " ")
	}
}

// containsJapanese checks if a string contains Japanese characters
// Uses unicode package for robust detection of Hiragana, Katakana, and Han (Kanji) characters
func containsJapanese(s string) bool {
	for _, r := range s {
		if unicode.In(r, unicode.Hiragana, unicode.Katakana, unicode.Han) {
			return true
		}
	}
	return false
}

// parseDate parses various date formats commonly found in NFO files
func parseDate(dateStr string) (time.Time, error) {
	dateStr = strings.TrimSpace(dateStr)
	if dateStr == "" {
		return time.Time{}, fmt.Errorf("empty date string")
	}

	// Try RFC3339 formats first (most strict, timezone-aware)
	if t, err := time.Parse(time.RFC3339, dateStr); err == nil {
		return t, nil
	}
	if t, err := time.Parse(time.RFC3339Nano, dateStr); err == nil {
		return t, nil
	}

	// Try common formats with timezone support
	formats := []string{
		"2006-01-02",                // YYYY-MM-DD (ISO 8601, most common)
		"2006/01/02",                // YYYY/MM/DD
		"2006-01-02 15:04:05",       // YYYY-MM-DD HH:MM:SS
		"2006-01-02T15:04:05Z",      // ISO 8601 with time
		"2006-01-02T15:04:05Z07:00", // ISO 8601 with timezone offset
		"2006-01-02T15:04:05-07:00", // ISO 8601 with negative offset
		"02-01-2006",                // DD-MM-YYYY
		"01/02/2006",                // MM/DD/YYYY (ambiguous, US format)
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date %q with known formats", dateStr)
}
