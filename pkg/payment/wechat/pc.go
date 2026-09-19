package wechat

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"

	"github.com/ilaziness/orange-tv/pkg/payment"
)

func (driver) createPC(ctx context.Context, raw json.RawMessage, req payment.CreateRequest) (*payment.CreateResult, error) {
	cfg, err := readyConfig(raw, payment.ChannelPCWeb)
	if err != nil {
		return nil, err
	}
	client, err := newClient(ctx, cfg)
	if err != nil {
		return nil, err
	}
	svc := native.NativeApiService{Client: client}
	prepay := native.PrepayRequest{
		Appid:       core.String(cfg.AppIDWeb),
		Mchid:       core.String(cfg.MchID),
		Description: core.String(req.Subject),
		OutTradeNo:  core.String(req.OutTradeNo),
		NotifyUrl:   core.String(cfg.NotifyURL),
		Amount: &native.Amount{
			Currency: core.String("CNY"),
			Total:    core.Int64(req.AmountFen),
		},
	}
	if !req.ExpireAt.IsZero() {
		prepay.TimeExpire = core.Time(req.ExpireAt)
	}
	if req.ClientIP != "" {
		prepay.SceneInfo = &native.SceneInfo{
			PayerClientIp: core.String(req.ClientIP),
		}
	}
	resp, _, err := svc.Prepay(ctx, prepay)
	if err != nil {
		return nil, fmt.Errorf("wechat native prepay: %w", err)
	}
	codeURL := ""
	if resp != nil && resp.CodeUrl != nil {
		codeURL = *resp.CodeUrl
	}
	return &payment.CreateResult{
		Channel: payment.ChannelPCWeb,
		CodeURL: codeURL,
	}, nil
}
