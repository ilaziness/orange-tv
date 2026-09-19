package alipay

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/ilaziness/orange-tv/pkg/payment"
)

const (
	signModeKey  = "key"
	signModeCert = "cert"
)

// Config is the JSON stored in system_settings.payment_alipay.
type Config struct {
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

func parseConfig(raw json.RawMessage) (Config, error) {
	raw = json.RawMessage(strings.TrimSpace(string(raw)))
	if len(raw) == 0 || string(raw) == "null" || string(raw) == "{}" {
		return Config{}, nil
	}
	var cfg Config
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return Config{}, fmt.Errorf("%w: %v", payment.ErrInvalidConfig, err)
	}
	cfg.AppID = strings.TrimSpace(cfg.AppID)
	cfg.SignMode = strings.ToLower(strings.TrimSpace(cfg.SignMode))
	cfg.NotifyURL = strings.TrimSpace(cfg.NotifyURL)
	cfg.ReturnURL = strings.TrimSpace(cfg.ReturnURL)
	return cfg, nil
}

func validateConfig(raw json.RawMessage) error {
	cfg, err := parseConfig(raw)
	if err != nil {
		return err
	}
	return checkConfig(cfg)
}

func checkConfig(cfg Config) error {
	if err := optionalHTTPURL(cfg.NotifyURL); err != nil {
		return fmt.Errorf("%w: notify_url: %v", payment.ErrInvalidConfig, err)
	}
	if err := optionalHTTPURL(cfg.ReturnURL); err != nil {
		return fmt.Errorf("%w: return_url: %v", payment.ErrInvalidConfig, err)
	}
	if !cfg.Enabled {
		return nil
	}
	if cfg.NotifyURL == "" {
		return fmt.Errorf("%w: notify_url is required", payment.ErrInvalidConfig)
	}
	if cfg.AppID == "" {
		return fmt.Errorf("%w: app_id is required", payment.ErrInvalidConfig)
	}
	if strings.TrimSpace(cfg.PrivateKey) == "" {
		return fmt.Errorf("%w: private_key is required", payment.ErrInvalidConfig)
	}
	mode := cfg.SignMode
	if mode == "" {
		mode = signModeKey
	}
	switch mode {
	case signModeKey:
		if strings.TrimSpace(cfg.AlipayPublicKey) == "" {
			return fmt.Errorf("%w: alipay_public_key is required for key mode", payment.ErrInvalidConfig)
		}
	case signModeCert:
		if strings.TrimSpace(cfg.AppCert) == "" || strings.TrimSpace(cfg.AlipayPublicCert) == "" || strings.TrimSpace(cfg.AlipayRootCert) == "" {
			return fmt.Errorf("%w: cert mode requires app_cert, alipay_public_cert and alipay_root_cert", payment.ErrInvalidConfig)
		}
	default:
		return fmt.Errorf("%w: sign_mode must be key or cert", payment.ErrInvalidConfig)
	}
	return nil
}

func optionalHTTPURL(raw string) error {
	if raw == "" {
		return nil
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return fmt.Errorf("must be http(s) URL")
	}
	return nil
}

func ensureReady(cfg Config, channel payment.Channel) error {
	if !cfg.Enabled {
		return payment.ErrProviderDisabled
	}
	switch channel {
	case payment.ChannelPCWeb:
		if !cfg.PCWebEnabled {
			return payment.ErrChannelDisabled
		}
	case payment.ChannelApp:
		if !cfg.AppEnabled {
			return payment.ErrChannelDisabled
		}
	default:
		return fmt.Errorf("%w: %s", payment.ErrChannelNotRegistered, channel)
	}
	return checkConfig(cfg)
}
