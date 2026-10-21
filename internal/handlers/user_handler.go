package handlers

import (
	"errors"
	"net/http"

	"blindtag_api/internal/models"
	"blindtag_api/internal/repository"
	"blindtag_api/internal/utils"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	repo      *repository.UserRepository
	jwtSecret string
}

func NewUserHandler(repo *repository.UserRepository, jwtSecret string) *UserHandler {
	return &UserHandler{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

// LoginGuest handles guest authentication using device_uuid + client_secret.
// Issues signed JWT access token upon successful authentication.
func (h *UserHandler) LoginGuest(c *gin.Context) {
	var req models.GuestLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.DeviceUUID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Field 'device_uuid' is required"})
		return
	}

	user, issuedSecret, err := h.repo.LoginOrRegisterGuest(c.Request.Context(), req.DeviceUUID, req.ClientSecret)
	if err != nil {
		if errors.Is(err, repository.ErrInvalidSecret) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: invalid or missing client_secret token"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Issue signed JWT access token (valid for 30 days)
	accessToken, err := utils.GenerateToken(user.ID, user.DisplayID, h.jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate session token: " + err.Error()})
		return
	}

	result := models.LoginResult{
		User:         user,
		ClientSecret: issuedSecret, // Only non-empty on first registration
		AccessToken:  accessToken,
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Guest login successful",
		"data":    result,
	})
}

// GetProfile retrieves player profile by UID, display_id, or device_uuid.
func (h *UserHandler) GetProfile(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User identifier parameter is required"})
		return
	}

	user, err := h.repo.GetUserByIdentifier(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

// UpdateProfile updates player's display name and profile icon (protected by AuthMiddleware).
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User identifier parameter is required"})
		return
	}

	// Verify requester identity matches target user
	tokenDisplayID, _ := c.Get("display_id")
	tokenUID, _ := c.Get("uid")
	if id != tokenDisplayID && id != tokenUID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: cannot modify other user's profile"})
		return
	}

	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body (display_name: 2-16 chars, profile_icon_index: 0-4)"})
		return
	}

	user, err := h.repo.UpdateProfile(c.Request.Context(), id, req.DisplayName, req.ProfileIconIndex)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"data":    user,
	})
}
