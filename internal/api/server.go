package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/javinizer/javinizer-go/internal/aggregator"
	"github.com/javinizer/javinizer-go/internal/config"
	"github.com/javinizer/javinizer-go/internal/database"
	"github.com/javinizer/javinizer-go/internal/logging"
	"github.com/javinizer/javinizer-go/internal/matcher"
	"github.com/javinizer/javinizer-go/internal/models"
	ws "github.com/javinizer/javinizer-go/internal/websocket"
	"github.com/javinizer/javinizer-go/internal/worker"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var (
	wsHub      *ws.Hub
	wsUpgrader websocket.Upgrader
)

// ServerDependencies holds all dependencies needed to create the API server
type ServerDependencies struct {
	Config      *config.Config
	ConfigFile  string
	Registry    *models.ScraperRegistry
	DB          *database.DB
	Aggregator  *aggregator.Aggregator
	MovieRepo   *database.MovieRepository
	ActressRepo *database.ActressRepository
	Matcher     *matcher.Matcher
	JobQueue    *worker.JobQueue
}

// isOriginAllowed checks if an origin is allowed based on configuration
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	// Empty list = same-origin only (most secure default)
	if len(allowedOrigins) == 0 {
		return false
	}

	// Check each allowed origin
	for _, allowed := range allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}

	return false
}

// NewServer creates and configures the Gin router with all API endpoints
func NewServer(deps *ServerDependencies) *gin.Engine {
	// Initialize job queue and WebSocket hub
	wsHub = ws.NewHub()
	go wsHub.Run()

	// Configure WebSocket upgrader with origin checking from config
	allowedOrigins := deps.Config.API.Security.AllowedOrigins
	wsUpgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			return isOriginAllowed(origin, allowedOrigins)
		},
	}

	// Setup Gin router
	if deps.Config.Logging.Level != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Enable CORS for web UI with origin validation
	router.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check for wildcard in allowed origins
		hasWildcard := false
		for _, allowed := range allowedOrigins {
			if allowed == "*" {
				hasWildcard = true
				break
			}
		}

		if hasWildcard {
			// Wildcard: allow all origins
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		} else if isOriginAllowed(origin, allowedOrigins) {
			// Specific origin allowed
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Serve OpenAPI spec directly for Scalar
	router.StaticFile("/docs/openapi.json", "./docs/swagger/swagger.json")

	// Scalar API documentation (modern, beautiful UI)
	router.GET("/docs", serveScalarDocs)

	// Also provide traditional Swagger UI as fallback
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check endpoint
	router.GET("/health", healthCheck(deps.Registry))

	// WebSocket endpoint for progress updates
	router.GET("/ws/progress", handleWebSocket(wsHub))

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Movie endpoints
		v1.POST("/scrape", scrapeMovie(deps.Registry, deps.Aggregator, deps.MovieRepo, deps.Config))
		v1.GET("/movie/:id", getMovie(deps.MovieRepo))
		v1.GET("/movies", listMovies(deps.MovieRepo))

		// Actress endpoints
		v1.GET("/actresses/search", searchActresses(deps.ActressRepo))

		// System endpoints
		v1.GET("/config", getConfig(deps.Config))
		v1.PUT("/config", updateConfig(deps.Config, deps.ConfigFile))
		v1.GET("/scrapers", getAvailableScrapers(deps.Registry))

		// File endpoints
		v1.GET("/cwd", getCurrentWorkingDirectory())
		v1.POST("/scan", scanDirectory(deps.Matcher, deps.Config))
		v1.POST("/browse", browseDirectory(deps.Config))

		// Batch endpoints
		v1.POST("/batch/scrape", batchScrape(deps.Registry, deps.Aggregator, deps.MovieRepo, deps.Matcher, deps.JobQueue, deps.Config))
		v1.GET("/batch/:id", getBatchJob(deps.JobQueue))
		v1.POST("/batch/:id/cancel", cancelBatchJob(deps.JobQueue))
		v1.PATCH("/batch/:id/movies/:movieId", updateBatchMovie(deps.MovieRepo, deps.JobQueue))
		v1.POST("/batch/:id/movies/:movieId/preview", previewOrganize(deps.JobQueue, deps.Config))
		v1.POST("/batch/:id/organize", organizeJob(deps.Matcher, deps.JobQueue, deps.DB, deps.Config))
	}

	return router
}

// serveScalarDocs serves the Scalar API documentation UI
func serveScalarDocs(c *gin.Context) {
	html := `<!doctype html>
<html>
  <head>
    <title>Javinizer API Documentation</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
  </head>
  <body>
    <script
      id="api-reference"
      data-url="/docs/openapi.json"></script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
  </body>
</html>`
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

// LogServerInfo logs information about the running server
func LogServerInfo(cfg *config.Config) {
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	logging.Infof("Starting API server on %s", addr)
	logging.Infof("📚 API Documentation (Scalar): http://%s/docs", addr)
	logging.Infof("📖 Swagger UI: http://%s/swagger/index.html", addr)
	logging.Infof("🏥 Health check: http://%s/health", addr)
}
