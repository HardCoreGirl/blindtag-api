package handlers

import (
	"net/http"

	"blindtag_api/internal/models"
	"blindtag_api/internal/repository"

	"github.com/gin-gonic/gin"
)

type StageHandler struct {
	repo *repository.StageRepository
}

func NewStageHandler(repo *repository.StageRepository) *StageHandler {
	return &StageHandler{repo: repo}
}

// CheckAltar verifies whether the player is at the master exit altar stored in s_stage_config.
// Exit coordinates are shared by all runners in the stage, but hidden on server to prevent client hacking.
func (h *StageHandler) CheckAltar(c *gin.Context) {
	uid, exists := c.Get("uid")
	if !exists || uid.(string) == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: valid session token required"})
		return
	}

	var req models.CheckAltarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid check payload: " + err.Error()})
		return
	}

	// 1. Query fixed exit coordinates for this stage from s_stage_config
	goalX, goalY, err := h.repo.GetStageGoal(c.Request.Context(), req.StageID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query stage goal: " + err.Error()})
		return
	}

	// 2. Compare player's current room with the server-side master exit
	isGoal := (req.RoomX == goalX && req.RoomY == goalY)

	if !isGoal {
		c.JSON(http.StatusOK, gin.H{
			"message": "Wrong altar",
			"data": models.CheckAltarResponse{
				IsGoal:  false,
				Message: "※ 탈출 실패: 이곳은 출구가 아닙니다. 벽면의 낙서를 확인하며 탐색하세요.",
			},
		})
		return
	}

	// 3. Goal matched! Compute server-verified 10,000 pts deduction score
	timePenalty := int(req.ElapsedTime * 10)
	checkPenalty := req.FailedChecks * 500
	movePenalty := req.RoomMoves * 20
	finalScore := 10000 - timePenalty - checkPenalty - movePenalty
	if finalScore < 0 {
		finalScore = 0
	}

	rank := "D"
	switch {
	case finalScore >= 9000:
		rank = "S"
	case finalScore >= 8000:
		rank = "A"
	case finalScore >= 7000:
		rank = "B"
	case finalScore >= 6000:
		rank = "C"
	}

	// 4. Record official score receipt in l_game_score
	if err := h.repo.RecordClearScore(c.Request.Context(), uid.(string), req.StageID, req.ElapsedTime, req.FailedChecks, req.RoomMoves, finalScore, rank); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record clear score: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Altar verified and game cleared",
		"data": models.CheckAltarResponse{
			IsGoal:      true,
			FinalScore:  finalScore,
			Rank:        rank,
			ElapsedTime: req.ElapsedTime,
			RoomMoves:   req.RoomMoves,
			Message:     "🎉 탈출구 일치! 미로를 성공적으로 탈출했습니다.",
		},
	})
}
