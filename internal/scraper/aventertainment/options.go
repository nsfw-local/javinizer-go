package aventertainment

import (
	"github.com/javinizer/javinizer-go/internal/api/contracts"
	"github.com/javinizer/javinizer-go/internal/scraperutil"
)

func init() {
	scraperutil.RegisterScraperOptions("aventertainment", scraperutil.ScraperOptionsProvider{
		DisplayName: "AV Entertainment",
		Options: []any{
			contracts.ScraperOption{
				Key:         "language",
				Label:       "Language",
				Description: "Language for metadata fields",
				Type:        "select",
				Default:     "en",
				Choices: []contracts.ScraperChoice{
					{Value: "en", Label: "English"},
					{Value: "ja", Label: "Japanese"},
				},
			},
			contracts.ScraperOption{
				Key:         "request_delay",
				Label:       "Request Delay",
				Description: "Delay between requests to avoid rate limiting",
				Type:        "number",
				Min:         intPtr(0),
				Max:         intPtr(5000),
				Unit:        "ms",
			},
			contracts.ScraperOption{
				Key:         "base_url",
				Label:       "Base URL",
				Description: "AV Entertainment base URL",
				Type:        "string",
			},
			contracts.ScraperOption{
				Key:         "scrape_bonus_screens",
				Label:       "Scrape bonus screenshots",
				Description: "Append bonus image files to screenshots",
				Type:        "boolean",
			},
		},
	})
}

func intPtr(i int) *int { return &i }
