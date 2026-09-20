package alipay

import (
	"fmt"
	"strings"
	"time"

	sdk "github.com/smartwalle/alipay/v3"

	"github.com/ilaziness/orange-tv/pkg/payment"
)

func newClient(cfg Config) (*sdk.Client, error) {
	production := !cfg.Sandbox
	client, err := sdk.New(cfg.AppID, strings.TrimSpace(cfg.PrivateKey), production)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", payment.ErrInvalidConfig, err)
	}
	mode := cfg.SignMode
	if mode == "" {
		mode = signModeKey
	}
	switch mode {
	case signModeKey:
		if err := client.LoadAliPayPublicKey(strings.TrimSpace(cfg.AlipayPublicKey)); err != nil {
			return nil, fmt.Errorf("%w: alipay public key: %v", payment.ErrInvalidConfig, err)
		}
	case signModeCert:
		if err := client.LoadAppCertPublicKey(strings.TrimSpace(cfg.AppCert)); err != nil {
			return nil, fmt.Errorf("%w: app cert: %v", payment.ErrInvalidConfig, err)
		}
		if err := client.LoadAliPayRootCert(strings.TrimSpace(cfg.AlipayRootCert)); err != nil {
			return nil, fmt.Errorf("%w: alipay root cert: %v", payment.ErrInvalidConfig, err)
		}
		if err := client.LoadAlipayCertPublicKey(strings.TrimSpace(cfg.AlipayPublicCert)); err != nil {
			return nil, fmt.Errorf("%w: alipay public cert: %v", payment.ErrInvalidConfig, err)
		}
	default:
		return nil, fmt.Errorf("%w: sign_mode must be key or cert", payment.ErrInvalidConfig)
	}
	return client, nil
}

func applyTradeCommon(t *sdk.Trade, cfg Config, req payment.CreateRequest) {
	t.NotifyURL = cfg.NotifyURL
	t.ReturnURL = firstNonEmpty(req.ReturnURL, cfg.ReturnURL)
	t.Subject = req.Subject
	t.OutTradeNo = req.OutTradeNo
	t.TotalAmount = payment.FenToYuan(req.AmountFen)
	if !req.ExpireAt.IsZero() {
		t.TimeExpire = req.ExpireAt.In(time.Local).Format("2006-01-02 15:04:05")
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func mapAlipayStatus(status sdk.TradeStatus) string {
	switch status {
	case sdk.TradeStatusSuccess, sdk.TradeStatusFinished:
		return payment.TradeStatusSuccess
	case sdk.TradeStatusClosed:
		return payment.TradeStatusClosed
	case sdk.TradeStatusWaitBuyerPay:
		return payment.TradeStatusPending
	default:
		if strings.Contains(strings.ToLower(string(status)), "refund") {
			return payment.TradeStatusRefund
		}
		return payment.TradeStatusPending
	}
}
