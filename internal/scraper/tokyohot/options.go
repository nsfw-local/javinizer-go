package tokyohot

import (
	"github.com/javinizer/javinizer-go/internal/api/contracts"
	"github.com/javinizer/javinizer-go/internal/scraperutil"
)

func init() {
	scraperutil.RegisterScraperOptions("tokyohot", scraperutil.ScraperOptionsProvider{
		DisplayName: "Tokyo-Hot",
		Options: []any{
			contracts.ScraperOption{
				Key:         "language",
				Label:       "Language",
				Description: "Language for metadata fields",
				Type:        "select",
				Default:     "ja",
				Choices: []contracts.ScraperChoice{
					{Value: "ja", Label: "Japanese"},
					{Value: "en", Label: "English"},
					{Value: "zh", Label: "Chinese"},
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
				Description: "Tokyo-Hot base URL",
				Type:        "string",
			},
		},
	})
}

func intPtr(i int) *int { return &i }
