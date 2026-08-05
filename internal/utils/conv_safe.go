package utils

import (
	"math"
)

// Safe conversion helpers for narrowing integer types.
//
// These functions saturate (clamp) at the target type's bounds instead of
// silently wrapping, which avoids the integer-overflow issues flagged by gosec
// (G115/G109). They are intended for values that are expected to fit in the
// target type under normal operation (years, durations, counts, IDs, etc.);
// out-of-range inputs are clamped to the nearest representable value rather
// than returning an error, since callers treat these as best-effort fields.

// IntToInt32 converts an int to int32, clamping to the int32 range.
func IntToInt32(i int) int32 {
	if i > math.MaxInt32 {
		return math.MaxInt32
	}
	if i < math.MinInt32 {
		return math.MinInt32
	}
	return int32(i)
}

// IntToUint32 converts an int to uint32, clamping to the uint32 range.
func IntToUint32(i int) uint32 {
	if i < 0 {
		return 0
	}
	// Convert to uint first: uint can hold MaxUint32 on both 32-bit and
	// 64-bit platforms, so the comparison is portable. On 32-bit, int can
	// never exceed MaxUint32 (since MaxInt32 < MaxUint32), so the clamp
	// branch is effectively dead code there.
	u := uint(i)
	if u > math.MaxUint32 {
		return math.MaxUint32
	}
	return uint32(u)
}

// Int32ToUint32 converts an int32 to uint32, clamping negatives to 0.
func Int32ToUint32(i int32) uint32 {
	if i < 0 {
		return 0
	}
	return uint32(i)
}

// Uint32ToInt32 converts a uint32 to int32, clamping to MaxInt32.
func Uint32ToInt32(u uint32) int32 {
	if u > math.MaxInt32 {
		return math.MaxInt32
	}
	return int32(u)
}

// IntToUint64 converts an int to uint64, clamping negatives to 0.
func IntToUint64(i int) uint64 {
	if i < 0 {
		return 0
	}
	return uint64(i)
}

// Uint64ToInt64 converts a uint64 to int64, clamping to MaxInt64.
func Uint64ToInt64(u uint64) int64 {
	if u > math.MaxInt64 {
		return math.MaxInt64
	}
	return int64(u)
}

// Uint64ToInt converts a uint64 to int, clamping to MaxInt.
func Uint64ToInt(u uint64) int {
	if u > math.MaxInt {
		return math.MaxInt
	}
	return int(u)
}

// IntToInt8 converts an int to int8, clamping to the int8 range.
func IntToInt8(i int) int8 {
	if i > math.MaxInt8 {
		return math.MaxInt8
	}
	if i < math.MinInt8 {
		return math.MinInt8
	}
	return int8(i)
}
