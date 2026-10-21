package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// Crockford's Base32 alphabet (excludes ambiguous chars: 0, O, 1, I, L)
const base32Alphabet = "23456789ABCDEFGHJKMNPQRSTVWXYZ"

// GenerateDisplayID creates a unique runner tag in format BT-XXXXXX (e.g. BT-7K9M2P)
func GenerateDisplayID() (string, error) {
	length := 6
	result := make([]byte, length)
	alphabetLen := big.NewInt(int64(len(base32Alphabet)))

	for i := 0; i < length; i++ {
		idx, err := rand.Int(rand.Reader, alphabetLen)
		if err != nil {
			return "", fmt.Errorf("failed to generate random display ID: %w", err)
		}
		result[i] = base32Alphabet[idx.Int64()]
	}

	return fmt.Sprintf("BT-%s", string(result)), nil
}
