package service

import (
	"testing"

	"github.com/ilaziness/orange-tv/internal/constant"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateLiveStreamKey(t *testing.T) {
	t.Parallel()
	key, err := GenerateLiveStreamKey()
	require.NoError(t, err)
	assert.Len(t, key, 32)
}

func TestApplyStreamKeyTemplate(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "https://x/live/abc.m3u8", ApplyStreamKeyTemplate("https://x/live/{stream_key}.m3u8", "abc"))
	assert.Equal(t, "", ApplyStreamKeyTemplate("", "abc"))
	assert.Equal(t, "", ApplyStreamKeyTemplate("https://x/{stream_key}", ""))
}

func TestLiveStreamConfigReadyForBroadcast(t *testing.T) {
	t.Parallel()
	assert.False(t, LiveStreamConfig{}.ReadyForBroadcast())
	assert.False(t, LiveStreamConfig{PushServer: "rtmp://x/live"}.ReadyForBroadcast())
	assert.True(t, LiveStreamConfig{
		PushServer:      "rtmp://x/live",
		PlayHLSTemplate: "https://x/{stream_key}.m3u8",
	}.ReadyForBroadcast())
	assert.True(t, LiveStreamConfig{
		PushServer:      "rtmp://x/live",
		PlayFLVTemplate: "https://x/{stream_key}.flv",
	}.ReadyForBroadcast())
}

func TestParseLiveStreamConfig(t *testing.T) {
	t.Parallel()
	m := map[string]model.SystemSettings{
		constant.SettingLivePushServer:      {SettingValue: " rtmp://a/live "},
		constant.SettingLivePlayHLSTemplate: {SettingValue: "https://a/{stream_key}.m3u8"},
	}
	cfg := ParseLiveStreamConfig(m)
	assert.Equal(t, "rtmp://a/live", cfg.PushServer)
	hls, flv := cfg.ResolvePlayURLs("k1")
	assert.Equal(t, "https://a/k1.m3u8", hls)
	assert.Equal(t, "", flv)
}

func TestResolvePublicPlayURLs_OnlyWhenLiving(t *testing.T) {
	t.Parallel()
	cfg := LiveStreamConfig{
		PlayHLSTemplate: "https://x/{stream_key}.m3u8",
		PlayFLVTemplate: "https://x/{stream_key}.flv",
	}
	hls, flv := cfg.ResolvePublicPlayURLs("abc", constant.LiveRoomLiving)
	assert.Equal(t, "https://x/abc.m3u8", hls)
	assert.Equal(t, "https://x/abc.flv", flv)

	hls, flv = cfg.ResolvePublicPlayURLs("abc", constant.LiveRoomIdle)
	assert.Equal(t, "", hls)
	assert.Equal(t, "", flv)

	hls, flv = cfg.ResolvePublicPlayURLs("abc", constant.LiveRoomBanned)
	assert.Equal(t, "", hls)
	assert.Equal(t, "", flv)
}

func TestResolveStreamerApplyStatus(t *testing.T) {
	t.Parallel()
	assert.Equal(t, constant.StreamerApplyStatusApproved, ResolveStreamerApplyStatus(1, nil))
	assert.Equal(t, constant.StreamerApplyStatusNone, ResolveStreamerApplyStatus(0, nil))

	pending := &model.StreamerApplies{Status: constant.StreamerApplyPending}
	assert.Equal(t, constant.StreamerApplyStatusPending, ResolveStreamerApplyStatus(0, pending))

	rejected := &model.StreamerApplies{Status: constant.StreamerApplyRejected}
	assert.Equal(t, constant.StreamerApplyStatusRejected, ResolveStreamerApplyStatus(0, rejected))

	// Revoked streamer: latest apply is still approved, but is_streamer=0 → allow re-apply.
	approved := &model.StreamerApplies{Status: constant.StreamerApplyApproved}
	assert.Equal(t, constant.StreamerApplyStatusNone, ResolveStreamerApplyStatus(0, approved))
}
