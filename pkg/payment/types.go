package payment

import (
	"encoding/json"
	"net/http"
	"time"
)

// Channel identifies a payment channel independent of merchant.
type Channel string

const (
	ChannelPCWeb Channel = "pc_web"
	ChannelApp   Channel = "app"
)

// Merchant identifiers. New vendors add a constant here and a subpackage.
const (
	ProviderAlipay = "alipay"
	ProviderWechat = "wechat"
)

// CreateRequest is the vendor-agnostic prepay input.
type CreateRequest struct {
	OutTradeNo string
	AmountFen  int64
	Subject    string
	ExpireAt   time.Time
	ClientIP   string
	ReturnURL  string
}

// CreateResult is the vendor-agnostic prepay output. Fields are filled by channel.
type CreateResult struct {
	Channel    Channel
	PayURL     string
	CodeURL    string
	FormHTML   string
	AppPayload string
}

// QueryRequest looks up a trade by merchant or platform trade no.
type QueryRequest struct {
	OutTradeNo string
	TradeNo    string
}

// Trade status values returned by Query/Notify after vendor mapping.
const (
	TradeStatusPending = "pending"
	TradeStatusSuccess = "success"
	TradeStatusClosed  = "closed"
	TradeStatusRefund  = "refund"
)

// QueryResult is a unified trade snapshot.
type QueryResult struct {
	OutTradeNo string
	TradeNo    string
	Status     string
	AmountFen  int64
}

// NotifyRequest is the raw HTTP callback from a payment vendor.
type NotifyRequest struct {
	Headers http.Header
	Body    []byte
}

// NotifyResult is a verified callback payload.
type NotifyResult struct {
	OutTradeNo string
	TradeNo    string
	Status     string
	AmountFen  int64
	Raw        json.RawMessage
}
