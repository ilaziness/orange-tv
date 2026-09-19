package wechat

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"

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
	client, err := newClient(ctx, cfg)
	if err != nil {
		return nil, err
	}
	svc := native.NativeApiService{Client: client}
	var txn *payments.Transaction
	if req.OutTradeNo != "" {
		txn, _, err = svc.QueryOrderByOutTradeNo(ctx, native.QueryOrderByOutTradeNoRequest{
			OutTradeNo: core.String(req.OutTradeNo),
			Mchid:      core.String(cfg.MchID),
		})
	} else {
		txn, _, err = svc.QueryOrderById(ctx, native.QueryOrderByIdRequest{
			TransactionId: core.String(req.TradeNo),
			Mchid:         core.String(cfg.MchID),
		})
	}
	if err != nil {
		return nil, fmt.Errorf("wechat query: %w", err)
	}
	return mapTransaction(txn), nil
}

func mapTransaction(txn *payments.Transaction) *payment.QueryResult {
	if txn == nil {
		return &payment.QueryResult{Status: payment.TradeStatusPending}
	}
	out := &payment.QueryResult{
		OutTradeNo: deref(txn.OutTradeNo),
		TradeNo:    deref(txn.TransactionId),
		Status:     mapWechatState(deref(txn.TradeState)),
	}
	if txn.Amount != nil && txn.Amount.Total != nil {
		out.AmountFen = *txn.Amount.Total
	}
	return out
}

func transactionJSON(txn *payments.Transaction) json.RawMessage {
	b, err := json.Marshal(txn)
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return b
}
