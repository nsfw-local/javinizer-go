package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/javinizer/javinizer-go/internal/database"
	"github.com/javinizer/javinizer-go/internal/models"
	"gorm.io/gorm"
)

type actressRequest struct {
	DMMID        int    `json:"dmm_id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	JapaneseName string `json:"japanese_name"`
	ThumbURL     string `json:"thumb_url"`
	Aliases      string `json:"aliases"`
}

type actressesResponse struct {
	Actresses []models.Actress `json:"actresses"`
	Count     int              `json:"count"`
	Total     int64            `json:"total"`
	Limit     int              `json:"limit"`
	Offset    int              `json:"offset"`
}

func normalizeActressRequest(req *actressRequest) {
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)
	req.JapaneseName = strings.TrimSpace(req.JapaneseName)
	req.ThumbURL = strings.TrimSpace(req.ThumbURL)
	req.Aliases = strings.TrimSpace(req.Aliases)
}

func validateActressRequest(req *actressRequest) error {
	if req.DMMID < 0 {
		return errors.New("dmm_id must be greater than or equal to 0")
	}
	if req.FirstName == "" && req.JapaneseName == "" {
		return errors.New("either first_name or japanese_name is required")
	}
	return nil
}

func parsePagination(c *gin.Context) (int, int) {
	limit := 50
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
			if limit > 500 {
				limit = 500
			}
		}
	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	return limit, offset
}

func parseSort(c *gin.Context) (string, string) {
	sortBy := strings.TrimSpace(strings.ToLower(c.Query("sort_by")))
	sortOrder := strings.TrimSpace(strings.ToLower(c.Query("sort_order")))

	if sortBy == "" {
		sortBy = "name"
	}
	if sortOrder == "" {
		sortOrder = "asc"
	}
	return sortBy, sortOrder
}

func parseActressID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid actress id"})
		return 0, false
	}
	return uint(id), true
}

// listActresses handles GET /api/v1/actresses.
func listActresses(actressRepo *database.ActressRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit, offset := parsePagination(c)
		query := strings.TrimSpace(c.Query("q"))
		sortBy, sortOrder := parseSort(c)

		var actresses []models.Actress
		var total int64
		var err error

		if query == "" {
			total, err = actressRepo.Count()
			if err != nil {
				c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
				return
			}

			actresses, err = actressRepo.ListSorted(limit, offset, sortBy, sortOrder)
			if err != nil {
				c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
				return
			}
		} else {
			total, err = actressRepo.CountSearch(query)
			if err != nil {
				c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
				return
			}

			actresses, err = actressRepo.SearchPagedSorted(query, limit, offset, sortBy, sortOrder)
			if err != nil {
				c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
				return
			}
		}

		c.JSON(http.StatusOK, actressesResponse{
			Actresses: actresses,
			Count:     len(actresses),
			Total:     total,
			Limit:     limit,
			Offset:    offset,
		})
	}
}

// getActress handles GET /api/v1/actresses/:id.
func getActress(actressRepo *database.ActressRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := parseActressID(c)
		if !ok {
			return
		}

		actress, err := actressRepo.FindByID(id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, ErrorResponse{Error: "actress not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
			return
		}

		c.JSON(http.StatusOK, actress)
	}
}

// createActress handles POST /api/v1/actresses.
func createActress(actressRepo *database.ActressRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req actressRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}

		normalizeActressRequest(&req)
		if err := validateActressRequest(&req); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}

		actress := &models.Actress{
			DMMID:        req.DMMID,
			FirstName:    req.FirstName,
			LastName:     req.LastName,
			JapaneseName: req.JapaneseName,
			ThumbURL:     req.ThumbURL,
			Aliases:      req.Aliases,
		}

		if err := actressRepo.Create(actress); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
			return
		}

		c.JSON(http.StatusCreated, actress)
	}
}

// updateActress handles PUT /api/v1/actresses/:id.
func updateActress(actressRepo *database.ActressRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := parseActressID(c)
		if !ok {
			return
		}

		existing, err := actressRepo.FindByID(id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, ErrorResponse{Error: "actress not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
			return
		}

		var req actressRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}

		normalizeActressRequest(&req)
		if err := validateActressRequest(&req); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}

		existing.DMMID = req.DMMID
		existing.FirstName = req.FirstName
		existing.LastName = req.LastName
		existing.JapaneseName = req.JapaneseName
		existing.ThumbURL = req.ThumbURL
		existing.Aliases = req.Aliases

		if err := actressRepo.Update(existing); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
			return
		}

		c.JSON(http.StatusOK, existing)
	}
}

// deleteActress handles DELETE /api/v1/actresses/:id.
func deleteActress(actressRepo *database.ActressRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := parseActressID(c)
		if !ok {
			return
		}

		existing, err := actressRepo.FindByID(id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, ErrorResponse{Error: "actress not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
			return
		}

		if err := actressRepo.Delete(id); err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "actress deleted", "id": existing.ID})
	}
}

func writeActressMergeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, database.ErrActressMergeInvalidID),
		errors.Is(err, database.ErrActressMergeSameID),
		errors.Is(err, database.ErrActressMergeInvalidField),
		errors.Is(err, database.ErrActressMergeInvalidDecision):
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	case errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "actress not found"})
	case errors.Is(err, database.ErrActressMergeUniqueConstraint):
		c.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}
}

// previewActressMerge handles POST /api/v1/actresses/merge/preview.
// @Summary Preview actress merge
// @Description Build a merge preview and field conflict set for target/source actress IDs.
// @Tags actress
// @Accept json
// @Produce json
// @Param request body ActressMergePreviewRequest true "Merge preview request"
// @Success 200 {object} ActressMergePreviewResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/actresses/merge/preview [post]
func previewActressMerge(actressRepo *database.ActressRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ActressMergePreviewRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}

		preview, err := actressRepo.PreviewMerge(req.TargetID, req.SourceID)
		if err != nil {
			writeActressMergeError(c, err)
			return
		}

		conflicts := make([]ActressMergeConflict, 0, len(preview.Conflicts))
		for _, conflict := range preview.Conflicts {
			conflicts = append(conflicts, ActressMergeConflict{
				Field:             conflict.Field,
				TargetValue:       conflict.TargetValue,
				SourceValue:       conflict.SourceValue,
				DefaultResolution: conflict.DefaultResolution,
			})
		}

		c.JSON(http.StatusOK, ActressMergePreviewResponse{
			Target:             preview.Target,
			Source:             preview.Source,
			ProposedMerged:     preview.ProposedMerged,
			Conflicts:          conflicts,
			DefaultResolutions: preview.DefaultResolutions,
		})
	}
}

// mergeActresses handles POST /api/v1/actresses/merge.
// @Summary Merge duplicated actresses
// @Description Merge a source actress into a target actress with field-level target/source resolutions.
// @Tags actress
// @Accept json
// @Produce json
// @Param request body ActressMergeRequest true "Merge request"
// @Success 200 {object} ActressMergeResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/actresses/merge [post]
func mergeActresses(actressRepo *database.ActressRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ActressMergeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}

		result, err := actressRepo.Merge(req.TargetID, req.SourceID, req.Resolutions)
		if err != nil {
			writeActressMergeError(c, err)
			return
		}

		c.JSON(http.StatusOK, ActressMergeResponse{
			MergedActress:     result.MergedActress,
			MergedFromID:      result.MergedFromID,
			UpdatedMovies:     result.UpdatedMovies,
			ConflictsResolved: result.ConflictsResolved,
			AliasesAdded:      result.AliasesAdded,
		})
	}
}

// searchActresses handles GET /api/v1/actresses/search?q=query
// @Summary Search actresses
// @Description Search for actresses by name (first, last, or Japanese)
// @Tags actress
// @Produce json
// @Param q query string true "Search query"
// @Success 200 {array} models.Actress
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/actresses/search [get]
func searchActresses(actressRepo *database.ActressRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := c.Query("q")

		// Empty query is allowed - returns all actresses
		// Search using GORM with LIKE query for all name fields
		// This searches in first_name, last_name, and japanese_name
		actresses, err := actressRepo.Search(query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, actresses)
	}
}
