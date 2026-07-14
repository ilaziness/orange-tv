// Package crypto provides cryptographic utility functions.
package crypto

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword generates a bcrypt hash of the password using the default cost.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bytes), nil
}

// CheckPassword compares a bcrypt hashed password with its possible plaintext equivalent.
func CheckPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// HashPasswordWithCost generates a bcrypt hash with a specific cost.
// Cost must be between bcrypt.MinCost (4) and bcrypt.MaxCost (31).
// If cost is 0, uses bcrypt.DefaultCost (10).
func HashPasswordWithCost(password string, cost int) (string, error) {
	if cost == 0 {
		cost = bcrypt.DefaultCost
	}
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		return "", fmt.Errorf("invalid cost %d: must be between %d and %d", cost, bcrypt.MinCost, bcrypt.MaxCost)
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password with cost %d: %w", cost, err)
	}
	return string(bytes), nil
}
