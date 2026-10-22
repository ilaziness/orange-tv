package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/ilaziness/orange-tv/internal/constant"
	errcode "github.com/ilaziness/orange-tv/internal/errcode"
)

// IssueLiveMedia signs push and play URLs for the active provider.
// Self with auth disabled keeps the bare push server, stream name, and templates.
// A config that is not ready returns empty URLs and a nil error.
func IssueLiveMedia(cfg LiveStreamConfig, streamName string, now time.Time) (LiveMediaURLs, error) {
	streamName = strings.TrimSpace(streamName)
	if streamName == "" {
		return LiveMediaURLs{}, errcode.WithMessage(errcode.ParamError, "流名为空")
	}
	if cfg.Validate() != nil {
		return LiveMediaURLs{}, nil
	}
	switch cfg.ProviderName() {
	case constant.LiveProviderSelf:
		if !cfg.AuthEnabled {
			hls, flv := cfg.ResolvePlayURLs(streamName)
			return LiveMediaURLs{
				PushServer: cfg.PushServer,
				StreamKey:  streamName,
				PlayHLS:    hls,
				PlayFLV:    flv,
			}, nil
		}
		return issueSelf(cfg, streamName, now)
	case constant.LiveProviderAliyun:
		return issueAliyun(cfg, streamName, now)
	case constant.LiveProviderQiniu:
		return issueQiniu(cfg, streamName, now)
	default:
		return LiveMediaURLs{}, errcode.WithMessage(errcode.SettingInvalid, "直播厂商无效")
	}
}

// LivePushServer is the OBS server field without a signed stream key.
// An incomplete config returns an empty server so the studio can show that broadcast is unavailable.
func LivePushServer(cfg LiveStreamConfig) string {
	if cfg.Validate() != nil {
		return ""
	}
	switch cfg.ProviderName() {
	case constant.LiveProviderSelf:
		return cfg.PushServer
	case constant.LiveProviderAliyun:
		return "rtmp://" + cfg.Aliyun.PushDomain + "/" + cfg.Aliyun.AppName
	case constant.LiveProviderQiniu:
		return "rtmp://" + cfg.Qiniu.PushDomain + "/" + cfg.Qiniu.Hub
	default:
		return ""
	}
}

func randomHex(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("live token rand: %w", err)
	}
	return hex.EncodeToString(b), nil
}
