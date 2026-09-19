package wechat

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/ilaziness/orange-tv/pkg/payment"
)

func TestValidateConfigDisabledEmpty(t *testing.T) {
	t.Parallel()
	if err := validateConfig(json.RawMessage(`{}`)); err != nil {
		t.Fatal(err)
	}
}

func TestValidateConfigEnabledRequiresMch(t *testing.T) {
	t.Parallel()
	err := validateConfig(json.RawMessage(`{"enabled":true}`))
	if !errors.Is(err, payment.ErrInvalidConfig) {
		t.Fatalf("got %v", err)
	}
}

func TestEnsureReadyProviderDisabled(t *testing.T) {
	t.Parallel()
	err := ensureReady(Config{}, payment.ChannelApp)
	if !errors.Is(err, payment.ErrProviderDisabled) {
		t.Fatalf("got %v", err)
	}
}

func TestParseConfigInvalidJSON(t *testing.T) {
	t.Parallel()
	_, err := parseConfig(json.RawMessage(`[`))
	if !errors.Is(err, payment.ErrInvalidConfig) {
		t.Fatalf("got %v", err)
	}
}
