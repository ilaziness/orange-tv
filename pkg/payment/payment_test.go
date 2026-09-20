package payment

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestCreateUnregistered(t *testing.T) {
	t.Parallel()
	pay := New()
	_, err := pay.Create(context.Background(), "unknown", ChannelPCWeb, json.RawMessage(`{}`), CreateRequest{
		OutTradeNo: "o1",
		AmountFen:  100,
		Subject:    "t",
	})
	if !errors.Is(err, ErrProviderNotRegistered) {
		t.Fatalf("got %v want ErrProviderNotRegistered", err)
	}
}

func TestCreateInvalidRequest(t *testing.T) {
	t.Parallel()
	pay := New()
	_, err := pay.Create(context.Background(), ProviderAlipay, ChannelPCWeb, json.RawMessage(`{}`), CreateRequest{})
	if !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("got %v want ErrInvalidRequest", err)
	}
}

func TestQueryRequiresID(t *testing.T) {
	t.Parallel()
	pay := New()
	_, err := pay.Query(context.Background(), ProviderAlipay, json.RawMessage(`{}`), QueryRequest{})
	if !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("got %v want ErrInvalidRequest", err)
	}
}

func TestFenToYuan(t *testing.T) {
	t.Parallel()
	if got := FenToYuan(1); got != "0.01" {
		t.Fatalf("got %s", got)
	}
	if got := FenToYuan(1230); got != "12.30" {
		t.Fatalf("got %s", got)
	}
}

func TestYuanStringToFen(t *testing.T) {
	t.Parallel()
	got, err := YuanStringToFen("12.3")
	if err != nil {
		t.Fatal(err)
	}
	if got != 1230 {
		t.Fatalf("got %d", got)
	}
	got, err = YuanStringToFen("")
	if err != nil || got != 0 {
		t.Fatalf("empty: %d %v", got, err)
	}
	if _, err := YuanStringToFen("12abc"); err == nil {
		t.Fatal("expected invalid amount")
	}
	if _, err := YuanStringToFen("12.309"); err == nil {
		t.Fatal("expected invalid amount")
	}
}

func TestRegisteredUnknown(t *testing.T) {
	t.Parallel()
	if New().Registered("no-such-provider") {
		t.Fatal("expected false")
	}
}

func TestValidateConfigUnregistered(t *testing.T) {
	t.Parallel()
	err := New().ValidateConfig("no-such-provider", json.RawMessage(`{}`))
	if !errors.Is(err, ErrProviderNotRegistered) {
		t.Fatalf("got %v", err)
	}
}

func TestNopNotImplemented(t *testing.T) {
	t.Parallel()
	var p Payment = Nop{}
	_, err := p.Create(context.Background(), json.RawMessage(`{}`), ChannelPCWeb, CreateRequest{})
	if !errors.Is(err, ErrNotImplemented) {
		t.Fatalf("Create: %v", err)
	}
	_, err = p.Query(context.Background(), json.RawMessage(`{}`), QueryRequest{})
	if !errors.Is(err, ErrNotImplemented) {
		t.Fatalf("Query: %v", err)
	}
	_, err = p.Notify(context.Background(), json.RawMessage(`{}`), NotifyRequest{Body: []byte("x")})
	if !errors.Is(err, ErrNotImplemented) {
		t.Fatalf("Notify: %v", err)
	}
	if err := p.ValidateConfig(json.RawMessage(`{}`)); !errors.Is(err, ErrNotImplemented) {
		t.Fatalf("ValidateConfig: %v", err)
	}
}

func TestDispatchCreateUsesNopWhenUnimplemented(t *testing.T) {
	provider := "nop-vendor"
	Register(provider, Nop{})
	pay := New()
	_, err := pay.Create(context.Background(), provider, ChannelPCWeb, json.RawMessage(`{}`), CreateRequest{
		OutTradeNo: "o1",
		AmountFen:  100,
		Subject:    "t",
	})
	if !errors.Is(err, ErrNotImplemented) {
		t.Fatalf("got %v", err)
	}
}

func TestNewSnapshotsRegistry(t *testing.T) {
	pay := New()
	Register("late-vendor", Nop{})
	if pay.Registered("late-vendor") {
		t.Fatal("existing client should not see vendors registered after New")
	}
	if !New().Registered("late-vendor") {
		t.Fatal("New after Register should see late-vendor")
	}
}
