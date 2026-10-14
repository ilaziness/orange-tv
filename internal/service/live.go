package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/ilaziness/orange-tv/internal/constant"
	"github.com/ilaziness/orange-tv/internal/model"
)

const liveStreamKeyBytes = 16 // 32 hex chars

// LiveStreamConfig holds admin-only push/play configuration.
type LiveStreamConfig struct {
	PushServer      string
	PlayHLSTemplate string
	PlayFLVTemplate string
}

// GenerateLiveStreamKey returns a 32+ hex stream key.
func GenerateLiveStreamKey() (string, error) {
	b := make([]byte, liveStreamKeyBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate live stream key: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// ApplyStreamKeyTemplate replaces {stream_key} in a play URL template.
func ApplyStreamKeyTemplate(template, streamKey string) string {
	template = strings.TrimSpace(template)
	if template == "" || streamKey == "" {
		return ""
	}
	return strings.ReplaceAll(template, "{stream_key}", streamKey)
}

// ParseLiveStreamConfig maps live settings rows to LiveStreamConfig.
func ParseLiveStreamConfig(m map[string]model.SystemSettings) LiveStreamConfig {
	return LiveStreamConfig{
		PushServer:      strings.TrimSpace(StrVal(m, constant.SettingLivePushServer)),
		PlayHLSTemplate: strings.TrimSpace(StrVal(m, constant.SettingLivePlayHLSTemplate)),
		PlayFLVTemplate: strings.TrimSpace(StrVal(m, constant.SettingLivePlayFLVTemplate)),
	}
}

// ReadyForBroadcast reports whether push server and at least one play template are set.
func (c LiveStreamConfig) ReadyForBroadcast() bool {
	if c.PushServer == "" {
		return false
	}
	return c.PlayHLSTemplate != "" || c.PlayFLVTemplate != ""
}

// ResolvePlayURLs returns resolved HLS/FLV URLs for a stream key (empty if template blank).
func (c LiveStreamConfig) ResolvePlayURLs(streamKey string) (hls, flv string) {
	return ApplyStreamKeyTemplate(c.PlayHLSTemplate, streamKey), ApplyStreamKeyTemplate(c.PlayFLVTemplate, streamKey)
}

// ResolvePublicPlayURLs returns play URLs only while the room is living.
// Idle/banned rooms must not expose URLs that embed stream_key.
func (c LiveStreamConfig) ResolvePublicPlayURLs(streamKey string, roomStatus uint8) (hls, flv string) {
	if roomStatus != constant.LiveRoomLiving {
		return "", ""
	}
	return c.ResolvePlayURLs(streamKey)
}

// ResolveStreamerApplyStatus maps DB state to profile streamer_apply_status.
// An approved apply after revoke (is_streamer=0) maps to "none" so the user can re-apply.
func ResolveStreamerApplyStatus(isStreamer uint8, apply *model.StreamerApplies) string {
	if isStreamer == 1 {
		return constant.StreamerApplyStatusApproved
	}
	if apply == nil {
		return constant.StreamerApplyStatusNone
	}
	switch apply.Status {
	case constant.StreamerApplyPending:
		return constant.StreamerApplyStatusPending
	case constant.StreamerApplyRejected:
		return constant.StreamerApplyStatusRejected
	case constant.StreamerApplyApproved:
		return constant.StreamerApplyStatusNone
	default:
		return constant.StreamerApplyStatusNone
	}
}

// NewLiveSessionSnapshot builds a session row from room + user snapshot fields.
func NewLiveSessionSnapshot(room *model.LiveRooms, user *model.Users, startedAt time.Time) *model.LiveSessions {
	return &model.LiveSessions{
		RoomID:           room.ID,
		UserID:           room.UserID,
		Title:            room.Title,
		CategoryID:       room.CategoryID,
		StreamerNickname: user.Nickname,
		StreamerAvatar:   user.Avatar,
		StartedAt:        startedAt,
	}
}
