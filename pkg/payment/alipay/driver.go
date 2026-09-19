package alipay

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ilaziness/orange-tv/pkg/payment"
)

func init() {
	payment.Register(payment.ProviderAlipay, driver{})
}

var _ payment.Payment = driver{}

type driver struct{ payment.Nop }

func (d driver) Create(ctx context.Context, cfg json.RawMessage, channel payment.Channel, req payment.CreateRequest) (*payment.CreateResult, error) {
	switch channel {
	case payment.ChannelPCWeb:
		return d.createPC(ctx, cfg, req)
	case payment.ChannelApp:
		return d.createApp(ctx, cfg, req)
	default:
		return nil, fmt.Errorf("%w: %s", payment.ErrChannelNotRegistered, channel)
	}
}

func (driver) ValidateConfig(raw json.RawMessage) error {
	return validateConfig(raw)
}

func readyConfig(raw json.RawMessage, channel payment.Channel) (Config, error) {
	cfg, err := parseConfig(raw)
	if err != nil {
		return Config{}, err
	}
	if err := ensureReady(cfg, channel); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
