package utils

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
)

// GenerateNumericID returns a random numeric string of the given length,
// with the leading digit non-zero (e.g. 10-digit: 1000000000 - 9999999999).
func GenerateNumericID(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("invalid numeric id length: %d", length)
	}
	if length > 18 {
		return "", fmt.Errorf("numeric id length too large: %d", length)
	}

	min := int64(1)
	for i := 1; i < length; i++ {
		min *= 10
	}
	max := min*10 - 1

	rangeSize := big.NewInt(max - min + 1)
	n, err := rand.Int(rand.Reader, rangeSize)
	if err != nil {
		return "", fmt.Errorf("generate numeric id: %w", err)
	}
	value := n.Int64() + min
	return fmt.Sprintf("%0*d", length, value), nil
}

// GenerateUniqueNumericID generates a random numeric id and ensures it does not
// already exist by calling existsFn. It makes up to 10 attempts.
func GenerateUniqueNumericID(ctx context.Context, length int, existsFn func(context.Context, string) (bool, error)) (string, error) {
	for i := 0; i < 10; i++ {
		id, err := GenerateNumericID(length)
		if err != nil {
			return "", err
		}
		exists, err := existsFn(ctx, id)
		if err != nil {
			return "", err
		}
		if !exists {
			return id, nil
		}
	}
	return "", errors.New("failed to generate unique numeric id")
}
