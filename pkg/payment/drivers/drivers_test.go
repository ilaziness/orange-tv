package drivers

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/ilaziness/orange-tv/pkg/payment"
)

func TestRegisteredProviders(t *testing.T) {
	pay := payment.New()
	if !pay.Registered(payment.ProviderAlipay) {
		t.Fatal("alipay not registered")
	}
	if !pay.Registered(payment.ProviderWechat) {
		t.Fatal("wechat not registered")
	}
}

func TestValidateConfigEmptyJSON(t *testing.T) {
	pay := payment.New()
	if err := pay.ValidateConfig(payment.ProviderAlipay, json.RawMessage(`{}`)); err != nil {
		t.Fatal(err)
	}
	if err := pay.ValidateConfig(payment.ProviderWechat, json.RawMessage(`{}`)); err != nil {
		t.Fatal(err)
	}
}

func TestUnknownChannel(t *testing.T) {
	pay := payment.New()
	_, err := pay.Create(context.Background(), payment.ProviderAlipay, "h5", json.RawMessage(`{}`), payment.CreateRequest{
		OutTradeNo: "o1",
		AmountFen:  100,
		Subject:    "t",
	})
	if !errors.Is(err, payment.ErrChannelNotRegistered) {
		t.Fatalf("got %v", err)
	}
}

func TestCreateNormalizesChannel(t *testing.T) {
	pay := payment.New()
	_, err := pay.Create(context.Background(), payment.ProviderAlipay, "PC_WEB", json.RawMessage(`{}`), payment.CreateRequest{
		OutTradeNo: "o1",
		AmountFen:  100,
		Subject:    "t",
	})
	if !errors.Is(err, payment.ErrProviderDisabled) {
		t.Fatalf("got %v want disabled after channel normalize", err)
	}
}
