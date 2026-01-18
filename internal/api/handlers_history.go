package api

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/javinizer/javinizer-go/internal/database"
	"github.com/javinizer/javinizer-go/internal/logging"
)

// HistoryRecord represents a single history record in API responses
type HistoryRecord struct {
	ID           uint   `json:"id"`
	MovieID      string `json:"movie_id"`
	Operation    string `json:"operation"`
	OriginalPath string `json:"original_path"`
	NewPath      string `json:"new_path"`
	Status       string `json:"status"`
	ErrorMessage string `json:"error_message"`
	Metadata     string `json:"metadata"`
	DryRun       bool   `json:"dry_run"`
	CreatedAt    string `json:"created_at"`
}

// HistoryListResponse is the response for listing history records
type HistoryListResponse struct {
	Records []HistoryRecord `json:"records"`
	Total   int64           `json:"total"`
	Limit   int             `json:"limit"`
	Offset  int             `json:"offset"`
}

// HistoryStats represents aggregated history statistics
type HistoryStats struct {
	Total       int64            `json:"total"`
	Success     int64            `json:"success"`
	Failed      int64            `json:"failed"`
	Reverted    int64            `json:"reverted"`
	ByOperation map[string]int64 `json:"by_operation"`
}

// DeleteHistoryBulkResponse is the response for bulk deletion
type DeleteHistoryBulkResponse struct {
	Deleted int64 `json:"deleted"`
}

// getHistory godoc
// @Summary List history records
// @Description Get a paginated list of history records with optional filtering
// @Tags history
// @Produce json
// @Param limit query int false "Number of records to return (default: 50, max: 500)"
// @Param offset query int false "Number of records to skip (default: 0)"
// @Param operation query string false "Filter by operation type (scrape, organize, download, nfo)"
// @Param status query string false "Filter by status (success, failed, reverted)"
// @Param movie_id query string false "Filter by movie ID"
// @Success 200 {object} HistoryListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/history [get]
func getHistory(historyRepo *database.HistoryRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse pagination params
		limit := 50
		offset := 0

		if limitStr := c.Query("limit"); limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
				limit = l
				if limit > 500 {
					limit = 500
				}
			}
		}

		if offsetStr := c.Query("offset"); offsetStr != "" {
			if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
				offset = o
			}
		}

		// Get filter params
		operation := c.Query("operation")
		status := c.Query("status")
		movieID := c.Query("movie_id")

		var records []HistoryRecord
		var total int64
		var err error

		// Apply filters
		if movieID != "" {
			// Filter by movie ID
			history, findErr := historyRepo.FindByMovieID(movieID)
			if findErr != nil {
				logging.Errorf("Failed to get history by movie ID: %v", findErr)
				c.JSON(500, ErrorResponse{Error: "Failed to retrieve history"})
				return
			}
			total = int64(len(history))

			// Apply pagination manually
			start := offset
			end := offset + limit
			if start > len(history) {
				start = len(history)
			}
			if end > len(history) {
				end = len(history)
			}
			for _, h := range history[start:end] {
				records = append(records, HistoryRecord{
					ID:           h.ID,
					MovieID:      h.MovieID,
					Operation:    h.Operation,
					OriginalPath: h.OriginalPath,
					NewPath:      h.NewPath,
					Status:       h.Status,
					ErrorMessage: h.ErrorMessage,
					Metadata:     h.Metadata,
					DryRun:       h.DryRun,
					CreatedAt:    h.CreatedAt.Format(time.RFC3339),
				})
			}
		} else if operation != "" {
			// Filter by operation
			history, findErr := historyRepo.FindByOperation(operation, 0) // Get all, then paginate
			if findErr != nil {
				logging.Errorf("Failed to get history by operation: %v", findErr)
				c.JSON(500, ErrorResponse{Error: "Failed to retrieve history"})
				return
			}
			total = int64(len(history))

			// Apply pagination manually
			start := offset
			end := offset + limit
			if start > len(history) {
				start = len(history)
			}
			if end > len(history) {
				end = len(history)
			}
			for _, h := range history[start:end] {
				records = append(records, HistoryRecord{
					ID:           h.ID,
					MovieID:      h.MovieID,
					Operation:    h.Operation,
					OriginalPath: h.OriginalPath,
					NewPath:      h.NewPath,
					Status:       h.Status,
					ErrorMessage: h.ErrorMessage,
					Metadata:     h.Metadata,
					DryRun:       h.DryRun,
					CreatedAt:    h.CreatedAt.Format(time.RFC3339),
				})
			}
		} else if status != "" {
			// Filter by status
			history, findErr := historyRepo.FindByStatus(status, 0) // Get all, then paginate
			if findErr != nil {
				logging.Errorf("Failed to get history by status: %v", findErr)
				c.JSON(500, ErrorResponse{Error: "Failed to retrieve history"})
				return
			}
			total = int64(len(history))

			// Apply pagination manually
			start := offset
			end := offset + limit
			if start > len(history) {
				start = len(history)
			}
			if end > len(history) {
				end = len(history)
			}
			for _, h := range history[start:end] {
				records = append(records, HistoryRecord{
					ID:           h.ID,
					MovieID:      h.MovieID,
					Operation:    h.Operation,
					OriginalPath: h.OriginalPath,
					NewPath:      h.NewPath,
					Status:       h.Status,
					ErrorMessage: h.ErrorMessage,
					Metadata:     h.Metadata,
					DryRun:       h.DryRun,
					CreatedAt:    h.CreatedAt.Format(time.RFC3339),
				})
			}
		} else {
			// No filter - get paginated list
			total, err = historyRepo.Count()
			if err != nil {
				logging.Errorf("Failed to count history: %v", err)
				c.JSON(500, ErrorResponse{Error: "Failed to count history"})
				return
			}

			history, findErr := historyRepo.List(limit, offset)
			if findErr != nil {
				logging.Errorf("Failed to list history: %v", findErr)
				c.JSON(500, ErrorResponse{Error: "Failed to retrieve history"})
				return
			}

			for _, h := range history {
				records = append(records, HistoryRecord{
					ID:           h.ID,
					MovieID:      h.MovieID,
					Operation:    h.Operation,
					OriginalPath: h.OriginalPath,
					NewPath:      h.NewPath,
					Status:       h.Status,
					ErrorMessage: h.ErrorMessage,
					Metadata:     h.Metadata,
					DryRun:       h.DryRun,
					CreatedAt:    h.CreatedAt.Format(time.RFC3339),
				})
			}
		}

		// Ensure records is never nil
		if records == nil {
			records = []HistoryRecord{}
		}

		c.JSON(200, HistoryListResponse{
			Records: records,
			Total:   total,
			Limit:   limit,
			Offset:  offset,
		})
	}
}

// getHistoryStats godoc
// @Summary Get history statistics
// @Description Get aggregated statistics about history records
// @Tags history
// @Produce json
// @Success 200 {object} HistoryStats
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/history/stats [get]
func getHistoryStats(historyRepo *database.HistoryRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		total, err := historyRepo.Count()
		if err != nil {
			logging.Errorf("Failed to count history: %v", err)
			c.JSON(500, ErrorResponse{Error: "Failed to get statistics"})
			return
		}

		success, err := historyRepo.CountByStatus("success")
		if err != nil {
			logging.Errorf("Failed to count success history: %v", err)
			c.JSON(500, ErrorResponse{Error: "Failed to get statistics"})
			return
		}

		failed, err := historyRepo.CountByStatus("failed")
		if err != nil {
			logging.Errorf("Failed to count failed history: %v", err)
			c.JSON(500, ErrorResponse{Error: "Failed to get statistics"})
			return
		}

		reverted, err := historyRepo.CountByStatus("reverted")
		if err != nil {
			logging.Errorf("Failed to count reverted history: %v", err)
			c.JSON(500, ErrorResponse{Error: "Failed to get statistics"})
			return
		}

		// Get counts by operation
		byOperation := make(map[string]int64)
		operations := []string{"scrape", "organize", "download", "nfo"}
		for _, op := range operations {
			count, err := historyRepo.CountByOperation(op)
			if err != nil {
				logging.Errorf("Failed to count %s history: %v", op, err)
				continue
			}
			byOperation[op] = count
		}

		c.JSON(200, HistoryStats{
			Total:       total,
			Success:     success,
			Failed:      failed,
			Reverted:    reverted,
			ByOperation: byOperation,
		})
	}
}

// deleteHistory godoc
// @Summary Delete a history record
// @Description Delete a single history record by ID
// @Tags history
// @Produce json
// @Param id path int true "History record ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/history/{id} [delete]
func deleteHistory(historyRepo *database.HistoryRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(400, ErrorResponse{Error: "Invalid history ID"})
			return
		}

		// Check if record exists
		_, err = historyRepo.FindByID(uint(id))
		if err != nil {
			c.JSON(404, ErrorResponse{Error: "History record not found"})
			return
		}

		// Delete the record
		if err := historyRepo.Delete(uint(id)); err != nil {
			logging.Errorf("Failed to delete history record %d: %v", id, err)
			c.JSON(500, ErrorResponse{Error: "Failed to delete history record"})
			return
		}

		c.JSON(200, gin.H{"message": "History record deleted"})
	}
}

// deleteHistoryBulk godoc
// @Summary Delete history records in bulk
// @Description Delete multiple history records based on criteria
// @Tags history
// @Produce json
// @Param older_than_days query int false "Delete records older than N days"
// @Param movie_id query string false "Delete all records for a specific movie"
// @Success 200 {object} DeleteHistoryBulkResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/history [delete]
func deleteHistoryBulk(historyRepo *database.HistoryRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		olderThanDaysStr := c.Query("older_than_days")
		movieID := c.Query("movie_id")

		if olderThanDaysStr == "" && movieID == "" {
			c.JSON(400, ErrorResponse{Error: "Must specify either older_than_days or movie_id"})
			return
		}

		var deleted int64

		if movieID != "" {
			// Count records before deletion
			records, err := historyRepo.FindByMovieID(movieID)
			if err != nil {
				logging.Errorf("Failed to find history by movie ID: %v", err)
				c.JSON(500, ErrorResponse{Error: "Failed to delete history"})
				return
			}
			deleted = int64(len(records))

			// Delete by movie ID
			if err := historyRepo.DeleteByMovieID(movieID); err != nil {
				logging.Errorf("Failed to delete history by movie ID: %v", err)
				c.JSON(500, ErrorResponse{Error: "Failed to delete history"})
				return
			}
		} else if olderThanDaysStr != "" {
			days, err := strconv.Atoi(olderThanDaysStr)
			if err != nil || days < 1 {
				c.JSON(400, ErrorResponse{Error: "Invalid older_than_days value"})
				return
			}

			// Calculate cutoff date
			cutoffDate := time.Now().AddDate(0, 0, -days)

			// Count records before deletion (approximate)
			countBefore, _ := historyRepo.Count()

			// Delete old records
			if err := historyRepo.DeleteOlderThan(cutoffDate); err != nil {
				logging.Errorf("Failed to delete old history: %v", err)
				c.JSON(500, ErrorResponse{Error: "Failed to delete history"})
				return
			}

			// Count records after deletion
			countAfter, _ := historyRepo.Count()
			deleted = countBefore - countAfter
		}

		c.JSON(200, DeleteHistoryBulkResponse{Deleted: deleted})
	}
}
