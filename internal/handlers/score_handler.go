package handlers

import (
	"net/http"
	"strconv"

	"blindtag_api/internal/models"
	"blindtag_api/internal/repository"

	"github.com/gin-gonic/gin"
)

type ScoreHandler struct {
	repo *repository.ScoreRepository
}

func NewScoreHandler(repo *repository.ScoreRepository) *ScoreHandler {
	return &ScoreHandler{repo: repo}
}

// SubmitScore saves stage completion receipt.
func (h *ScoreHandler) SubmitScore(c *gin.Context) {
	var req models.SubmitScoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid score submission payload: " + err.Error()})
		return
	}

	score, err := h.repo.InsertScore(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit score: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Game clear score saved successfully",
		"data":    score,
	})
}

// GetLeaderboard returns ranking list for a stage.
func (h *ScoreHandler) GetLeaderboard(c *gin.Context) {
	stageIDStr := c.Param("stage_id")
	stageID, err := strconv.Atoi(stageIDStr)
	if err != nil || stageID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid stage_id parameter"})
		return
	}

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	leaderboard, err := h.repo.GetLeaderboard(c.Request.Context(), stageID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve leaderboard: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stage_id": stageID,
		"count":    len(leaderboard),
		"data":     leaderboard,
	})
}
