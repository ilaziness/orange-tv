package alipay

import (
	"context"
	"encoding/json"
	"fmt"

	sdk "github.com/smartwalle/alipay/v3"

	"github.com/ilaziness/orange-tv/pkg/payment"
)

func (driver) createApp(_ context.Context, raw json.RawMessage, req payment.CreateRequest) (*payment.CreateResult, error) {
	cfg, err := readyConfig(raw, payment.ChannelApp)
	if err != nil {
		return nil, err
	}
	client, err := newClient(cfg)
	if err != nil {
		return nil, err
	}
	param := sdk.TradeAppPay{}
	applyTradeCommon(&param.Trade, cfg, req)
	param.ProductCode = "QUICK_MSECURITY_PAY"
	orderStr, err := client.TradeAppPay(param)
	if err != nil {
		return nil, fmt.Errorf("alipay app pay: %w", err)
	}
	return &payment.CreateResult{
		Channel:    payment.ChannelApp,
		AppPayload: orderStr,
	}, nil
}
