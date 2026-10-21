package repository

import (
	"context"
	"errors"
	"fmt"

	"blindtag_api/internal/database"
)

type StageRepository struct {
	db *database.DB
}

func NewStageRepository(db *database.DB) *StageRepository {
	return &StageRepository{db: db}
}

// GetStageGoal retrieves the master fixed exit coordinates stored in s_stage_config.
// All runners share this same exit point so wall graffitis guide each other.
func (r *StageRepository) GetStageGoal(ctx context.Context, stageID int) (int, int, error) {
	if r.db.Pool == nil {
		return 0, 0, errors.New("database pool not initialized")
	}

	query := `
		SELECT goal_coord_x, goal_coord_y 
		FROM s_stage_config 
		WHERE stage_id = $1 
		LIMIT 1;
	`

	var goalX, goalY int
	err := r.db.Pool.QueryRow(ctx, query, stageID).Scan(&goalX, &goalY)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to retrieve stage goal from s_stage_config: %w", err)
	}

	return goalX, goalY, nil
}

// RecordClearScore inserts an official verified clear receipt into l_game_score.
func (r *StageRepository) RecordClearScore(ctx context.Context, userID string, stageID int, elapsedTime float32, failedChecks, roomMoves, finalScore int, rank string) error {
	if r.db.Pool == nil {
		return errors.New("database pool not initialized")
	}

	query := `
		INSERT INTO l_game_score (user_id, stage_id, elapsed_time, failed_checks, room_moves, final_score, rank, created_at)
		VALUES (
			(SELECT id FROM d_user_account WHERE id::text = $1 OR display_id = $1 OR device_uuid = $1 LIMIT 1),
			$2, $3, $4, $5, $6, $7, NOW()
		);
	`

	_, err := r.db.Pool.Exec(ctx, query, userID, stageID, elapsedTime, failedChecks, roomMoves, finalScore, rank)
	if err != nil {
		return fmt.Errorf("failed to record clear score to l_game_score: %w", err)
	}

	return nil
}
