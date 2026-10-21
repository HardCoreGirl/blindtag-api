package models

import "time"

// Graffiti represents a drawn mural/decal placed on a wall.
type Graffiti struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	AuthorDisplayID string    `json:"author_display_id"`
	AuthorName      string    `json:"author_name"`
	AuthorIconIndex int       `json:"author_icon_index"`
	StageID         int       `json:"stage_id"`
	RoomCoordX      int       `json:"room_coord_x"`
	RoomCoordY      int       `json:"room_coord_y"`
	NormalX         float32   `json:"normal_x"`
	NormalY         float32   `json:"normal_y"`
	NormalZ         float32   `json:"normal_z"`
	PositionX       float32   `json:"position_x"`
	PositionY       float32   `json:"position_y"`
	PositionZ       float32   `json:"position_z"`
	ImageData       string    `json:"image_data"` // PNG base64
	CreatedAt       time.Time `json:"created_at"`
}

// CreateGraffitiRequest defines request body for saving graffiti.
type CreateGraffitiRequest struct {
	UserID          string  `json:"user_id"`
	AuthorDisplayID string  `json:"author_display_id"`
	AuthorName      string  `json:"author_name"`
	AuthorIconIndex int     `json:"author_icon_index"`
	StageID         int     `json:"stage_id" binding:"required,min=1"`
	RoomCoordX      int     `json:"room_coord_x"`
	RoomCoordY      int     `json:"room_coord_y"`
	NormalX         float32 `json:"normal_x"`
	NormalY         float32 `json:"normal_y"`
	NormalZ         float32 `json:"normal_z"`
	PositionX       float32 `json:"position_x"`
	PositionY       float32 `json:"position_y"`
	PositionZ       float32 `json:"position_z"`
	ImageData       string  `json:"image_data" binding:"required"`
}

// DeleteGraffitiRequest defines query/body params to erase user's graffiti.
type DeleteGraffitiRequest struct {
	UserID string `json:"user_id"`
}
