package repository

import (
	"context"
	"errors"
	"fmt"

	"blindtag_api/internal/database"
	"blindtag_api/internal/models"
)

type GraffitiRepository struct {
	db *database.DB
}

func NewGraffitiRepository(db *database.DB) *GraffitiRepository {
	return &GraffitiRepository{db: db}
}

// InsertGraffiti saves new graffiti painted on a wall.
func (r *GraffitiRepository) InsertGraffiti(ctx context.Context, req *models.CreateGraffitiRequest) (*models.Graffiti, error) {
	if r.db.Pool == nil {
		return nil, errors.New("database pool not initialized")
	}

	query := `
		INSERT INTO d_graffiti_wall (
			user_id, author_display_id, author_name, author_icon_index, stage_id,
			room_coord_x, room_coord_y, normal_x, normal_y, normal_z,
			position_x, position_y, position_z, image_data, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, NOW())
		RETURNING id, user_id, author_display_id, author_name, author_icon_index, stage_id,
		          room_coord_x, room_coord_y, normal_x, normal_y, normal_z,
		          position_x, position_y, position_z, image_data, created_at;
	`

	var g models.Graffiti
	err := r.db.Pool.QueryRow(ctx, query,
		req.UserID,
		req.AuthorDisplayID,
		req.AuthorName,
		req.AuthorIconIndex,
		req.StageID,
		req.RoomCoordX,
		req.RoomCoordY,
		req.NormalX,
		req.NormalY,
		req.NormalZ,
		req.PositionX,
		req.PositionY,
		req.PositionZ,
		req.ImageData,
	).Scan(
		&g.ID,
		&g.UserID,
		&g.AuthorDisplayID,
		&g.AuthorName,
		&g.AuthorIconIndex,
		&g.StageID,
		&g.RoomCoordX,
		&g.RoomCoordY,
		&g.NormalX,
		&g.NormalY,
		&g.NormalZ,
		&g.PositionX,
		&g.PositionY,
		&g.PositionZ,
		&g.ImageData,
		&g.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert graffiti: %w", err)
	}

	return &g, nil
}

// GetGraffitisByRoom retrieves all graffitis in a specific room of a stage.
func (r *GraffitiRepository) GetGraffitisByRoom(ctx context.Context, stageID, roomX, roomY int) ([]models.Graffiti, error) {
	if r.db.Pool == nil {
		return nil, errors.New("database pool not initialized")
	}

	query := `
		SELECT id, user_id, author_display_id, author_name, author_icon_index, stage_id,
		       room_coord_x, room_coord_y, normal_x, normal_y, normal_z,
		       position_x, position_y, position_z, image_data, created_at
		FROM d_graffiti_wall
		WHERE stage_id = $1 AND room_coord_x = $2 AND room_coord_y = $3
		ORDER BY created_at DESC;
	`

	rows, err := r.db.Pool.Query(ctx, query, stageID, roomX, roomY)
	if err != nil {
		return nil, fmt.Errorf("failed to query room graffitis: %w", err)
	}
	defer rows.Close()

	list := make([]models.Graffiti, 0)
	for rows.Next() {
		var g models.Graffiti
		if err := rows.Scan(
			&g.ID,
			&g.UserID,
			&g.AuthorDisplayID,
			&g.AuthorName,
			&g.AuthorIconIndex,
			&g.StageID,
			&g.RoomCoordX,
			&g.RoomCoordY,
			&g.NormalX,
			&g.NormalY,
			&g.NormalZ,
			&g.PositionX,
			&g.PositionY,
			&g.PositionZ,
			&g.ImageData,
			&g.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan graffiti: %w", err)
		}
		list = append(list, g)
	}

	return list, nil
}

// DeleteGraffiti removes graffiti if requested by author.
func (r *GraffitiRepository) DeleteGraffiti(ctx context.Context, id, userID string) error {
	if r.db.Pool == nil {
		return errors.New("database pool not initialized")
	}

	query := `DELETE FROM d_graffiti_wall WHERE id::text = $1 AND user_id = $2;`
	tag, err := r.db.Pool.Exec(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete graffiti: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return errors.New("graffiti not found or unauthorized")
	}

	return nil
}
