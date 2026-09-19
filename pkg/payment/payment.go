package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// Client is the shared dispatcher. Business code should hold one instance from New.
type Client struct{}

// New returns a payment client that routes by provider and channel.
func New() *Client {
	return &Client{}
}

// Create dispatches prepay to the registered provider.
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
	impl, err := lookup(provider)
	if err != nil {
		return nil, err
	}
	return impl.Create(ctx, cfg, channel, req)
}

// Query dispatches trade query to the registered provider.
func (c *Client) Query(ctx context.Context, provider string, cfg json.RawMessage, req QueryRequest) (*QueryResult, error) {
	req.OutTradeNo = strings.TrimSpace(req.OutTradeNo)
	req.TradeNo = strings.TrimSpace(req.TradeNo)
	if req.OutTradeNo == "" && req.TradeNo == "" {
		return nil, fmt.Errorf("%w: out_trade_no or trade_no is required", ErrInvalidRequest)
	}
	impl, err := lookup(provider)
	if err != nil {
		return nil, err
	}
	return impl.Query(ctx, cfg, req)
}

// Notify dispatches callback verification to the registered provider.
func (c *Client) Notify(ctx context.Context, provider string, cfg json.RawMessage, req NotifyRequest) (*NotifyResult, error) {
	if len(req.Body) == 0 {
		return nil, fmt.Errorf("%w: empty notify body", ErrInvalidRequest)
	}
	impl, err := lookup(provider)
	if err != nil {
		return nil, err
	}
	return impl.Notify(ctx, cfg, req)
}

// ValidateConfig runs the provider JSON validator.
func (c *Client) ValidateConfig(provider string, cfg json.RawMessage) error {
	impl, err := lookup(provider)
	if err != nil {
		return err
	}
	return impl.ValidateConfig(cfg)
}

// Registered reports whether a provider implementation is registered.
func (c *Client) Registered(provider string) bool {
	_, err := lookup(provider)
	return err == nil
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
