package r18dev

import (
	"github.com/javinizer/javinizer-go/internal/api/contracts"
	"github.com/javinizer/javinizer-go/internal/scraperutil"
)

func init() {
	scraperutil.RegisterScraperOptions("r18dev", scraperutil.ScraperOptionsProvider{
		DisplayName: "R18.dev",
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
		},
	})
}
