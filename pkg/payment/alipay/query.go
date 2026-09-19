package alipay

import (
	"context"
	"encoding/json"
	"fmt"

	sdk "github.com/smartwalle/alipay/v3"

	"github.com/ilaziness/orange-tv/pkg/payment"
)

func (driver) Query(ctx context.Context, raw json.RawMessage, req payment.QueryRequest) (*payment.QueryResult, error) {
	cfg, err := parseConfig(raw)
	if err != nil {
		return nil, err
	}
	if !cfg.Enabled {
		return nil, payment.ErrProviderDisabled
	}
	client, err := newClient(cfg)
	if err != nil {
		return nil, err
	}
	param := sdk.TradeQuery{
		OutTradeNo: req.OutTradeNo,
		TradeNo:    req.TradeNo,
	}
	rsp, err := client.TradeQuery(ctx, param)
	if err != nil {
		return nil, fmt.Errorf("alipay trade query: %w", err)
	}
	if rsp == nil {
		return nil, fmt.Errorf("alipay trade query: empty response")
	}
	amount, err := payment.YuanStringToFen(rsp.TotalAmount)
	if err != nil {
		return nil, fmt.Errorf("alipay trade query amount: %w", err)
	}
	return &payment.QueryResult{
		OutTradeNo: rsp.OutTradeNo,
		TradeNo:    rsp.TradeNo,
		Status:     mapAlipayStatus(rsp.TradeStatus),
		AmountFen:  amount,
	}, nil
}
