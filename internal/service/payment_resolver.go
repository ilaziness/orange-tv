package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/ilaziness/orange-tv/internal/constant"
	"github.com/ilaziness/orange-tv/internal/dto"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/ilaziness/orange-tv/pkg/payment"
	"go.uber.org/zap"
)

type paymentAlipayRaw struct {
	Enabled          bool   `json:"enabled"`
	Sandbox          bool   `json:"sandbox"`
	AppID            string `json:"app_id"`
	SignMode         string `json:"sign_mode"`
	PrivateKey       string `json:"private_key"`
	AlipayPublicKey  string `json:"alipay_public_key"`
	AppCert          string `json:"app_cert"`
	AlipayPublicCert string `json:"alipay_public_cert"`
	AlipayRootCert   string `json:"alipay_root_cert"`
	NotifyURL        string `json:"notify_url"`
	ReturnURL        string `json:"return_url"`
	PCWebEnabled     bool   `json:"pc_web_enabled"`
	AppEnabled       bool   `json:"app_enabled"`
}

type paymentWechatRaw struct {
	Enabled      bool   `json:"enabled"`
	MchID        string `json:"mch_id"`
	MchSerialNo  string `json:"mch_serial_no"`
	APIv3Key     string `json:"api_v3_key"`
	PrivateKey   string `json:"private_key"`
	AppIDWeb     string `json:"app_id_web"`
	AppIDApp     string `json:"app_id_app"`
	NotifyURL    string `json:"notify_url"`
	PCWebEnabled bool   `json:"pc_web_enabled"`
	AppEnabled   bool   `json:"app_enabled"`
}

func parsePaymentAlipayRaw(raw string) paymentAlipayRaw {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "{}" || raw == "null" {
		return paymentAlipayRaw{}
	}
	var out paymentAlipayRaw
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return paymentAlipayRaw{}
	}
	return out
}

func parsePaymentWechatRaw(raw string) paymentWechatRaw {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "{}" || raw == "null" {
		return paymentWechatRaw{}
	}
	var out paymentWechatRaw
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return paymentWechatRaw{}
	}
	return out
}

func marshalPaymentJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func maskAlipay(raw paymentAlipayRaw) dto.PaymentAlipayConfig {
	mode := strings.TrimSpace(raw.SignMode)
	if mode == "" {
		mode = "key"
	}
	return dto.PaymentAlipayConfig{
		Enabled:                    raw.Enabled,
		Sandbox:                    raw.Sandbox,
		AppID:                      raw.AppID,
		SignMode:                   mode,
		PrivateKeyConfigured:       strings.TrimSpace(raw.PrivateKey) != "",
		AlipayPublicKeyConfigured:  strings.TrimSpace(raw.AlipayPublicKey) != "",
		AppCertConfigured:          strings.TrimSpace(raw.AppCert) != "",
		AlipayPublicCertConfigured: strings.TrimSpace(raw.AlipayPublicCert) != "",
		AlipayRootCertConfigured:   strings.TrimSpace(raw.AlipayRootCert) != "",
		NotifyURL:                  raw.NotifyURL,
		ReturnURL:                  raw.ReturnURL,
		PCWebEnabled:               raw.PCWebEnabled,
		AppEnabled:                 raw.AppEnabled,
	}
}

func maskWechat(raw paymentWechatRaw) dto.PaymentWechatConfig {
	return dto.PaymentWechatConfig{
		Enabled:              raw.Enabled,
		MchID:                raw.MchID,
		MchSerialNo:          raw.MchSerialNo,
		APIv3KeyConfigured:   strings.TrimSpace(raw.APIv3Key) != "",
		PrivateKeyConfigured: strings.TrimSpace(raw.PrivateKey) != "",
		AppIDWeb:             raw.AppIDWeb,
		AppIDApp:             raw.AppIDApp,
		NotifyURL:            raw.NotifyURL,
		PCWebEnabled:         raw.PCWebEnabled,
		AppEnabled:           raw.AppEnabled,
	}
}

// MapToPaymentSettings maps settings rows to admin PaymentSettings (secrets masked).
func MapToPaymentSettings(m map[string]model.SystemSettings) dto.PaymentSettings {
	return dto.PaymentSettings{
		Alipay: maskAlipay(parsePaymentAlipayRaw(StrVal(m, constant.SettingPaymentAlipay))),
		Wechat: maskWechat(parsePaymentWechatRaw(StrVal(m, constant.SettingPaymentWechat))),
	}
}

func paymentSettingKey(provider string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case payment.ProviderAlipay:
		return constant.SettingPaymentAlipay, nil
	case payment.ProviderWechat:
		return constant.SettingPaymentWechat, nil
	default:
		return "", fmt.Errorf("未知支付商")
	}
}

// PaymentResolver loads merchant JSON from settings and dispatches via pkg/payment.
type PaymentResolver interface {
	LoadJSON(ctx context.Context, provider string) (json.RawMessage, error)
	Validate(provider string, raw json.RawMessage) error
	Create(ctx context.Context, provider string, channel payment.Channel, req payment.CreateRequest) (*payment.CreateResult, error)
	Query(ctx context.Context, provider string, req payment.QueryRequest) (*payment.QueryResult, error)
	Notify(ctx context.Context, provider string, req payment.NotifyRequest) (*payment.NotifyResult, error)
}

type paymentResolver struct {
	settings SettingsService
	pay      *payment.Client
	log      *zap.Logger
}

// NewPaymentResolver creates a PaymentResolver. pay is the process-wide payment.New() client.
func NewPaymentResolver(settings SettingsService, pay *payment.Client, log *zap.Logger) PaymentResolver {
	return &paymentResolver{settings: settings, pay: pay, log: log}
}

func (r *paymentResolver) loadGroup(ctx context.Context) (map[string]model.SystemSettings, error) {
	return r.settings.LoadMapByGroup(ctx, constant.SettingGroupPayment)
}

func (r *paymentResolver) LoadJSON(ctx context.Context, provider string) (json.RawMessage, error) {
	key, err := paymentSettingKey(provider)
	if err != nil {
		return nil, errcode.WithMessage(errcode.PaymentConfigInvalid, err.Error())
	}
	m, err := r.loadGroup(ctx)
	if err != nil {
		return nil, err
	}
	raw := strings.TrimSpace(StrVal(m, key))
	if raw == "" {
		raw = "{}"
	}
	return json.RawMessage(raw), nil
}

func (r *paymentResolver) Validate(provider string, raw json.RawMessage) error {
	if err := r.pay.ValidateConfig(provider, raw); err != nil {
		r.log.Warn("payment config invalid", zap.String("provider", provider), zap.Error(err))
		return mapPaymentError(err, errcode.PaymentConfigInvalid)
	}
	return nil
}

func (r *paymentResolver) Create(ctx context.Context, provider string, channel payment.Channel, req payment.CreateRequest) (*payment.CreateResult, error) {
	raw, err := r.LoadJSON(ctx, provider)
	if err != nil {
		return nil, err
	}
	out, err := r.pay.Create(ctx, provider, channel, raw, req)
	if err != nil {
		return nil, mapPaymentError(err, errcode.PaymentCreateFailed)
	}
	return out, nil
}

func (r *paymentResolver) Query(ctx context.Context, provider string, req payment.QueryRequest) (*payment.QueryResult, error) {
	raw, err := r.LoadJSON(ctx, provider)
	if err != nil {
		return nil, err
	}
	out, err := r.pay.Query(ctx, provider, raw, req)
	if err != nil {
		return nil, mapPaymentError(err, errcode.PaymentQueryFailed)
	}
	return out, nil
}

func (r *paymentResolver) Notify(ctx context.Context, provider string, req payment.NotifyRequest) (*payment.NotifyResult, error) {
	raw, err := r.LoadJSON(ctx, provider)
	if err != nil {
		return nil, err
	}
	out, err := r.pay.Notify(ctx, provider, raw, req)
	if err != nil {
		return nil, mapPaymentError(err, errcode.PaymentNotifyInvalid)
	}
	return out, nil
}

func mapPaymentError(err error, fallback *errcode.Code) error {
	switch {
	case errors.Is(err, payment.ErrInvalidConfig):
		return errcode.WithMessage(errcode.PaymentConfigInvalid, err.Error())
	case errors.Is(err, payment.ErrInvalidRequest):
		return errcode.WithMessage(errcode.ParamError, err.Error())
	case errors.Is(err, payment.ErrProviderDisabled), errors.Is(err, payment.ErrProviderNotRegistered):
		return errcode.Wrap(errcode.PaymentProviderDisabled, err)
	case errors.Is(err, payment.ErrChannelDisabled), errors.Is(err, payment.ErrChannelNotRegistered):
		return errcode.Wrap(errcode.PaymentChannelDisabled, err)
	default:
		return errcode.Wrap(fallback, err)
	}
}
