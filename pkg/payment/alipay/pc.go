package alipay

import (
	"context"
	"encoding/json"
	"fmt"

	sdk "github.com/smartwalle/alipay/v3"

	"github.com/ilaziness/orange-tv/pkg/payment"
)

func (driver) createPC(_ context.Context, raw json.RawMessage, req payment.CreateRequest) (*payment.CreateResult, error) {
	cfg, err := readyConfig(raw, payment.ChannelPCWeb)
	if err != nil {
		return nil, err
	}
	client, err := newClient(cfg)
	if err != nil {
		return nil, err
	}
	param := sdk.TradePagePay{}
	applyTradeCommon(&param.Trade, cfg, req)
	param.ProductCode = "FAST_INSTANT_TRADE_PAY"
	u, err := client.TradePagePay(param)
	if err != nil {
		return nil, fmt.Errorf("alipay page pay: %w", err)
	}
	return &payment.CreateResult{
		Channel: payment.ChannelPCWeb,
		PayURL:  u.String(),
	}, nil
}
