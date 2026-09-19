package alipay

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/ilaziness/orange-tv/pkg/payment"
)

func (driver) Notify(ctx context.Context, raw json.RawMessage, req payment.NotifyRequest) (*payment.NotifyResult, error) {
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
	values, err := url.ParseQuery(string(req.Body))
	if err != nil {
		return nil, fmt.Errorf("%w: notify body: %v", payment.ErrInvalidRequest, err)
	}
	n, err := client.DecodeNotification(ctx, values)
	if err != nil {
		return nil, fmt.Errorf("alipay notify: %w", err)
	}
	amount, err := payment.YuanStringToFen(n.TotalAmount)
	if err != nil {
		return nil, fmt.Errorf("alipay notify amount: %w", err)
	}
	status := mapAlipayStatus(n.TradeStatus)
	if n.RefundStatus != "" {
		status = payment.TradeStatusRefund
	}
	rawJSON, _ := json.Marshal(map[string]string{
		"out_trade_no": n.OutTradeNo,
		"trade_no":     n.TradeNo,
		"trade_status": string(n.TradeStatus),
		"total_amount": n.TotalAmount,
	})
	return &payment.NotifyResult{
		OutTradeNo: n.OutTradeNo,
		TradeNo:    n.TradeNo,
		Status:     status,
		AmountFen:  amount,
		Raw:        rawJSON,
	}, nil
}
