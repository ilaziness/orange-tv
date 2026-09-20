// Package payment is a vendor-agnostic payment facade.
//
// Import drivers once at process start, call New once, and reuse the Client:
//
//	import _ "github.com/ilaziness/orange-tv/pkg/payment/drivers"
//
//	pay := payment.New()
//	pay.Create(ctx, payment.ProviderAlipay, payment.ChannelPCWeb, cfg, req)
//	pay.Query(ctx, payment.ProviderWechat, cfg, queryReq)
package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// Client is the application-facing dispatcher. Create it once with New and inject it.
type Client struct {
	impls map[string]Payment
}

// New snapshots registered vendors. Import payment/drivers before calling New.
func New() *Client {
	return &Client{impls: snapshot()}
}

func (c *Client) Create(ctx context.Context, provider string, channel Channel, cfg json.RawMessage, req CreateRequest) (*CreateResult, error) {
	channel = Channel(strings.ToLower(strings.TrimSpace(string(channel))))
	if channel == "" {
		return nil, fmt.Errorf("%w: channel is required", ErrInvalidRequest)
	}
	req.OutTradeNo = strings.TrimSpace(req.OutTradeNo)
	req.Subject = strings.TrimSpace(req.Subject)
	req.ClientIP = strings.TrimSpace(req.ClientIP)
	req.ReturnURL = strings.TrimSpace(req.ReturnURL)
	if err := validateCreateRequest(req); err != nil {
		return nil, err
	}
	impl, err := c.lookup(provider)
	if err != nil {
		return nil, err
	}
	return impl.Create(ctx, cfg, channel, req)
}

func (c *Client) Query(ctx context.Context, provider string, cfg json.RawMessage, req QueryRequest) (*QueryResult, error) {
	req.OutTradeNo = strings.TrimSpace(req.OutTradeNo)
	req.TradeNo = strings.TrimSpace(req.TradeNo)
	if req.OutTradeNo == "" && req.TradeNo == "" {
		return nil, fmt.Errorf("%w: out_trade_no or trade_no is required", ErrInvalidRequest)
	}
	impl, err := c.lookup(provider)
	if err != nil {
		return nil, err
	}
	return impl.Query(ctx, cfg, req)
}

func (c *Client) Notify(ctx context.Context, provider string, cfg json.RawMessage, req NotifyRequest) (*NotifyResult, error) {
	if len(req.Body) == 0 {
		return nil, fmt.Errorf("%w: empty notify body", ErrInvalidRequest)
	}
	impl, err := c.lookup(provider)
	if err != nil {
		return nil, err
	}
	return impl.Notify(ctx, cfg, req)
}

func (c *Client) ValidateConfig(provider string, cfg json.RawMessage) error {
	impl, err := c.lookup(provider)
	if err != nil {
		return err
	}
	return impl.ValidateConfig(cfg)
}

func (c *Client) Registered(provider string) bool {
	_, err := c.lookup(provider)
	return err == nil
}

func (c *Client) lookup(provider string) (Payment, error) {
	p := normalizeProvider(provider)
	impl := c.impls[p]
	if impl == nil {
		return nil, fmt.Errorf("%w: %s", ErrProviderNotRegistered, p)
	}
	return impl, nil
}

func validateCreateRequest(req CreateRequest) error {
	if req.OutTradeNo == "" {
		return fmt.Errorf("%w: out_trade_no is required", ErrInvalidRequest)
	}
	if req.AmountFen <= 0 {
		return fmt.Errorf("%w: amount must be positive fen", ErrInvalidRequest)
	}
	if req.Subject == "" {
		return fmt.Errorf("%w: subject is required", ErrInvalidRequest)
	}
	return nil
}
