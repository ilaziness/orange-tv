package payment

import (
	"context"
	"encoding/json"
	"fmt"
)

// Payment is implemented by a payment vendor.
// Unused methods should keep the embedded Nop default.
type Payment interface {
	Create(ctx context.Context, cfg json.RawMessage, channel Channel, req CreateRequest) (*CreateResult, error)
	Query(ctx context.Context, cfg json.RawMessage, req QueryRequest) (*QueryResult, error)
	Notify(ctx context.Context, cfg json.RawMessage, req NotifyRequest) (*NotifyResult, error)
	ValidateConfig(cfg json.RawMessage) error
}

// Nop is the default empty Payment. Embed it and override only needed methods.
type Nop struct{}

func (Nop) Create(context.Context, json.RawMessage, Channel, CreateRequest) (*CreateResult, error) {
	return nil, fmt.Errorf("%w: Create", ErrNotImplemented)
}

func (Nop) Query(context.Context, json.RawMessage, QueryRequest) (*QueryResult, error) {
	return nil, fmt.Errorf("%w: Query", ErrNotImplemented)
}

func (Nop) Notify(context.Context, json.RawMessage, NotifyRequest) (*NotifyResult, error) {
	return nil, fmt.Errorf("%w: Notify", ErrNotImplemented)
}

func (Nop) ValidateConfig(json.RawMessage) error {
	return fmt.Errorf("%w: ValidateConfig", ErrNotImplemented)
}

var _ Payment = Nop{}
