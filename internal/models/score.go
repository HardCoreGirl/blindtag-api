package models

import "time"

// Score represents a cleared game run score receipt.
type Score struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	Nickname     string    `json:"nickname,omitempty"`
	IconIndex    int       `json:"profile_icon_index,omitempty"`
	StageID      int       `json:"stage_id"`
	ElapsedTime  float32   `json:"elapsed_time"`
	FailedChecks int       `json:"failed_checks"`
	RoomMoves    int       `json:"room_moves"`
	FinalScore   int       `json:"final_score"`
	Rank         string    `json:"rank"`
	CreatedAt    time.Time `json:"created_at"`
}

// SubmitScoreRequest defines request body for saving clear score.
type SubmitScoreRequest struct {
	UserID       string  `json:"user_id" binding:"required"`
	StageID      int     `json:"stage_id" binding:"required,min=1"`
	ElapsedTime  float32 `json:"elapsed_time" binding:"required,min=0"`
	FailedChecks int     `json:"failed_checks" binding:"min=0"`
	RoomMoves    int     `json:"room_moves" binding:"min=0"`
	FinalScore   int     `json:"final_score" binding:"required,min=0"`
	Rank         string  `json:"rank" binding:"required"`
}

// LeaderboardItem represents an entry in stage ranking.
type LeaderboardItem struct {
	RankPosition int       `json:"rank_position"`
	UserID       string    `json:"user_id"`
	Nickname     string    `json:"nickname"`
	IconIndex    int       `json:"profile_icon_index"`
	StageID      int       `json:"stage_id"`
	FinalScore   int       `json:"final_score"`
	Rank         string    `json:"rank"`
	ElapsedTime  float32   `json:"elapsed_time"`
	CreatedAt    time.Time `json:"created_at"`
}
