package models

// CheckAltarRequest defines the payload sent by client when inspecting an altar.
type CheckAltarRequest struct {
	StageID      int     `json:"stage_id" binding:"required,min=1"`
	RoomX        int     `json:"room_x"`
	RoomY        int     `json:"room_y"`
	ElapsedTime  float32 `json:"elapsed_time"`
	FailedChecks int     `json:"failed_checks"`
	RoomMoves    int     `json:"room_moves"`
}

// CheckAltarResponse defines server verification result against s_stage_config.
type CheckAltarResponse struct {
	IsGoal      bool    `json:"is_goal"`
	FinalScore  int     `json:"final_score,omitempty"`
	Rank        string  `json:"rank,omitempty"`
	ElapsedTime float32 `json:"elapsed_time,omitempty"`
	RoomMoves   int     `json:"room_moves,omitempty"`
	Message     string  `json:"message"`
}
