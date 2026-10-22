package service

import (
	"testing"
	"time"

	"github.com/ilaziness/orange-tv/internal/constant"
	"github.com/ilaziness/orange-tv/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecideSRSHook(t *testing.T) {
	t.Parallel()
	cfg := readySelfAuth()
	now := time.Unix(1_700_000_000, 0)
	exp := now.Add(time.Hour).Unix()
	token := signSelfToken(cfg.SignSecret, liveTokenKindPush, "stream1", "abcd1234abcd1234", exp)
	room := &model.LiveRooms{StreamKey: "stream1", Status: constant.LiveRoomLiving}

	require.NoError(t, DecideSRSHook(cfg, cfg.HookSecret, "on_publish", "live", "stream1", "?token="+token, room, now))

	idle := &model.LiveRooms{StreamKey: "stream1", Status: constant.LiveRoomIdle}
	assert.ErrorIs(t, DecideSRSHook(cfg, cfg.HookSecret, "on_publish", "live", "stream1", "?token="+token, idle, now), ErrLiveTokenInvalid)
	assert.ErrorIs(t, DecideSRSHook(cfg, "nope", "on_publish", "live", "stream1", "?token="+token, room, now), ErrLiveTokenInvalid)

	expired := signSelfToken(cfg.SignSecret, liveTokenKindPush, "stream1", "abcd1234abcd1234", now.Add(-time.Second).Unix())
	assert.ErrorIs(t, DecideSRSHook(cfg, cfg.HookSecret, "on_publish", "live", "stream1", "token="+expired, room, now), ErrLiveTokenExpired)

	assert.ErrorIs(t, DecideSRSHook(cfg, cfg.HookSecret, "on_publish", "other", "stream1", "?token="+token, room, now), ErrLiveTokenInvalid)

	playTok := signSelfToken(cfg.SignSecret, liveTokenKindPlay, "stream1", "abcd1234abcd1234", exp)
	assert.ErrorIs(t, DecideSRSHook(cfg, cfg.HookSecret, "on_publish", "live", "stream1", "?token="+playTok, room, now), ErrLiveTokenInvalid)

	require.NoError(t, VerifyLiveHookSecret(cfg, cfg.HookSecret))
	assert.ErrorIs(t, VerifyLiveHookSecret(cfg, "nope"), ErrLiveTokenInvalid)
	assert.ErrorIs(t, VerifyLiveHookSecret(cfg, ""), ErrLiveTokenInvalid)

	require.NoError(t, DecideSRSHook(cfg, cfg.HookSecret, "on_unpublish", "live", "stream1", "", room, now))
}

func TestDecidePlayAuth(t *testing.T) {
	t.Parallel()
	cfg := readySelfAuth()
	now := time.Unix(1_700_000_000, 0)
	token := signSelfToken(cfg.SignSecret, liveTokenKindPlay, "stream1", "abcd1234abcd1234", now.Add(time.Hour).Unix())
	room := &model.LiveRooms{StreamKey: "stream1", Status: constant.LiveRoomLiving}
	lookup := func(stream string) *model.LiveRooms {
		if stream == room.StreamKey {
			return room
		}
		return nil
	}
	uri := "/live/stream1.m3u8?token=" + token
	require.NoError(t, DecidePlayAuth(cfg, cfg.HookSecret, uri, lookup, now))
	assert.ErrorIs(t, DecidePlayAuth(cfg, "bad", uri, lookup, now), ErrLiveTokenInvalid)
	assert.ErrorIs(t, DecidePlayAuth(cfg, cfg.HookSecret, "/live/stream1.ts?token="+token, lookup, now), ErrLiveTokenInvalid)

	expired := signSelfToken(cfg.SignSecret, liveTokenKindPlay, "stream1", "abcd1234abcd1234", now.Add(-time.Minute).Unix())
	assert.ErrorIs(t, DecidePlayAuth(cfg, cfg.HookSecret, "/live/stream1.flv?token="+expired, lookup, now), ErrLiveTokenExpired)

	room.Status = constant.LiveRoomIdle
	assert.ErrorIs(t, DecidePlayAuth(cfg, cfg.HookSecret, uri, lookup, now), ErrLiveTokenInvalid)
}
