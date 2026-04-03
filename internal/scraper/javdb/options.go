package javdb

import (
	"github.com/javinizer/javinizer-go/internal/api/contracts"
	"github.com/javinizer/javinizer-go/internal/scraperutil"
)

func init() {
	scraperutil.RegisterScraperOptions("javdb", scraperutil.ScraperOptionsProvider{
		DisplayName: "JavDB",
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
				Description: "JavDB base URL (leave default unless you need a mirror)",
				Type:        "string",
			},
			contracts.ScraperOption{
				Key:         "use_flaresolverr",
				Label:       "Use FlareSolverr",
				Description: "Route requests through FlareSolverr to bypass Cloudflare protection",
				Type:        "boolean",
			},
		},
	})
}

func intPtr(i int) *int { return &i }
