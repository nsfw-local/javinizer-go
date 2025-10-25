# Development Guide

Guide for contributing to and developing Javinizer Go.

## Project Structure

```
javinizer-go/
├── cmd/
│   ├── api/              # API server (planned)
│   └── cli/              # CLI application
├── internal/
│   ├── aggregator/       # Metadata aggregation
│   ├── config/           # Configuration management
│   ├── database/         # Database layer (GORM)
│   ├── downloader/       # Media downloads
│   ├── matcher/          # File/ID matching
│   ├── models/           # Data models
│   ├── nfo/              # NFO generation
│   ├── organizer/        # File organization
│   ├── scanner/          # File scanning
│   ├── scraper/          # Scrapers
│   │   ├── dmm/          # DMM/Fanza
│   │   └── r18dev/       # R18.dev
│   └── template/         # Template engine
├── configs/              # Default configuration
├── data/                 # Runtime data
├── docs/                 # Documentation
└── tests/                # Integration tests
```

## Development Setup

### Prerequisites

- Go 1.21+
- Git
- SQLite3 (for DB inspection)

### Setup

```bash
# Clone repository
git clone https://github.com/javinizer/javinizer-go.git
cd javinizer-go

# Install dependencies
go mod download

# Build
go build -o bin/javinizer ./cmd/cli

# Run
./bin/javinizer --help
```

### Running Tests

```bash
# All tests
go test ./...

# With coverage
go test ./... -cover

# Specific package
go test ./internal/matcher

# Verbose
go test ./... -v
```

## Adding a New Scraper

### 1. Create Scraper Package

```go
// internal/scraper/newscraper/newscraper.go
package newscraper

import (
    "github.com/javinizer/javinizer-go/internal/config"
    "github.com/javinizer/javinizer-go/internal/models"
)

type Scraper struct {
    config *config.ScraperConfig
    client *http.Client
}

func New(cfg *config.Config) *Scraper {
    return &Scraper{
        config: &cfg.Scrapers.NewScraper,
        client: &http.Client{Timeout: 30 * time.Second},
    }
}

func (s *Scraper) Name() string {
    return "newscraper"
}

func (s *Scraper) IsEnabled() bool {
    return s.config.Enabled
}

func (s *Scraper) Search(id string) (*models.ScraperResult, error) {
    // Implement scraping logic
    return &models.ScraperResult{
        ID:    id,
        Title: "...",
        // ... other fields
    }, nil
}

func (s *Scraper) GetURL(id string) string {
    return fmt.Sprintf("https://newscraper.com/movie/%s", id)
}
```

### 2. Register Scraper

```go
// cmd/cli/main.go
import "github.com/javinizer/javinizer-go/internal/scraper/newscraper"

registry := models.NewScraperRegistry()
registry.Register(r18dev.New(cfg))
registry.Register(dmm.New(cfg))
registry.Register(newscraper.New(cfg))  // Add here
```

## Building and Releasing

### Build for Current Platform

```bash
go build -o bin/javinizer ./cmd/cli
```

### Cross-Compile

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o bin/javinizer-linux-amd64 ./cmd/cli

# macOS
GOOS=darwin GOARCH=amd64 go build -o bin/javinizer-darwin-amd64 ./cmd/cli

# Windows
GOOS=windows GOARCH=amd64 go build -o bin/javinizer-windows-amd64.exe ./cmd/cli
```

## Contributing

### Workflow

1. Fork the repository
2. Create feature branch: `git checkout -b feature/my-feature`
3. Make changes
4. Run tests: `go test ./...`
5. Commit: `git commit -m "Add my feature"`
6. Push: `git push origin feature/my-feature`
7. Create Pull Request

## Resources

- **Go Documentation**: https://go.dev/doc/
- **GORM Documentation**: https://gorm.io/docs/
- **Cobra Documentation**: https://github.com/spf13/cobra
- **Original Javinizer**: https://github.com/jvlflame/Javinizer

---

**Next**: [Troubleshooting](./10-troubleshooting.md)
