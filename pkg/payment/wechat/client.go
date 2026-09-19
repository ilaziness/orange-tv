package wechat

import (
	"context"
	"fmt"
	"strings"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"

	"github.com/ilaziness/orange-tv/pkg/payment"
)

func newClient(ctx context.Context, cfg Config) (*core.Client, error) {
	key, err := utils.LoadPrivateKey(strings.TrimSpace(cfg.PrivateKey))
	if err != nil {
		return nil, fmt.Errorf("%w: private_key: %v", payment.ErrInvalidConfig, err)
	}
	opts := []core.ClientOption{
		option.WithWechatPayAutoAuthCipher(cfg.MchID, cfg.MchSerialNo, key, strings.TrimSpace(cfg.APIv3Key)),
	}
	client, err := core.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("%w: wechat client: %v", payment.ErrInvalidConfig, err)
	}
	return client, nil
}

func mapWechatState(state string) string {
	switch strings.ToUpper(strings.TrimSpace(state)) {
	case "SUCCESS":
		return payment.TradeStatusSuccess
	case "REFUND":
		return payment.TradeStatusRefund
	case "CLOSED", "REVOKED", "PAYERROR":
		return payment.TradeStatusClosed
	default:
		return payment.TradeStatusPending
	}
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
