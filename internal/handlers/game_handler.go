package handlers

import (
	"crypto/rand"
	"math/big"
	"net/http"

	"blindtag_api/internal/models"
	"blindtag_api/internal/repository"

	"github.com/gin-gonic/gin"
)

type GameHandler struct {
	repo *repository.GameRepository
}

func NewGameHandler(repo *repository.GameRepository) *GameHandler {
	return &GameHandler{repo: repo}
}

// StartGame initiates a new run and securely generates exit coordinates hidden on the server.
func (h *GameHandler) StartGame(c *gin.Context) {
	uid, exists := c.Get("uid")
	if !exists || uid.(string) == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: valid session token required"})
		return
	}

	var req models.StartGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload: " + err.Error()})
		return
	}

	// Determine stage grid specifications
	var gridW, gridH, minDist int
	switch req.StageID {
	case 1:
		gridW, gridH, minDist = 4, 4, 2
	case 2:
		gridW, gridH, minDist = 6, 6, 3
	default:
		gridW, gridH, minDist = 10, 10, 4
	}

	// Calculate candidates with minimum distance from starting room (0, 0)
	type point struct{ x, y int }
	var candidates []point
	for x := 0; x < gridW; x++ {
		for y := 0; y < gridH; y++ {
			dx := x
			if gridW-x < dx {
				dx = gridW - x
			}
			dy := y
			if gridH-y < dy {
				dy = gridH - y
			}
			if dx+dy >= minDist {
				candidates = append(candidates, point{x, y})
			}
		}
	}

	if len(candidates) == 0 {
		candidates = append(candidates, point{gridW / 2, gridH / 2})
	}

	// Securely choose a random exit room
	idxBig, err := rand.Int(rand.Reader, big.NewInt(int64(len(candidates))))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate random goal"})
		return
	}
	chosen := candidates[idxBig.Int64()]

	session, err := h.repo.CreateSession(c.Request.Context(), uid.(string), req.StageID, chosen.x, chosen.y)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create game session: " + err.Error()})
		return
	}

	// Return session ID without exposing goal coordinates
	c.JSON(http.StatusCreated, gin.H{
		"message": "Game session started successfully",
		"data": models.StartGameResponse{
			SessionID: session.ID,
			StageID:   session.StageID,
			StartedAt: session.StartedAt,
		},
	})
}

// CheckGoal verifies altar interaction against secret server coordinates.
func (h *GameHandler) CheckGoal(c *gin.Context) {
	uid, exists := c.Get("uid")
	if !exists || uid.(string) == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: valid session token required"})
		return
	}

	var req models.CheckGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid check payload: " + err.Error()})
		return
	}

	session, err := h.repo.GetActiveSession(c.Request.Context(), req.SessionID, uid.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Active session not found or already completed"})
		return
	}

	// Compare player room coordinates with secret server-side exit
	isGoal := (req.RoomX == session.GoalCoordX && req.RoomY == session.GoalCoordY)

	if !isGoal {
		// Wrong altar: record failure and increment penalty
		failedCount, err := h.repo.RecordFailedCheck(c.Request.Context(), session.ID, req.RoomMoves)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update session: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Wrong altar",
			"data": models.CheckGoalResponse{
				IsGoal:       false,
				FailedChecks: failedCount,
				RoomMoves:    req.RoomMoves,
				ElapsedTime:  req.ElapsedTime,
				Message:      "※ 탈출 실패: 이곳은 출구가 아닙니다. 벽면 낙서를 참고하여 탐색하세요.",
			},
		})
		return
	}

	// Correct altar: Compute server-verified 10,000 pts deduction score
	timePenalty := int(req.ElapsedTime * 10)
	checkPenalty := session.FailedChecks * 500
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

	// Persist clear state in session and write official score receipt to l_game_score
	if err := h.repo.RecordClearedSession(c.Request.Context(), session.ID, uid.(string), session.StageID, req.RoomMoves, req.ElapsedTime, finalScore, rank); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to finalize cleared session: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Goal verified and game cleared",
		"data": models.CheckGoalResponse{
			IsGoal:       true,
			FinalScore:   finalScore,
			Rank:         rank,
			ElapsedTime:  req.ElapsedTime,
			FailedChecks: session.FailedChecks,
			RoomMoves:    req.RoomMoves,
			Message:      "🎉 탈출구 일치! 미로를 성공적으로 탈출했습니다.",
		},
	})
}
