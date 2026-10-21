package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
)

// GenerateClientSecret generates a secure 64-character random hex token.
func GenerateClientSecret() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate secure client secret: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// HashSecret computes SHA-256 hash of a secret string.
func HashSecret(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}

// VerifySecret compares plain secret against expected hash in constant time (prevents timing attacks).
func VerifySecret(plainSecret, expectedHash string) bool {
	actualHash := HashSecret(plainSecret)
	return subtle.ConstantTimeCompare([]byte(actualHash), []byte(expectedHash)) == 1
}
