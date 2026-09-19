package wechat

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/ilaziness/orange-tv/pkg/payment"
)

// Config is the JSON stored in system_settings.payment_wechat.
type Config struct {
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

func parseConfig(raw json.RawMessage) (Config, error) {
	raw = json.RawMessage(strings.TrimSpace(string(raw)))
	if len(raw) == 0 || string(raw) == "null" || string(raw) == "{}" {
		return Config{}, nil
	}
	var cfg Config
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return Config{}, fmt.Errorf("%w: %v", payment.ErrInvalidConfig, err)
	}
	cfg.MchID = strings.TrimSpace(cfg.MchID)
	cfg.MchSerialNo = strings.TrimSpace(cfg.MchSerialNo)
	cfg.AppIDWeb = strings.TrimSpace(cfg.AppIDWeb)
	cfg.AppIDApp = strings.TrimSpace(cfg.AppIDApp)
	cfg.NotifyURL = strings.TrimSpace(cfg.NotifyURL)
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
	if err := optionalHTTPSURL(cfg.NotifyURL); err != nil {
		return fmt.Errorf("%w: notify_url: %v", payment.ErrInvalidConfig, err)
	}
	if !cfg.Enabled {
		return nil
	}
	if cfg.NotifyURL == "" {
		return fmt.Errorf("%w: notify_url is required", payment.ErrInvalidConfig)
	}
	if cfg.MchID == "" {
		return fmt.Errorf("%w: mch_id is required", payment.ErrInvalidConfig)
	}
	if cfg.MchSerialNo == "" {
		return fmt.Errorf("%w: mch_serial_no is required", payment.ErrInvalidConfig)
	}
	if strings.TrimSpace(cfg.APIv3Key) == "" {
		return fmt.Errorf("%w: api_v3_key is required", payment.ErrInvalidConfig)
	}
	if strings.TrimSpace(cfg.PrivateKey) == "" {
		return fmt.Errorf("%w: private_key is required", payment.ErrInvalidConfig)
	}
	if cfg.PCWebEnabled && cfg.AppIDWeb == "" {
		return fmt.Errorf("%w: app_id_web is required when pc_web is enabled", payment.ErrInvalidConfig)
	}
	if cfg.AppEnabled && cfg.AppIDApp == "" {
		return fmt.Errorf("%w: app_id_app is required when app is enabled", payment.ErrInvalidConfig)
	}
	return nil
}

func optionalHTTPSURL(raw string) error {
	if raw == "" {
		return nil
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || u.Scheme != "https" {
		return fmt.Errorf("must be https URL")
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
