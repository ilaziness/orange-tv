package payment

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// FenToYuan converts integer fen to a two-decimal yuan string.
func FenToYuan(fen int64) string {
	if fen < 0 {
		fen = -fen
		return fmt.Sprintf("-%d.%02d", fen/100, fen%100)
	}
	return fmt.Sprintf("%d.%02d", fen/100, fen%100)
}

// YuanStringToFen parses an amount string like "12.30" into fen.
// Empty input is treated as 0.
func YuanStringToFen(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}
	neg := false
	if strings.HasPrefix(s, "-") {
		neg = true
		s = s[1:]
	}
	if s == "" || strings.Count(s, ".") > 1 {
		return 0, fmt.Errorf("%w: invalid amount", ErrInvalidRequest)
	}
	intPart, fracPart := s, ""
	if i := strings.IndexByte(s, '.'); i >= 0 {
		intPart, fracPart = s[:i], s[i+1:]
		if len(fracPart) == 0 || len(fracPart) > 2 || !allDigits(fracPart) {
			return 0, fmt.Errorf("%w: invalid amount", ErrInvalidRequest)
		}
		if len(fracPart) == 1 {
			fracPart += "0"
		}
	}
	if !allDigits(intPart) {
		return 0, fmt.Errorf("%w: invalid amount", ErrInvalidRequest)
	}
	yuan, err := strconv.ParseInt(intPart, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: invalid amount", ErrInvalidRequest)
	}
	var fen int64
	if fracPart != "" {
		fen, err = strconv.ParseInt(fracPart, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("%w: invalid amount", ErrInvalidRequest)
		}
	}
	if yuan > math.MaxInt64/100 {
		return 0, fmt.Errorf("%w: invalid amount", ErrInvalidRequest)
	}
	total := yuan*100 + fen
	if neg {
		return -total, nil
	}
	return total, nil
}

func allDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}
