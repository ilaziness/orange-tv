package service

import (
	"testing"

	"github.com/ilaziness/orange-tv/internal/constant"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestMapToPaymentSettingsMasksSecrets(t *testing.T) {
	t.Parallel()
	m := map[string]model.SystemSettings{
		constant.SettingPaymentAlipay: {
			SettingKey: constant.SettingPaymentAlipay,
			SettingValue: marshalPaymentJSON(paymentAlipayRaw{
				Enabled:         true,
				AppID:           "2021001",
				PrivateKey:      "secret-key",
				AlipayPublicKey: "pub",
				NotifyURL:       "https://example.com/alipay",
			}),
		},
		constant.SettingPaymentWechat: {
			SettingKey: constant.SettingPaymentWechat,
			SettingValue: marshalPaymentJSON(paymentWechatRaw{
				Enabled:    true,
				MchID:      "123456",
				APIv3Key:   "v3key",
				PrivateKey: "wx-pem",
			}),
		},
	}
	out := MapToPaymentSettings(m)
	assert.Equal(t, "2021001", out.Alipay.AppID)
	assert.True(t, out.Alipay.PrivateKeyConfigured)
	assert.True(t, out.Alipay.AlipayPublicKeyConfigured)
	assert.Empty(t, out.Alipay.PrivateKey)
	assert.Empty(t, out.Alipay.AlipayPublicKey)
	assert.Equal(t, "123456", out.Wechat.MchID)
	assert.True(t, out.Wechat.APIv3KeyConfigured)
	assert.True(t, out.Wechat.PrivateKeyConfigured)
	assert.Empty(t, out.Wechat.APIv3Key)
	assert.Empty(t, out.Wechat.PrivateKey)
}
