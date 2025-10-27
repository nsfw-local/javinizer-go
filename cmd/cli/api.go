package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/javinizer/javinizer-go/internal/aggregator"
	"github.com/javinizer/javinizer-go/internal/database"
	"github.com/javinizer/javinizer-go/internal/logging"
	"github.com/javinizer/javinizer-go/internal/models"
	"github.com/javinizer/javinizer-go/internal/scraper/dmm"
	"github.com/javinizer/javinizer-go/internal/scraper/r18dev"
	"github.com/spf13/cobra"
)

func newAPICmd() *cobra.Command {
	var (
		host string
		port int
	)

	cmd := &cobra.Command{
		Use:   "api",
		Short: "Start the Javinizer API server",
		Long:  `Start a REST API server for scraping and retrieving JAV metadata`,
		Run: func(cmd *cobra.Command, args []string) {
			runAPI(host, port)
		},
	}

	cmd.Flags().StringVar(&host, "host", "", "Server host address (default from config)")
	cmd.Flags().IntVar(&port, "port", 0, "Server port (default from config)")

	return cmd
}

func runAPI(hostFlag string, portFlag int) {
	// Load configuration
	if err := loadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Override config with flags if provided
	if hostFlag != "" {
		cfg.Server.Host = hostFlag
	}
	if portFlag != 0 {
		cfg.Server.Port = portFlag
	}

	logging.Infof("Loaded configuration from %s", cfgFile)

	// Ensure data directory exists
	dataDir := filepath.Dir(cfg.Database.DSN)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	// Initialize database
	db, err := database.New(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.AutoMigrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	logging.Info("Database initialized and migrated")

	// Initialize scrapers
	registry := models.NewScraperRegistry()
	registry.Register(r18dev.New(cfg))
	registry.Register(dmm.New(cfg))

	logging.Infof("Registered %d scrapers", len(registry.GetAll()))

	// Initialize aggregator
	agg := aggregator.New(cfg)

	// Initialize repositories
	movieRepo := database.NewMovieRepository(db)

	// Setup Gin router
	if cfg.Logging.Level != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"scrapers": func() []string {
				names := []string{}
				for _, s := range registry.GetEnabled() {
					names = append(names, s.Name())
				}
				return names
			}(),
		})
	})

	// Scrape endpoint
	router.POST("/api/v1/scrape", func(c *gin.Context) {
		var req struct {
			ID string `json:"id" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		// Check if already in database
		existing, err := movieRepo.FindByID(req.ID)
		if err == nil && existing != nil {
			c.JSON(200, gin.H{
				"cached": true,
				"movie":  existing,
			})
			return
		}

		// Scrape from sources in priority order
		results := []*models.ScraperResult{}
		errors := []string{}

		for _, scraper := range registry.GetByPriority(cfg.Scrapers.Priority) {
			result, err := scraper.Search(req.ID)
			if err != nil {
				errors = append(errors, fmt.Sprintf("%s: %v", scraper.Name(), err))
				continue
			}
			results = append(results, result)
		}

		if len(results) == 0 {
			c.JSON(404, gin.H{
				"error":  "Movie not found",
				"errors": errors,
			})
			return
		}

		// Aggregate results
		movie, err := agg.Aggregate(results)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		movie.OriginalFileName = req.ID

		// Save to database (upsert: create or update)
		if err := movieRepo.Upsert(movie); err != nil {
			logging.Errorf("Failed to save movie to database: %v", err)
		}

		c.JSON(200, gin.H{
			"cached":       false,
			"movie":        movie,
			"sources_used": len(results),
			"errors":       errors,
		})
	})

	// Get movie by ID
	router.GET("/api/v1/movie/:id", func(c *gin.Context) {
		id := c.Param("id")

		movie, err := movieRepo.FindByID(id)
		if err != nil {
			c.JSON(404, gin.H{"error": "Movie not found"})
			return
		}

		c.JSON(200, gin.H{"movie": movie})
	})

	// List movies
	router.GET("/api/v1/movies", func(c *gin.Context) {
		limit := 20
		offset := 0

		movies, err := movieRepo.List(limit, offset)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{
			"movies": movies,
			"count":  len(movies),
		})
	})

	// Get configuration
	router.GET("/api/v1/config", func(c *gin.Context) {
		c.JSON(200, cfg)
	})

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	logging.Infof("Starting API server on %s", addr)
	logging.Infof("Health check: http://%s/health", addr)
	logging.Infof("API endpoints: http://%s/api/v1/...", addr)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
