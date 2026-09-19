package wechat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/wechatpay-apiv3/wechatpay-go/core/auth/verifiers"
	"github.com/wechatpay-apiv3/wechatpay-go/core/downloader"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"

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
	if _, err := newClient(ctx, cfg); err != nil {
		return nil, err
	}
	handler, err := notify.NewRSANotifyHandler(
		cfg.APIv3Key,
		verifiers.NewSHA256WithRSAVerifier(downloader.MgrInstance().GetCertificateVisitor(cfg.MchID)),
	)
	if err != nil {
		return nil, fmt.Errorf("wechat notify handler: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://notify.local/", bytes.NewReader(req.Body))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", payment.ErrInvalidRequest, err)
	}
	if req.Headers != nil {
		httpReq.Header = req.Headers.Clone()
	}
	content := new(payments.Transaction)
	if _, err := handler.ParseNotifyRequest(ctx, httpReq, content); err != nil {
		return nil, fmt.Errorf("wechat notify: %w", err)
	}
	q := mapTransaction(content)
	return &payment.NotifyResult{
		OutTradeNo: q.OutTradeNo,
		TradeNo:    q.TradeNo,
		Status:     q.Status,
		AmountFen:  q.AmountFen,
		Raw:        transactionJSON(content),
	}, nil
}
