package repository

import (
	"context"
	"errors"
	"fmt"

	"blindtag_api/internal/database"
	"blindtag_api/internal/models"
)

type GameRepository struct {
	db *database.DB
}

func NewGameRepository(db *database.DB) *GameRepository {
	return &GameRepository{db: db}
}

// CreateSession invalidates old active sessions for this user and inserts a new game session with hidden exit coordinates.
func (r *GameRepository) CreateSession(ctx context.Context, userID string, stageID, goalX, goalY int) (*models.GameSession, error) {
	if r.db.Pool == nil {
		return nil, errors.New("database pool not initialized")
	}

	// Invalidate any previously running sessions for this user
	_, _ = r.db.Pool.Exec(ctx,
		"UPDATE d_game_session SET is_cleared = true, ended_at = NOW() WHERE user_id = $1 AND is_cleared = false",
		userID,
	)

	query := `
		INSERT INTO d_game_session (user_id, stage_id, goal_coord_x, goal_coord_y, started_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id, user_id, stage_id, goal_coord_x, goal_coord_y, failed_checks, room_moves, is_cleared, started_at;
	`

	var s models.GameSession
	err := r.db.Pool.QueryRow(ctx, query, userID, stageID, goalX, goalY).Scan(
		&s.ID,
		&s.UserID,
		&s.StageID,
		&s.GoalCoordX,
		&s.GoalCoordY,
		&s.FailedChecks,
		&s.RoomMoves,
		&s.IsCleared,
		&s.StartedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create game session: %w", err)
	}

	return &s, nil
}

// GetActiveSession retrieves uncleared game session by session ID and authenticated user ID.
func (r *GameRepository) GetActiveSession(ctx context.Context, sessionID, userID string) (*models.GameSession, error) {
	if r.db.Pool == nil {
		return nil, errors.New("database pool not initialized")
	}

	query := `
		SELECT id, user_id, stage_id, goal_coord_x, goal_coord_y, failed_checks, room_moves, is_cleared, started_at
		FROM d_game_session
		WHERE id::text = $1 AND user_id = $2 AND is_cleared = false
		LIMIT 1;
	`

	var s models.GameSession
	err := r.db.Pool.QueryRow(ctx, query, sessionID, userID).Scan(
		&s.ID,
		&s.UserID,
		&s.StageID,
		&s.GoalCoordX,
		&s.GoalCoordY,
		&s.FailedChecks,
		&s.RoomMoves,
		&s.IsCleared,
		&s.StartedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("active session not found: %w", err)
	}

	return &s, nil
}

// RecordFailedCheck increments failed altar check count and updates player moves.
func (r *GameRepository) RecordFailedCheck(ctx context.Context, sessionID string, roomMoves int) (int, error) {
	if r.db.Pool == nil {
		return 0, errors.New("database pool not initialized")
	}

	query := `
		UPDATE d_game_session
		SET failed_checks = failed_checks + 1, room_moves = $2
		WHERE id::text = $1 AND is_cleared = false
		RETURNING failed_checks;
	`

	var count int
	err := r.db.Pool.QueryRow(ctx, query, sessionID, roomMoves).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to update failed checks: %w", err)
	}

	return count, nil
}

// RecordClearedSession marks session as cleared and writes official receipt to l_game_score.
func (r *GameRepository) RecordClearedSession(ctx context.Context, sessionID, userID string, stageID int, roomMoves int, elapsedTime float32, finalScore int, rank string) error {
	if r.db.Pool == nil {
		return errors.New("database pool not initialized")
	}

	// 1. Update session
	updateQuery := `
		UPDATE d_game_session
		SET is_cleared = true, ended_at = NOW(), final_score = $2, rank = $3, room_moves = $4
		WHERE id::text = $1;
	`
	_, err := r.db.Pool.Exec(ctx, updateQuery, sessionID, finalScore, rank, roomMoves)
	if err != nil {
		return fmt.Errorf("failed to update cleared session: %w", err)
	}

	// 2. Insert verified score into l_game_score for leaderboards
	scoreQuery := `
		INSERT INTO l_game_score (user_id, stage_id, elapsed_time, failed_checks, room_moves, final_score, rank, created_at)
		VALUES (
			$1, $2, $3,
			(SELECT failed_checks FROM d_game_session WHERE id::text = $4 LIMIT 1),
			$5, $6, $7, NOW()
		);
	`
	_, err = r.db.Pool.Exec(ctx, scoreQuery, userID, stageID, elapsedTime, sessionID, roomMoves, finalScore, rank)
	if err != nil {
		return fmt.Errorf("failed to log game score: %w", err)
	}

	return nil
}
