package repository

import (
	"context"
	"errors"
	"fmt"

	"blindtag_api/internal/database"
	"blindtag_api/internal/models"
)

type ScoreRepository struct {
	db *database.DB
}

func NewScoreRepository(db *database.DB) *ScoreRepository {
	return &ScoreRepository{db: db}
}

// InsertScore records a clear score for a stage.
func (r *ScoreRepository) InsertScore(ctx context.Context, req *models.SubmitScoreRequest) (*models.Score, error) {
	if r.db.Pool == nil {
		return nil, errors.New("database pool not initialized")
	}

	query := `
		INSERT INTO l_game_score (user_id, stage_id, elapsed_time, failed_checks, room_moves, final_score, rank, created_at)
		VALUES (
			(SELECT id FROM d_user_account WHERE id::text = $1 OR display_id = $1 OR device_uuid = $1 LIMIT 1),
			$2, $3, $4, $5, $6, $7, NOW()
		)
		RETURNING id, user_id, stage_id, elapsed_time, failed_checks, room_moves, final_score, rank, created_at;
	`

	var score models.Score
	err := r.db.Pool.QueryRow(ctx, query,
		req.UserID,
		req.StageID,
		req.ElapsedTime,
		req.FailedChecks,
		req.RoomMoves,
		req.FinalScore,
		req.Rank,
	).Scan(
		&score.ID,
		&score.UserID,
		&score.StageID,
		&score.ElapsedTime,
		&score.FailedChecks,
		&score.RoomMoves,
		&score.FinalScore,
		&score.Rank,
		&score.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert score: %w", err)
	}

	return &score, nil
}

// GetLeaderboard retrieves top N scores for a stage sorted by final_score DESC, elapsed_time ASC.
func (r *ScoreRepository) GetLeaderboard(ctx context.Context, stageID, limit int) ([]models.LeaderboardItem, error) {
	if r.db.Pool == nil {
		return nil, errors.New("database pool not initialized")
	}

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	query := `
		SELECT 
			ROW_NUMBER() OVER(ORDER BY s.final_score DESC, s.elapsed_time ASC) AS rank_position,
			u.display_id,
			u.display_name,
			u.profile_icon_index,
			s.stage_id,
			s.final_score,
			s.rank,
			s.elapsed_time,
			s.created_at
		FROM l_game_score s
		INNER JOIN d_user_account u ON s.user_id = u.id
		WHERE s.stage_id = $1
		ORDER BY s.final_score DESC, s.elapsed_time ASC
		LIMIT $2;
	`

	rows, err := r.db.Pool.Query(ctx, query, stageID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query leaderboard: %w", err)
	}
	defer rows.Close()

	var list []models.LeaderboardItem
	for rows.Next() {
		var item models.LeaderboardItem
		if err := rows.Scan(
			&item.RankPosition,
			&item.UserID,
			&item.Nickname,
			&item.IconIndex,
			&item.StageID,
			&item.FinalScore,
			&item.Rank,
			&item.ElapsedTime,
			&item.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan leaderboard item: %w", err)
		}
		list = append(list, item)
	}

	return list, nil
}
