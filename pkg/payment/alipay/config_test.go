package alipay

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

func TestValidateConfigEnabledRequiresAppID(t *testing.T) {
	t.Parallel()
	err := validateConfig(json.RawMessage(`{"enabled":true}`))
	if !errors.Is(err, payment.ErrInvalidConfig) {
		t.Fatalf("got %v", err)
	}
}

func TestEnsureReadyChannelDisabled(t *testing.T) {
	t.Parallel()
	cfg := Config{Enabled: true, AppID: "a", PrivateKey: "k", AlipayPublicKey: "p", PCWebEnabled: false}
	err := ensureReady(cfg, payment.ChannelPCWeb)
	if !errors.Is(err, payment.ErrChannelDisabled) {
		t.Fatalf("got %v", err)
	}
}

func TestParseConfigInvalidJSON(t *testing.T) {
	t.Parallel()
	_, err := parseConfig(json.RawMessage(`{`))
	if !errors.Is(err, payment.ErrInvalidConfig) {
		t.Fatalf("got %v", err)
	}
}
