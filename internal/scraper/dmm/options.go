package dmm

import (
	"github.com/javinizer/javinizer-go/internal/api/contracts"
	"github.com/javinizer/javinizer-go/internal/scraperutil"
)

func init() {
	// Register DMM's display name and options
	scraperutil.RegisterScraperOptions("dmm", scraperutil.ScraperOptionsProvider{
		DisplayName: "DMM/Fanza",
		Options: []any{
			contracts.ScraperOption{
				Key:         "use_browser",
				Label:       "Use Browser",
				Description: "Enable browser automation for this scraper. Requires global browser settings.",
				Type:        "boolean",
			},
			contracts.ScraperOption{
				Key:         "scrape_actress",
				Label:       "Scrape Actress Information",
				Description: "Override global setting: Extract actress names and IDs",
				Type:        "boolean",
			},
		},
	})
}
