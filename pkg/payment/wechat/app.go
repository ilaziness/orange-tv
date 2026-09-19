package wechat

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/app"

	"github.com/ilaziness/orange-tv/pkg/payment"
)

func (driver) createApp(ctx context.Context, raw json.RawMessage, req payment.CreateRequest) (*payment.CreateResult, error) {
	cfg, err := readyConfig(raw, payment.ChannelApp)
	if err != nil {
		return nil, err
	}
	client, err := newClient(ctx, cfg)
	if err != nil {
		return nil, err
	}
	svc := app.AppApiService{Client: client}
	prepay := app.PrepayRequest{
		Appid:       core.String(cfg.AppIDApp),
		Mchid:       core.String(cfg.MchID),
		Description: core.String(req.Subject),
		OutTradeNo:  core.String(req.OutTradeNo),
		NotifyUrl:   core.String(cfg.NotifyURL),
		Amount: &app.Amount{
			Currency: core.String("CNY"),
			Total:    core.Int64(req.AmountFen),
		},
	}
	if !req.ExpireAt.IsZero() {
		prepay.TimeExpire = core.Time(req.ExpireAt)
	}
	if req.ClientIP != "" {
		prepay.SceneInfo = &app.SceneInfo{
			PayerClientIp: core.String(req.ClientIP),
		}
	}
	resp, _, err := svc.PrepayWithRequestPayment(ctx, prepay)
	if err != nil {
		return nil, fmt.Errorf("wechat app prepay: %w", err)
	}
	payload, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("wechat app payload: %w", err)
	}
	return &payment.CreateResult{
		Channel:    payment.ChannelApp,
		AppPayload: string(payload),
	}, nil
}
