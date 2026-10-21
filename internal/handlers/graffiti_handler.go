package handlers

import (
	"net/http"
	"strconv"

	"blindtag_api/internal/models"
	"blindtag_api/internal/repository"

	"github.com/gin-gonic/gin"
)

type GraffitiHandler struct {
	repo *repository.GraffitiRepository
}

func NewGraffitiHandler(repo *repository.GraffitiRepository) *GraffitiHandler {
	return &GraffitiHandler{repo: repo}
}

// CreateGraffiti saves new graffiti painted on a room wall.
func (h *GraffitiHandler) CreateGraffiti(c *gin.Context) {
	var req models.CreateGraffitiRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid graffiti payload: " + err.Error()})
		return
	}

	// Auto-inject authenticated user credentials from JWT context
	if tokenUID, exists := c.Get("uid"); exists && tokenUID.(string) != "" {
		req.UserID = tokenUID.(string)
	}
	if tokenDisplayID, exists := c.Get("display_id"); exists && tokenDisplayID.(string) != "" {
		req.AuthorDisplayID = tokenDisplayID.(string)
	}

	if req.UserID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: valid user token required"})
		return
	}

	if req.AuthorName == "" {
		req.AuthorName = "러너_" + req.AuthorDisplayID
	}

	graffiti, err := h.repo.InsertGraffiti(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save graffiti: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Graffiti recorded successfully",
		"data":    graffiti,
	})
}

// GetGraffitis retrieves graffitis by stage and optional room coordinate.
func (h *GraffitiHandler) GetGraffitis(c *gin.Context) {
	stageIDStr := c.Query("stage_id")
	if stageIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'stage_id' is required"})
		return
	}
	stageID, err := strconv.Atoi(stageIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'stage_id'"})
		return
	}

	roomXStr := c.Query("room_x")
	roomYStr := c.Query("room_y")

	if roomXStr != "" && roomYStr != "" {
		roomX, errX := strconv.Atoi(roomXStr)
		roomY, errY := strconv.Atoi(roomYStr)
		if errX != nil || errY != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room_x or room_y"})
			return
		}

		graffitis, err := h.repo.GetGraffitisByRoom(c.Request.Context(), stageID, roomX, roomY)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve room graffitis: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"stage_id": stageID,
			"room_x":   roomX,
			"room_y":   roomY,
			"count":    len(graffitis),
			"data":     graffitis,
		})
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{"error": "Please provide both 'room_x' and 'room_y' query parameters"})
}

// DeleteGraffiti erases user's graffiti.
func (h *GraffitiHandler) DeleteGraffiti(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Graffiti ID parameter is required"})
		return
	}

	userID := ""
	if tokenUID, exists := c.Get("uid"); exists && tokenUID.(string) != "" {
		userID = tokenUID.(string)
	}

	if userID == "" {
		var req models.DeleteGraffitiRequest
		if err := c.ShouldBindJSON(&req); err == nil && req.UserID != "" {
			userID = req.UserID
		}
	}

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: user identity required to verify ownership"})
		return
	}

	if err := h.repo.DeleteGraffiti(c.Request.Context(), id, userID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Graffiti erased successfully"})
}
