package dlgetchu

import (
	"github.com/javinizer/javinizer-go/internal/api/contracts"
	"github.com/javinizer/javinizer-go/internal/scraperutil"
)

func init() {
	scraperutil.RegisterScraperOptions("dlgetchu", scraperutil.ScraperOptionsProvider{
		DisplayName: "DLGetchu",
		Options: []any{
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
				Description: "DLGetchu base URL",
				Type:        "string",
			},
		},
	})
}

func intPtr(i int) *int { return &i }
