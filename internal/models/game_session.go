package models

import "time"

// GameSession represents an active or completed gameplay session with server-side goal coordinates.
type GameSession struct {
	ID           string     `json:"id"`
	UserID       string     `json:"user_id"`
	StageID      int        `json:"stage_id"`
	GoalCoordX   int        `json:"-"` // Hidden from client to prevent memory scanning/decompilation
	GoalCoordY   int        `json:"-"` // Hidden from client
	FailedChecks int        `json:"failed_checks"`
	RoomMoves    int        `json:"room_moves"`
	IsCleared    bool       `json:"is_cleared"`
	FinalScore   *int       `json:"final_score,omitempty"`
	Rank         *string    `json:"rank,omitempty"`
	StartedAt    time.Time  `json:"started_at"`
	EndedAt      *time.Time `json:"ended_at,omitempty"`
}

// StartGameRequest defines input for initiating a new game run.
type StartGameRequest struct {
	StageID int `json:"stage_id" binding:"required,min=1"`
}

// StartGameResponse defines the response after session is created.
// NOTE: Goal coordinates are intentionally omitted for anti-cheat security.
type StartGameResponse struct {
	SessionID string    `json:"session_id"`
	StageID   int       `json:"stage_id"`
	StartedAt time.Time `json:"started_at"`
}

// CheckGoalRequest defines input for verifying whether the player has found the correct exit altar.
type CheckGoalRequest struct {
	SessionID   string  `json:"session_id" binding:"required"`
	RoomX       int     `json:"room_x"`
	RoomY       int     `json:"room_y"`
	RoomMoves   int     `json:"room_moves"`
	ElapsedTime float32 `json:"elapsed_time"`
}

// CheckGoalResponse returns verification result from server.
type CheckGoalResponse struct {
	IsGoal       bool    `json:"is_goal"`
	FailedChecks int     `json:"failed_checks"`
	FinalScore   int     `json:"final_score,omitempty"`
	Rank         string  `json:"rank,omitempty"`
	ElapsedTime  float32 `json:"elapsed_time,omitempty"`
	RoomMoves    int     `json:"room_moves,omitempty"`
	Message      string  `json:"message"`
}
