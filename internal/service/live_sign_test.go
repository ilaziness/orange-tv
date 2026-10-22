package service

import (
	"net/url"
	"testing"
	"time"

	"github.com/ilaziness/orange-tv/internal/constant"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSelfTokenOfficialVector(t *testing.T) {
	t.Parallel()
	const (
		secret = "sign-secret"
		stream = "stream1"
		rand   = "abcd1234abcd1234"
		exp    = int64(1700000000)
	)
	play := signSelfToken(secret, liveTokenKindPlay, stream, rand, exp)
	assert.Equal(t, "1700000000-abcd1234abcd1234-2a2b8d568213e6466056ee2d4c2b00d29f8f1741079a2ce26789cc28e9e31742", play)
	push := signSelfToken(secret, liveTokenKindPush, stream, rand, exp)
	assert.Equal(t, "1700000000-abcd1234abcd1234-f8ba892629ad8c8214d8b265a33ab272ea92db3f3d571e2e6ce2186a4059ef90", push)
	assert.NotEqual(t, play, push)

	now := time.Unix(exp, 0)
	require.NoError(t, verifySelfToken(secret, liveTokenKindPlay, stream, play, now))
	assert.ErrorIs(t, verifySelfToken(secret, liveTokenKindPlay, "other", play, now), ErrLiveTokenInvalid)
	assert.ErrorIs(t, verifySelfToken(secret, liveTokenKindPush, stream, play, now), ErrLiveTokenInvalid)
	assert.ErrorIs(t, verifySelfToken("wrong", liveTokenKindPlay, stream, play, now), ErrLiveTokenInvalid)
	assert.ErrorIs(t, verifySelfToken(secret, liveTokenKindPlay, stream, play, time.Unix(exp+1, 0)), ErrLiveTokenExpired)
}

func TestSelfIssueKeepsPreviousToken(t *testing.T) {
	t.Parallel()
	cfg := readySelfAuth()
	now := time.Unix(1_700_000_000, 0)
	first, err := issueSelf(cfg, "stream1", now)
	require.NoError(t, err)
	second, err := issueSelf(cfg, "stream1", now.Add(time.Minute))
	require.NoError(t, err)
	assert.NotEqual(t, first.StreamKey, second.StreamKey)
	token := tokenFromQuery(first.StreamKey[len("stream1"):], "token")
	require.NoError(t, verifySelfToken(cfg.SignSecret, liveTokenKindPush, "stream1", token, now.Add(time.Minute)))
}

func TestSelfPlayAppendsTokenWithAmpersand(t *testing.T) {
	t.Parallel()
	cfg := readySelfAuth()
	cfg.PlayHLSTemplate = "https://media.example.com/live/{stream_key}.m3u8?cdn=1"
	urls, err := issueSelf(cfg, "stream1", time.Unix(1_700_000_000, 0))
	require.NoError(t, err)
	assert.Contains(t, urls.PlayHLS, ".m3u8?cdn=1&token=")
	assert.Contains(t, urls.StreamKey, "stream1?token=")
	assert.Equal(t, "rtmp://media.example.com/live", urls.PushServer)
}

func TestSelfAuthOffBareURLs(t *testing.T) {
	t.Parallel()
	cfg := LiveStreamConfig{
		Provider:        constant.LiveProviderSelf,
		PushServer:      "rtmp://media.example.com/live",
		PlayHLSTemplate: "https://media.example.com/live/{stream_key}.m3u8",
	}
	urls, err := IssueLiveMedia(cfg, "abc", time.Unix(1_700_000_000, 0))
	require.NoError(t, err)
	assert.Equal(t, "abc", urls.StreamKey)
	assert.Equal(t, "https://media.example.com/live/abc.m3u8", urls.PlayHLS)
	assert.True(t, urls.PlayExpiresAt.IsZero())
}

func TestAliyunTypeAOfficialVector(t *testing.T) {
	t.Parallel()
	// Aliyun URL auth type A sample string:
	// /video/standard/1K.html-1444435200-0-0-aliyuncdnexp1234
	key := aliyunAuthKey("/video/standard/1K.html", "aliyuncdnexp1234", "0", "0", 1444435200)
	assert.Equal(t, "1444435200-0-0-80cd3862d699b7118eed99103f2a3a4f", key)

	now := time.Unix(1444435200, 0)
	require.NoError(t, verifyAliyunAuthKey("/video/standard/1K.html", "aliyuncdnexp1234", key, now))
	assert.ErrorIs(t, verifyAliyunAuthKey("/video/other.mp4", "aliyuncdnexp1234", key, now), ErrLiveTokenInvalid)
	bad := key[:len(key)-1] + "0"
	assert.ErrorIs(t, verifyAliyunAuthKey("/video/standard/1K.html", "aliyuncdnexp1234", bad, now), ErrLiveTokenInvalid)
	assert.ErrorIs(t, verifyAliyunAuthKey("/video/standard/1K.html", "aliyuncdnexp1234", key, now.Add(time.Second)), ErrLiveTokenExpired)
}

func TestAliyunIssuePaths(t *testing.T) {
	t.Parallel()
	cfg := LiveStreamConfig{
		Provider:       constant.LiveProviderAliyun,
		PushTTLSeconds: 3600,
		PlayTTLSeconds: 3600,
		Aliyun: LiveAliyunConfig{
			PushDomain: "push.example.com", PlayDomain: "play.example.com", AppName: "live",
			Region: constant.LiveAliyunRegionDefault, PushAuthKey: "push-key", PlayAuthKey: "play-key",
			AccessKeyID: "ak", AccessKeySecret: "sk",
		},
	}
	urls, err := IssueLiveMedia(cfg, "room1", time.Unix(1_700_000_000, 0))
	require.NoError(t, err)
	assert.Equal(t, "rtmp://push.example.com/live", urls.PushServer)
	assert.Contains(t, urls.StreamKey, "room1?auth_key=")
	assert.Contains(t, urls.PlayHLS, "https://play.example.com/live/room1.m3u8?auth_key=")
	assert.Contains(t, urls.PlayFLV, "https://play.example.com/live/room1.flv?auth_key=")
	hls, err := url.Parse(urls.PlayHLS)
	require.NoError(t, err)
	require.NoError(t, verifyAliyunAuthKey("/live/room1.m3u8", "play-key", hls.Query().Get("auth_key"), time.Unix(1_700_000_000, 0)))
}

func TestQiniuTimestampOfficialVector(t *testing.T) {
	t.Parallel()
	// qiniu/go-sdk pili TestSignPlayURL tsStartMD5 vector:
	// key+path+hex(t) = test123/hub/streamTitle61703d00
	got, err := qiniuTimestampURL("https://play.domain.com/hub/streamTitle", "test123", 1634745600)
	require.NoError(t, err)
	assert.Equal(t, "https://play.domain.com/hub/streamTitle?sign=b1a151275da51ce22f50dd97df3cdc86&t=61703d00", got)

	now := time.Unix(1634745600, 0)
	require.NoError(t, verifyQiniuTimestamp(got, "test123", now))
	assert.ErrorIs(t, verifyQiniuTimestamp("https://play.domain.com/hub/other?sign=b1a151275da51ce22f50dd97df3cdc86&t=61703d00", "test123", now), ErrLiveTokenInvalid)
	assert.ErrorIs(t, verifyQiniuTimestamp(got, "test123", now.Add(time.Second)), ErrLiveTokenExpired)
	assert.ErrorIs(t, verifyQiniuTimestamp("https://play.domain.com/hub/streamTitle?sign=deadbeef&t=61703d00", "test123", now), ErrLiveTokenInvalid)

	withQuery, err := qiniuTimestampURL("https://play.domain.com/hub/streamTitle.m3u8?cdn=1", "test123", 1634745600)
	require.NoError(t, err)
	assert.Contains(t, withQuery, "?cdn=1&sign=")
}

func TestQiniuPushExpirySKVector(t *testing.T) {
	t.Parallel()
	const exp = int64(1584522520)
	cfg := LiveStreamConfig{
		Provider:       constant.LiveProviderQiniu,
		PushTTLSeconds: 3600,
		PlayTTLSeconds: 3600,
		Qiniu: LiveQiniuConfig{
			Hub:        "testhub",
			PushDomain: "publish.domain.com",
			PlayDomain: "play.domain.com",
			AccessKey:  "7O7hf7Ld1RrC_fpZdFvU8aCgOPuhw2K4eapYOdII",
			SecretKey:  "312ae9gd2BrCfpTdF4U8aIg9Puh62K4eEGY72Ea_",
		},
	}
	urls, err := IssueLiveMedia(cfg, "teststreamtitle", time.Unix(exp-3600, 0))
	require.NoError(t, err)
	assert.Equal(t, "rtmp://publish.domain.com/testhub", urls.PushServer)
	assert.Equal(t, "teststreamtitle?e=1584522520&token=7O7hf7Ld1RrC_fpZdFvU8aCgOPuhw2K4eapYOdII:NfI2OWGCMdFDTLOfeUd-zSPVrFY=", urls.StreamKey)
	assert.Contains(t, urls.PlayHLS, "https://play.domain.com/testhub/teststreamtitle.m3u8?sign=")
	assert.Contains(t, urls.PlayFLV, "https://play.domain.com/testhub/teststreamtitle.flv?sign=")
}

func TestMapToLiveSettingsHidesSignSecret(t *testing.T) {
	t.Parallel()
	m := map[string]model.SystemSettings{
		constant.SettingLiveProvider:    {SettingValue: constant.LiveProviderSelf},
		constant.SettingLiveAuthEnabled: {SettingValue: "1"},
		constant.SettingLiveSignSecret:  {SettingValue: "sign-secret-value"},
		constant.SettingLiveHookSecret:  {SettingValue: "hook-secret-value"},
		constant.SettingLiveSRSAPIBase:  {SettingValue: "http://127.0.0.1:1985"},
		constant.SettingLivePushServer:  {SettingValue: "rtmp://media.example.com/live"},
	}
	view := MapToLiveSettings(m, "https://tv.example.com")
	assert.True(t, view.SignSecretConfigured)
	assert.Empty(t, view.SignSecret)
	assert.Empty(t, view.HookSecret)
	assert.Contains(t, view.SRSHookURL, "hook_secret=hook-secret-value")
	assert.Contains(t, view.SRSHookURL, "https://tv.example.com/api/open/v1/live/hooks/srs")
	assert.NotContains(t, view.SRSHookURL, "sign-secret-value")
	assert.Contains(t, view.PlayAuthURL, "/api/open/v1/live/auth/play")
	assert.Equal(t, "https://tv.example.com/api/open/v1/live/hooks/srs?hook_secret=hook-secret-value", view.SRSHookURL)

	pathOnly := MapToLiveSettings(m, "")
	assert.Equal(t, "/api/open/v1/live/hooks/srs?hook_secret=hook-secret-value", pathOnly.SRSHookURL)
}

func readySelfAuth() LiveStreamConfig {
	return LiveStreamConfig{
		Provider:        constant.LiveProviderSelf,
		AuthEnabled:     true,
		PushTTLSeconds:  3600,
		PlayTTLSeconds:  3600,
		SignSecret:      "sign-secret",
		HookSecret:      "hook-secret",
		SRSAPIBase:      "http://127.0.0.1:1985",
		PushServer:      "rtmp://media.example.com/live",
		PlayHLSTemplate: "https://media.example.com/live/{stream_key}.m3u8",
		PlayFLVTemplate: "https://media.example.com/live/{stream_key}.flv",
	}
}
