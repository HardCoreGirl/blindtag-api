package models

import "time"

// User represents a player account in d_user_account table.
type User struct {
	ID               string    `json:"uid"`
	DeviceUUID       string    `json:"device_uuid,omitempty"`
	DisplayID        string    `json:"display_id"`
	DisplayName      string    `json:"display_name"`
	ProfileIconIndex int       `json:"profile_icon_index"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// GuestLoginRequest defines request payload sent by Unity client.
type GuestLoginRequest struct {
	DeviceUUID   string `json:"device_uuid" binding:"required"`
	ClientSecret string `json:"client_secret,omitempty"` // Omitted on first registration, required on subsequent logins
}

// LoginResult contains authenticated user profile, secret token (if newly issued), and signed JWT access token.
type LoginResult struct {
	User         *User  `json:"user"`
	ClientSecret string `json:"client_secret,omitempty"` // Only returned upon initial account registration
	AccessToken  string `json:"access_token"`            // JWT session token for API authorization
}

// UpdateProfileRequest defines request payload to update display name and profile icon.
type UpdateProfileRequest struct {
	DisplayName      string `json:"display_name" binding:"required,min=2,max=16"`
	ProfileIconIndex int    `json:"profile_icon_index" binding:"min=0,max=4"`
}
