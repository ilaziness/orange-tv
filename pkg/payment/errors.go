package payment

import "errors"

var (
	ErrProviderNotRegistered = errors.New("payment: provider not registered")
	ErrChannelNotRegistered  = errors.New("payment: channel not registered")
	ErrProviderDisabled      = errors.New("payment: provider disabled")
	ErrChannelDisabled       = errors.New("payment: channel disabled")
	ErrInvalidConfig         = errors.New("payment: invalid config")
	ErrInvalidRequest        = errors.New("payment: invalid request")
	ErrNotImplemented        = errors.New("payment: not implemented")
)
