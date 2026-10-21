package utils

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type RunnerClaims struct {
	UID       string `json:"uid"`
	DisplayID string `json:"display_id"`
	jwt.RegisteredClaims
}

// GenerateToken creates a signed JWT valid for 30 days.
func GenerateToken(uid, displayID, secretKey string) (string, error) {
	claims := RunnerClaims{
		UID:       uid,
		DisplayID: displayID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "blindtag-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken parses and verifies JWT token signature and expiration.
func ValidateToken(tokenString, secretKey string) (*RunnerClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RunnerClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*RunnerClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid or expired token")
	}

	return claims, nil
}
