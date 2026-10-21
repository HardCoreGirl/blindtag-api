package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"blindtag_api/internal/database"
	"blindtag_api/internal/models"
	"blindtag_api/internal/utils"

	"github.com/jackc/pgx/v5"
)

var (
	ErrInvalidSecret = errors.New("invalid or missing client secret token")
)

type UserRepository struct {
	db *database.DB
}

func NewUserRepository(db *database.DB) *UserRepository {
	return &UserRepository{db: db}
}

// LoginOrRegisterGuest verifies secret token for existing devices, or registers new account with newly generated secret and Runner No.
func (r *UserRepository) LoginOrRegisterGuest(ctx context.Context, deviceUUID, clientSecret string) (*models.User, string, error) {
	if r.db.Pool == nil {
		return nil, "", errors.New("database pool not initialized")
	}
	if deviceUUID == "" {
		return nil, "", errors.New("device_uuid cannot be empty")
	}

	// 1. Check if user already exists
	selectQuery := `
		SELECT id, device_uuid, secret_hash, display_id, display_name, profile_icon_index, created_at, updated_at
		FROM d_user_account
		WHERE device_uuid = $1
		LIMIT 1;
	`

	var user models.User
	var secretHash string
	err := r.db.Pool.QueryRow(ctx, selectQuery, deviceUUID).Scan(
		&user.ID,
		&user.DeviceUUID,
		&secretHash,
		&user.DisplayID,
		&user.DisplayName,
		&user.ProfileIconIndex,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == nil {
		// Existing account found: verify client_secret
		if clientSecret == "" || !utils.VerifySecret(clientSecret, secretHash) {
			return nil, "", ErrInvalidSecret
		}

		// Update last login timestamp
		_, _ = r.db.Pool.Exec(ctx, "UPDATE d_user_account SET updated_at = NOW() WHERE id = $1", user.ID)
		return &user, "", nil // clientSecret is already held by client
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, "", fmt.Errorf("failed to query user by device_uuid: %w", err)
	}

	// 2. New account registration: Generate secure client_secret & display_id
	newSecret, err := utils.GenerateClientSecret()
	if err != nil {
		return nil, "", err
	}
	newSecretHash := utils.HashSecret(newSecret)

	var displayID string
	for attempt := 0; attempt < 5; attempt++ {
		genID, genErr := utils.GenerateDisplayID()
		if genErr != nil {
			return nil, "", genErr
		}

		var exists bool
		_ = r.db.Pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM d_user_account WHERE display_id = $1)", genID).Scan(&exists)
		if !exists {
			displayID = genID
			break
		}
	}

	if displayID == "" {
		return nil, "", errors.New("failed to generate unique display_id")
	}

	defaultDisplayName := fmt.Sprintf("러너_%s", displayID[3:])
	defaultIconIndex := 0

	insertQuery := `
		INSERT INTO d_user_account (device_uuid, secret_hash, display_id, display_name, profile_icon_index, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, device_uuid, display_id, display_name, profile_icon_index, created_at, updated_at;
	`

	err = r.db.Pool.QueryRow(ctx, insertQuery, deviceUUID, newSecretHash, displayID, defaultDisplayName, defaultIconIndex).Scan(
		&user.ID,
		&user.DeviceUUID,
		&user.DisplayID,
		&user.DisplayName,
		&user.ProfileIconIndex,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, "", fmt.Errorf("failed to insert new guest user: %w", err)
	}

	return &user, newSecret, nil
}

// GetUserByIdentifier retrieves user by internal UID, display_id, or device_uuid.
func (r *UserRepository) GetUserByIdentifier(ctx context.Context, identifier string) (*models.User, error) {
	if r.db.Pool == nil {
		return nil, errors.New("database pool not initialized")
	}

	query := `
		SELECT id, device_uuid, display_id, display_name, profile_icon_index, created_at, updated_at
		FROM d_user_account
		WHERE id::text = $1 OR display_id = $1 OR device_uuid = $1
		LIMIT 1;
	`

	var user models.User
	err := r.db.Pool.QueryRow(ctx, query, identifier).Scan(
		&user.ID,
		&user.DeviceUUID,
		&user.DisplayID,
		&user.DisplayName,
		&user.ProfileIconIndex,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return &user, nil
}

// UpdateProfile updates display name and icon index.
func (r *UserRepository) UpdateProfile(ctx context.Context, identifier, displayName string, iconIndex int) (*models.User, error) {
	if r.db.Pool == nil {
		return nil, errors.New("database pool not initialized")
	}

	query := `
		UPDATE d_user_account
		SET display_name = $2, profile_icon_index = $3, updated_at = $4
		WHERE id::text = $1 OR display_id = $1 OR device_uuid = $1
		RETURNING id, device_uuid, display_id, display_name, profile_icon_index, created_at, updated_at;
	`

	var user models.User
	err := r.db.Pool.QueryRow(ctx, query, identifier, displayName, iconIndex, time.Now()).Scan(
		&user.ID,
		&user.DeviceUUID,
		&user.DisplayID,
		&user.DisplayName,
		&user.ProfileIconIndex,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update user profile: %w", err)
	}

	return &user, nil
}
