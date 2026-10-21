package database

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

// NewDatabase establishes a connection pool to Supabase PostgreSQL using pgxpool.
func NewDatabase(databaseURL string) (*DB, error) {
	if databaseURL == "" {
		log.Println("[Database] ?좑툘 DATABASE_URL is empty! Database operations will fail until configured.")
		return &DB{Pool: nil}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Supabase PgBouncer (Transaction Pooler, port 6543) compatibility:
	// PgBouncer transaction mode switches server connections across transactions.
	// Prepared statement caching causes "ERROR: prepared statement stmtcache_... already exists (SQLSTATE 42P05)".
	// Using QueryExecModeSimpleProtocol disables client-side prepared statement caching.
	poolConfig.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	// Cloud Run serverless connection tuning
	poolConfig.MaxConns = 15
	poolConfig.MinConns = 2
	poolConfig.MaxConnLifetime = 30 * time.Minute
	poolConfig.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Ping test
	if err := pool.Ping(ctx); err != nil {
		log.Printf("[Database] ?좑툘 Supabase DB ping failed: %v (Check connection string or network)", err)
		return &DB{Pool: pool}, nil
	}

	log.Println("[Database] ??Successfully connected to Supabase PostgreSQL pool!")

	db := &DB{Pool: pool}
	// Verify and auto-migrate stage goal columns and game schema
	_ = db.AutoMigrate(ctx)

	return db, nil
}

// AutoMigrate creates any missing columns/tables automatically on startup.
func (d *DB) AutoMigrate(ctx context.Context) error {
	if d.Pool == nil {
		return errors.New("database pool not initialized")
	}

	migrationSQL := `
		-- 1. s_stage_config??紐⑤뱺 ?좎?媛 怨듭쑀?섎뒗 ?곴뎄 異쒓뎄 ?뺣떟 醫뚰몴 異붽?
		ALTER TABLE public.s_stage_config 
		ADD COLUMN IF NOT EXISTS goal_coord_x INT NOT NULL DEFAULT 3,
		ADD COLUMN IF NOT EXISTS goal_coord_y INT NOT NULL DEFAULT 2;

		-- 2. ?ㅽ뀒?댁?蹂?怨듯넻 異쒓뎄 醫뚰몴 湲곕낯媛?吏??(紐⑤뱺 ?щ꼫媛 怨듭쑀?섎뒗 怨좎젙 誘몃줈 異쒓뎄)
		UPDATE public.s_stage_config SET goal_coord_x = 3, goal_coord_y = 2 WHERE stage_id = 1;
		UPDATE public.s_stage_config SET goal_coord_x = 4, goal_coord_y = 5 WHERE stage_id = 2;
		UPDATE public.s_stage_config SET goal_coord_x = 6, goal_coord_y = 7 WHERE stage_id = 3;
	`

	_, err := d.Pool.Exec(ctx, migrationSQL)
	if err != nil {
		log.Printf("[Database] ?좑툘 AutoMigrate warning: %v", err)
		return err
	}

	log.Println("[Database] ??Stage master exit goals verified and ready (s_stage_config).")
	return nil
}

// Ping checks if database connection is alive.
func (d *DB) Ping(ctx context.Context) error {
	if d.Pool == nil {
		return errors.New("database connection pool is not initialized")
	}
	return d.Pool.Ping(ctx)
}

// Close gracefully closes connection pool.
func (d *DB) Close() {
	if d.Pool != nil {
		d.Pool.Close()
		log.Println("[Database] Connection pool closed.")
	}
}