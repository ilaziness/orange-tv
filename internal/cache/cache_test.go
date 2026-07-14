package cache

import (
	"errors"
	"testing"
)

func TestCacheErrors(t *testing.T) {
	t.Run("ErrCacheMiss", func(t *testing.T) {
		if !errors.Is(ErrCacheMiss, ErrCacheMiss) {
			t.Error("ErrCacheMiss should be comparable with errors.Is")
		}
	})

	t.Run("ErrCacheUnavailable", func(t *testing.T) {
		if !errors.Is(ErrCacheUnavailable, ErrCacheUnavailable) {
			t.Error("ErrCacheUnavailable should be comparable with errors.Is")
		}
	})

	t.Run("ErrInvalidKey", func(t *testing.T) {
		if !errors.Is(ErrInvalidKey, ErrInvalidKey) {
			t.Error("ErrInvalidKey should be comparable with errors.Is")
		}
	})
}

func TestCacheError(t *testing.T) {
	t.Run("Error without wrapped error", func(t *testing.T) {
		err := &CacheError{
			Code:    "TEST_CODE",
			Message: "test message",
		}
		expected := "cache error [TEST_CODE]: test message"
		if err.Error() != expected {
			t.Errorf("expected %s, got %s", expected, err.Error())
		}
	})

	t.Run("Error with wrapped error", func(t *testing.T) {
		wrappedErr := errors.New("wrapped error")
		err := &CacheError{
			Code:    "TEST_CODE",
			Message: "test message",
			Err:     wrappedErr,
		}
		expected := "cache error [TEST_CODE]: test message: wrapped error"
		if err.Error() != expected {
			t.Errorf("expected %s, got %s", expected, err.Error())
		}
	})

	t.Run("Unwrap", func(t *testing.T) {
		wrappedErr := errors.New("wrapped error")
		err := &CacheError{
			Code:    "TEST_CODE",
			Message: "test message",
			Err:     wrappedErr,
		}
		unwrapped := err.Unwrap()
		if unwrapped != wrappedErr {
			t.Errorf("expected %v, got %v", wrappedErr, unwrapped)
		}
	})
}
